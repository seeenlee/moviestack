# moviestack

- An app that allows users to rank all the movies that they've watched
- Users can add comments and log when they watched the movie
- Users can add friends to see the activity of their friends

## Database migrations (Goose)

Install Goose:

```bash
brew install goose
```

Run from `server/`:

```bash
make migrate-status
make migrate-up
make migrate-down
```

Use a custom database URL:

```bash
make migrate-up DB_URL="postgres://USER:PASSWORD@localhost:5432/moviestack?sslmode=disable"
```

Create a new migration:

```bash
make migrate-create name=add_watch_logs
```
