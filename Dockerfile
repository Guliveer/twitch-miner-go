# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
ARG VERSION=dev
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/Guliveer/twitch-miner-go/internal/version.Number=${VERSION} -X github.com/Guliveer/twitch-miner-go/internal/version.GitCommit=${GIT_COMMIT}" \
    -o /twitch-miner-go ./cmd/twitch-miner-go

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.description="Efficient auto drops and points claim for Twitch"

COPY --from=builder /twitch-miner-go /twitch-miner-go
# Only the example config is copied; real configs are mounted via volume or
# created at runtime. .dockerignore already keeps them out of the build context.
COPY --from=builder /app/configs/example.yaml.example /configs/example.yaml.example

EXPOSE 8080

# The container is headless (no desktop/DBus session); the system tray icon
# has no surface to attach to, so it is disabled by default.
ENV NO_TRAY=true

ENTRYPOINT ["/twitch-miner-go"]
CMD ["-config", "/configs", "-port", "8080"]
