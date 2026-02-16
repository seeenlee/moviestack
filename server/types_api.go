package main

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

type MovieLogResponse struct {
	LogID         int64   `json:"log_id"`
	UserID        int64   `json:"user_id"`
	MovieID       int32   `json:"movie_id"`
	OriginalTitle string  `json:"original_title"`
	WatchedOn     string  `json:"watched_on"`
	Note          *string `json:"note"`
	RankPosition  *int32  `json:"rank_position"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type UpsertMovieLogRequest struct {
	MovieID   int32   `json:"movie_id"`
	WatchedOn *string `json:"watched_on"`
	Note      *string `json:"note"`
}

type MovieIDImportRow struct {
	ID            int32   `json:"id"`
	OriginalTitle string  `json:"original_title"`
	Adult         bool    `json:"adult"`
	Video         bool    `json:"video"`
	Popularity    float64 `json:"popularity"`
}

type ImportStatusResponse struct {
	Status        string  `json:"status"`
	StartedAt     *string `json:"started_at"`
	FinishedAt    *string `json:"finished_at"`
	SourceFile    string  `json:"source_file"`
	ProcessedRows int64   `json:"processed_rows"`
	UpsertedRows  int64   `json:"upserted_rows"`
	Error         string  `json:"error"`
}
