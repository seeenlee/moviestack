package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type MovieResult struct {
	ID            int32   `json:"id"`
	OriginalTitle string  `json:"original_title"`
	Adult         bool    `json:"adult"`
	Video         bool    `json:"video"`
	Popularity    float64 `json:"popularity"`
	Score         float32 `json:"score"`
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// sslmode=disable is for local dev only; enable SSL in production.
		databaseURL = "postgres://localhost:5432/moviestack?sslmode=disable"
	}

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

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{http.MethodGet},
	}))

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on :%s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
