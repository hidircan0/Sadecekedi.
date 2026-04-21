# Cat Uploader (Go + HTMX)

This project is a modular web skeleton built with Go's standard `net/http` package and HTMX.

## Run

```bash
go run ./cmd/web
```

Server starts on `http://localhost:8080`.

## Structure

- `cmd/web`: application entrypoint and dependency wiring
- `internal/http`: handlers and router
- `internal/app`: use-case/service layer
- `internal/storage/local`: local filesystem repository (`uploads`)
- `internal/domain`: domain entities
- `web/templates`: HTML templates and HTMX partials
- `web/static`: static CSS

## Future integration points

- **YOLO11 microservice**: add an adapter under `internal/integrations/yolo` and call it from the service layer (`internal/app`) after upload.
- **MinIO**: add a repository implementation under `internal/storage/minio` that satisfies the same repository contract used by `PhotoService`.
- Current implementation keeps local storage as an infrastructure detail behind the repository interface.
