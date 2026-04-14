# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder
WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/hr-backend .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/hr-backend /app/hr-backend
COPY adapters/db/migrations /app/adapters/db/migrations
COPY assets /app/assets

EXPOSE 8080

ENV BANU_MUSA_PORT=8080 \
    BANU_MUSA_DB_PATH=/app/data/banumusa.db \
    BANU_MUSA_DB_MIGRATIONS_PATH=/app/adapters/db/migrations \
    BANU_MUSA_FONT_PATH=/app/assets/fonts/Noto_Sans_Arabic/static/NotoSansArabic-Regular.ttf

VOLUME ["/app/data"]

ENTRYPOINT ["/app/hr-backend"]
