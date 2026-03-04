# Langues Backend

Go API phục vụ xử lý tự động cho bảng từ vựng kiểu Excel.

## Chạy local

```bash
go run ./cmd/server
```

## Test và build

```bash
go test ./...
go build ./...
```

## Endpoints

- `GET /healthz`
- `GET /api/v1/vocabularies` (giữ lại cho dữ liệu seed)
- `POST /api/v1/vocabularies/enrich`

Ví dụ payload:

```json
{
  "input": "language"
}
```

## Tối ưu hiện có

- Cache in-memory để giảm số lần gọi API ngoài cho từ đã xử lý.
- Không lưu DB trong phiên bản hiện tại.

## Biến môi trường

- `PORT` (mặc định `8080`)
- `CORS_ALLOWED_ORIGINS` (mặc định `http://localhost:3000,http://127.0.0.1:3000`)
