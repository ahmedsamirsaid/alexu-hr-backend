# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS builder
WORKDIR /src

# Cache modules first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/hr-backend .
RUN mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /out/hr-backend /app/hr-backend
COPY --from=builder --chown=nonroot:nonroot /out/data /app/data
COPY --chown=nonroot:nonroot adapters/db/migrations /app/adapters/db/migrations
COPY --chown=nonroot:nonroot assets /app/assets

EXPOSE 8080

ENV BANU_MUSA_PORT=8080 \
    BANU_MUSA_DB_PATH=/app/data/banumusa.db \
    BANU_MUSA_DB_MIGRATIONS_PATH=/app/adapters/db/migrations \
    BANU_MUSA_FONT_PATH=/app/assets/fonts/Noto_Sans_Arabic/static/NotoSansArabic-Regular.ttf \
    BANU_MUSA_FCM_ENABLED=false

ENTRYPOINT ["/app/hr-backend"]
