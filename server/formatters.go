package main

import (
	"time"

	db "github.com/seanlee/moviestack/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

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

func int4Ptr(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	v := value.Int32
	return &v
}

func dateISO(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func timePtrRFC3339(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	s := value.UTC().Format(time.RFC3339)
	return &s
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

func toMovieLogResponse(logEntry db.ListMovieLogByUserRow) MovieLogResponse {
	return MovieLogResponse{
		LogID:         logEntry.LogID,
		UserID:        logEntry.UserID,
		MovieID:       logEntry.MovieID,
		OriginalTitle: logEntry.OriginalTitle,
		WatchedOn:     dateISO(logEntry.WatchedOn),
		Note:          textPtr(logEntry.Note),
		RankPosition:  int4Ptr(logEntry.RankPosition),
		CreatedAt:     timestamptzRFC3339(logEntry.CreatedAt),
		UpdatedAt:     timestamptzRFC3339(logEntry.UpdatedAt),
	}
}
