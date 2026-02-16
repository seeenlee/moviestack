package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type movieImportJobState struct {
	mu            sync.Mutex
	running       bool
	status        string
	startedAt     time.Time
	finishedAt    time.Time
	sourceFile    string
	processedRows int64
	upsertedRows  int64
	lastErr       string
}

var errNoMovieIDFiles = errors.New("no .json.gz files found")

const importCopyBatchSize = 10000

const createMovieIDImportStagingSQL = `
CREATE TEMP TABLE movie_ids_import_staging (
	id             INTEGER        NOT NULL,
	original_title TEXT           NOT NULL,
	adult          BOOLEAN        NOT NULL,
	video          BOOLEAN        NOT NULL,
	popularity     NUMERIC(10, 4) NOT NULL
) ON COMMIT DROP
`

const mergeMovieIDImportStagingSQL = `
INSERT INTO movie_ids (id, original_title, adult, video, popularity)
SELECT id, original_title, adult, video, popularity
FROM movie_ids_import_staging
ON CONFLICT (id) DO UPDATE
SET
	original_title = EXCLUDED.original_title,
	adult = EXCLUDED.adult,
	video = EXCLUDED.video,
	popularity = EXCLUDED.popularity
`

func (s *movieImportJobState) startIfIdle(sourceFile string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	s.status = "running"
	s.startedAt = time.Now().UTC()
	s.finishedAt = time.Time{}
	s.sourceFile = sourceFile
	s.processedRows = 0
	s.upsertedRows = 0
	s.lastErr = ""
	return true
}

func (s *movieImportJobState) updateProgress(processedRows, upsertedRows int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processedRows = processedRows
	s.upsertedRows = upsertedRows
}

func (s *movieImportJobState) finishSuccess(processedRows, upsertedRows int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.status = "succeeded"
	s.finishedAt = time.Now().UTC()
	s.processedRows = processedRows
	s.upsertedRows = upsertedRows
	s.lastErr = ""
}

func (s *movieImportJobState) finishFailure(processedRows, upsertedRows int64, lastErr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.status = "failed"
	s.finishedAt = time.Now().UTC()
	s.processedRows = processedRows
	s.upsertedRows = upsertedRows
	s.lastErr = lastErr
}

func (s *movieImportJobState) snapshot() ImportStatusResponse {
	s.mu.Lock()
	defer s.mu.Unlock()
	return ImportStatusResponse{
		Status:        s.status,
		StartedAt:     timePtrRFC3339(s.startedAt),
		FinishedAt:    timePtrRFC3339(s.finishedAt),
		SourceFile:    s.sourceFile,
		ProcessedRows: s.processedRows,
		UpsertedRows:  s.upsertedRows,
		Error:         s.lastErr,
	}
}

func (s *movieImportJobState) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func findLatestMovieIDsGZ(dataDir string) (string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return "", fmt.Errorf("read data directory: %w", err)
	}

	var latestPath string
	var latestName string
	var latestModTime time.Time
	found := false

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return "", fmt.Errorf("read file info for %q: %w", entry.Name(), err)
		}

		if !found || info.ModTime().After(latestModTime) || (info.ModTime().Equal(latestModTime) && entry.Name() > latestName) {
			found = true
			latestModTime = info.ModTime()
			latestName = entry.Name()
			latestPath = filepath.Join(dataDir, entry.Name())
		}
	}

	if !found {
		return "", errNoMovieIDFiles
	}

	return latestPath, nil
}

func resolveDataDir() string {
	if _, err := os.Stat("data"); err == nil {
		return "data"
	}
	if _, err := os.Stat(filepath.Join("server", "data")); err == nil {
		return filepath.Join("server", "data")
	}
	return "data"
}

func copyMovieIDChunk(ctx context.Context, tx pgx.Tx, rows [][]any) error {
	if len(rows) == 0 {
		return nil
	}

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"movie_ids_import_staging"},
		[]string{"id", "original_title", "adult", "video", "popularity"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("copy to staging table failed: %w", err)
	}
	return nil
}

func runMovieIDsImport(ctx context.Context, pool *pgxpool.Pool, sourcePath string, state *movieImportJobState) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open import file: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip reader: %w", err)
	}
	defer gzReader.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, createMovieIDImportStagingSQL); err != nil {
		return fmt.Errorf("create staging table: %w", err)
	}

	scanner := bufio.NewScanner(gzReader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var processedRows int64
	lineNumber := 0
	copyRows := make([][]any, 0, importCopyBatchSize)

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		processedRows++

		var row MovieIDImportRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			return fmt.Errorf("line %d: invalid JSON: %w", lineNumber, err)
		}
		if row.ID <= 0 {
			return fmt.Errorf("line %d: id must be greater than zero", lineNumber)
		}
		if strings.TrimSpace(row.OriginalTitle) == "" {
			return fmt.Errorf("line %d: original_title is required", lineNumber)
		}

		copyRows = append(copyRows, []any{
			row.ID,
			row.OriginalTitle,
			row.Adult,
			row.Video,
			row.Popularity,
		})

		if len(copyRows) >= importCopyBatchSize {
			if err := copyMovieIDChunk(ctx, tx, copyRows); err != nil {
				return fmt.Errorf("line %d: %w", lineNumber, err)
			}
			copyRows = copyRows[:0]
			state.updateProgress(processedRows, processedRows)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan gzip payload: %w", err)
	}

	if err := copyMovieIDChunk(ctx, tx, copyRows); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, mergeMovieIDImportStagingSQL); err != nil {
		return fmt.Errorf("merge staging table into movie_ids: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	state.updateProgress(processedRows, processedRows)
	return nil
}
