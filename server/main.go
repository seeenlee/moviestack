package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

type AdminUserResponse struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type CreateAdminUserRequest struct {
	Username string `json:"username"`
}

func textPtr(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func timestamptzRFC3339(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func toAdminUserResponse(user db.User) AdminUserResponse {
	return AdminUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: textPtr(user.DisplayName),
		Bio:         textPtr(user.Bio),
		AvatarURL:   textPtr(user.AvatarUrl),
		CreatedAt:   timestamptzRFC3339(user.CreatedAt),
		UpdatedAt:   timestamptzRFC3339(user.UpdatedAt),
	}
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
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
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

	e.GET("/api/admin/users", func(c echo.Context) error {
		results, err := queries.ListUsers(c.Request().Context())
		if err != nil {
			log.Printf("list users error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to list users",
			})
		}

		users := make([]AdminUserResponse, len(results))
		for i, user := range results {
			users[i] = toAdminUserResponse(user)
		}

		return c.JSON(http.StatusOK, users)
	})

	e.POST("/api/admin/users", func(c echo.Context) error {
		var req CreateAdminUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		username := strings.TrimSpace(req.Username)
		if username == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "username is required",
			})
		}

		user, err := queries.CreateUser(c.Request().Context(), username)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_username_lower_unique" {
				return c.JSON(http.StatusConflict, map[string]string{
					"error": "username already exists",
				})
			}

			log.Printf("create user error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to create user",
			})
		}

		return c.JSON(http.StatusCreated, toAdminUserResponse(user))
	})

	e.DELETE("/api/admin/users/:id", func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		rowsAffected, err := queries.DeleteUser(c.Request().Context(), id)
		if err != nil {
			log.Printf("delete user error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to delete user",
			})
		}

		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		return c.NoContent(http.StatusNoContent)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on :%s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
