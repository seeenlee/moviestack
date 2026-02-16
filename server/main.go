package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	loadEnvFiles()
	databaseURL := buildDatabaseURL()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("unable to ping database: %v", err)
	}
	fmt.Println("Connected to database")

	queries := db.New(pool)
	importState := &movieImportJobState{status: "idle"}
	dataDir := resolveDataDir()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	registerMovieRoutes(e, queries, pool, importState, dataDir)
	registerAdminUserRoutes(e, queries)
	registerMovieLogRoutes(e, queries)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on :%s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
