package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

func registerAdminUserRoutes(e *echo.Echo, queries *db.Queries) {
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
}
