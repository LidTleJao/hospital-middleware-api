# syntax=docker/dockerfile:1

# ---------- base: dependencies only, so this layer is cached across code edits
FROM golang:1.24-alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download

# ---------- dev: hot reload with air, source is bind-mounted at runtime
FROM base AS dev
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# ---------- builder: compile a static binary
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api ./cmd/api

# ---------- runtime: only the binary ships, no Go toolchain
FROM alpine:3.21 AS runtime
RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /out/api /app/api
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/api"]
