# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder
WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/hr-backend .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /out/hr-backend /app/hr-backend
COPY --chown=nonroot:nonroot adapters/db/migrations /app/adapters/db/migrations
COPY --chown=nonroot:nonroot assets /app/assets

EXPOSE 8080

ENV BANU_MUSA_PORT=8080 \
    BANU_MUSA_DB_MIGRATIONS_PATH=/app/adapters/db/migrations \
    BANU_MUSA_FONT_PATH=/app/assets/fonts/Noto_Sans_Arabic/static/NotoSansArabic-Regular.ttf \
    BANU_MUSA_FCM_ENABLED=false \
    BANU_MUSA_AUTH_ENABLED=true \
    BANU_MUSA_DEV_OTP_BYPASS=true \
    BANU_MUSA_DEV_BYPASS_OTP=112233

# PostgreSQL configuration (update these for production)
ENV BANU_MUSA_POSTGRES_HOST=localhost \
    BANU_MUSA_POSTGRES_PORT=5432 \
    BANU_MUSA_POSTGRES_USER=banumusa \
    BANU_MUSA_POSTGRES_PASSWORD=banumusa-secret \
    BANU_MUSA_POSTGRES_DB=banumusa \
    BANU_MUSA_POSTGRES_SSLMODE=disable \
    BANU_MUSA_PGBOUNCER_ENABLED=true \
    BANU_MUSA_PGBOUNCER_HOST=localhost \
    BANU_MUSA_PGBOUNCER_PORT=6432

ENTRYPOINT ["/app/hr-backend"]