package main

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func registerMovieRoutes(e *echo.Echo, queries *db.Queries, pool *pgxpool.Pool, importState *movieImportJobState, dataDir string) {
	e.GET("/api/movies/search", func(c echo.Context) error {
		q := c.QueryParam("q")
		if q == "" {
			return c.JSON(http.StatusOK, []MovieResult{})
		}

		results, err := queries.SearchMovies(c.Request().Context(), q)
		if err != nil {
			log.Printf("search error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to search movies",
			})
		}

		movies := make([]MovieResult, len(results))
		for i, r := range results {
			pop, _ := r.Popularity.Float64Value()
			popularity := 0.0
			if pop.Valid {
				popularity = math.Round(pop.Float64*10000) / 10000
			}
			movies[i] = MovieResult{
				ID:            r.ID,
				OriginalTitle: r.OriginalTitle,
				Adult:         r.Adult,
				Video:         r.Video,
				Popularity:    popularity,
				Score:         r.Score,
			}
		}

		return c.JSON(http.StatusOK, movies)
	})

	e.POST("/api/admin/movies/import", func(c echo.Context) error {
		if importState.isRunning() {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "movie import is already running",
			})
		}

		sourceFile, err := findLatestMovieIDsGZ(dataDir)
		if err != nil {
			if errors.Is(err, errNoMovieIDFiles) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "no movie id .json.gz files found in data directory",
				})
			}
			log.Printf("find latest movie ids file error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to locate latest movie data file",
			})
		}

		if !importState.startIfIdle(sourceFile) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "movie import is already running",
			})
		}
		log.Printf("movie import started: source_file=%s", sourceFile)

		go func() {
			if err := runMovieIDsImport(context.Background(), pool, sourceFile, importState); err != nil {
				snapshot := importState.snapshot()
				importState.finishFailure(snapshot.ProcessedRows, snapshot.UpsertedRows, err.Error())
				log.Printf(
					"movie import failed: source_file=%s processed_rows=%d upserted_rows=%d err=%v",
					sourceFile,
					snapshot.ProcessedRows,
					snapshot.UpsertedRows,
					err,
				)
				return
			}

			snapshot := importState.snapshot()
			importState.finishSuccess(snapshot.ProcessedRows, snapshot.UpsertedRows)
			log.Printf(
				"movie import succeeded: source_file=%s processed_rows=%d upserted_rows=%d",
				sourceFile,
				snapshot.ProcessedRows,
				snapshot.UpsertedRows,
			)
		}()

		status := importState.snapshot()
		return c.JSON(http.StatusAccepted, map[string]any{
			"status":      status.Status,
			"started_at":  status.StartedAt,
			"source_file": status.SourceFile,
		})
	})

	e.GET("/api/admin/movies/import/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, importState.snapshot())
	})
}
