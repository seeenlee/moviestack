package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func loadEnvFiles() {
	candidates := []string{".env", filepath.Join("server", ".env")}
	for _, candidate := range candidates {
		if err := loadDotEnv(candidate); err == nil {
			return
		}
	}
}

func buildDatabaseURL() string {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		return databaseURL
	}

	dbPass := strings.Trim(os.Getenv("DB_PASS"), "\"'")
	// sslmode=disable is for local dev only; enable SSL in production.
	databaseURL = "postgres://moviestack@127.0.0.1:5432/moviestack_dev?sslmode=disable"
	if dbPass != "" {
		databaseURL = fmt.Sprintf(
			"postgres://moviestack:%s@127.0.0.1:5432/moviestack_dev?sslmode=disable",
			url.QueryEscape(dbPass),
		)
	}

	return databaseURL
}
