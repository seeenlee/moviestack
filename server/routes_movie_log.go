package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

func registerMovieLogRoutes(e *echo.Echo, queries *db.Queries) {
	e.GET("/api/users/:userId/log", func(c echo.Context) error {
		userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		userExists, err := queries.UserExists(c.Request().Context(), userID)
		if err != nil {
			log.Printf("user exists error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to verify user",
			})
		}
		if !userExists {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		results, err := queries.ListMovieLogByUser(c.Request().Context(), userID)
		if err != nil {
			log.Printf("list movie log error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to list movie log",
			})
		}

		response := make([]MovieLogResponse, len(results))
		for i, item := range results {
			response[i] = toMovieLogResponse(item)
		}

		return c.JSON(http.StatusOK, response)
	})

	e.POST("/api/users/:userId/log", func(c echo.Context) error {
		userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		userExists, err := queries.UserExists(c.Request().Context(), userID)
		if err != nil {
			log.Printf("user exists error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to verify user",
			})
		}
		if !userExists {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		var req UpsertMovieLogRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}
		if req.MovieID <= 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "movie_id is required",
			})
		}

		movieExists, err := queries.MovieExists(c.Request().Context(), req.MovieID)
		if err != nil {
			log.Printf("movie exists error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to verify movie",
			})
		}
		if !movieExists {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "movie not found",
			})
		}

		watchedOn := pgtype.Date{Time: time.Now().UTC(), Valid: true}
		if req.WatchedOn != nil && strings.TrimSpace(*req.WatchedOn) != "" {
			parsedDate, err := time.Parse("2006-01-02", strings.TrimSpace(*req.WatchedOn))
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "watched_on must be in YYYY-MM-DD format",
				})
			}
			watchedOn = pgtype.Date{Time: parsedDate, Valid: true}
		}

		note := pgtype.Text{}
		if req.Note != nil && strings.TrimSpace(*req.Note) != "" {
			note = pgtype.Text{String: strings.TrimSpace(*req.Note), Valid: true}
		}

		entry, err := queries.UpsertMovieLogEntry(c.Request().Context(), db.UpsertMovieLogEntryParams{
			UserID:    userID,
			MovieID:   req.MovieID,
			WatchedOn: watchedOn,
			Note:      note,
		})
		if err != nil {
			log.Printf("upsert movie log error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to save movie log entry",
			})
		}

		return c.JSON(http.StatusOK, map[string]any{
			"log_id":        entry.ID,
			"user_id":       entry.UserID,
			"movie_id":      entry.MovieID,
			"watched_on":    dateISO(entry.WatchedOn),
			"note":          textPtr(entry.Note),
			"rank_position": int4Ptr(entry.RankPosition),
			"created_at":    timestamptzRFC3339(entry.CreatedAt),
			"updated_at":    timestamptzRFC3339(entry.UpdatedAt),
		})
	})

	e.DELETE("/api/users/:userId/log/:logId", func(c echo.Context) error {
		userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid user id",
			})
		}

		logID, err := strconv.ParseInt(c.Param("logId"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid log id",
			})
		}

		rowsAffected, err := queries.DeleteMovieLogEntry(c.Request().Context(), db.DeleteMovieLogEntryParams{
			ID:     logID,
			UserID: userID,
		})
		if err != nil {
			log.Printf("delete movie log error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "failed to delete movie log entry",
			})
		}

		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "movie log entry not found",
			})
		}

		return c.NoContent(http.StatusNoContent)
	})
}
