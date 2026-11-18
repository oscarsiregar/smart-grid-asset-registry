# ============================
# Stage 1: Build
# ============================
FROM golang:1.26.9-alpine AS build

# go-common is re-implemented under pkg/gocommon (constitution, Template
# Exception 4), so every dependency comes from the public module proxy and
# the build needs no company credentials.

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/main ./cmd/api

# ============================
# Stage 2: Runtime
# ============================
FROM alpine:3.22

# curl: in-container debugging of /api/health. tzdata: time zone data.
RUN apk add --no-cache curl tzdata ca-certificates \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /app/main .

USER appuser

EXPOSE 8080
# config.Load() uses viper.AutomaticEnv(): real container env vars
# (Kubernetes Secret/ConfigMap) are read directly — never bake a .env
# with secrets into the image.
CMD ["./main"]
