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

COPY --from=builder /twitch-miner-go /twitch-miner-go
# Only example configs are copied; real configs should be mounted via volume or created at runtime
COPY --from=builder /app/configs /configs

EXPOSE 8080

ENTRYPOINT ["/twitch-miner-go"]
CMD ["-config", "/configs", "-port", "8080"]
