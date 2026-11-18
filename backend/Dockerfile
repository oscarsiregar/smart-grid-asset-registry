# ============================
# Stage 1: Build
# ============================
FROM golang:1.26.9-alpine AS build

RUN apk add --no-cache git

ARG CI_JOB_TOKEN

# Lets `go mod download` fetch the private go-common module straight from
# repo.smf-indonesia.co.id with the CI job token (no sibling checkout).
RUN git config --global url."https://gitlab-ci-token:${CI_JOB_TOKEN}@repo.smf-indonesia.co.id/".insteadOf "https://repo.smf-indonesia.co.id/"

# Never route the private module through the public proxy / checksum DB.
ENV GOPRIVATE=repo.smf-indonesia.co.id

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

# curl: in-container debugging of /v1/health. tzdata: Asia/Jakarta timestamps.
RUN apk add --no-cache curl tzdata ca-certificates \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=build --chown=appuser:appgroup /app/main .

USER appuser

EXPOSE 3030
# config.Load() uses viper.AutomaticEnv(): real container env vars
# (Kubernetes Secret/ConfigMap) are read directly — never bake a .env
# with secrets into the image.
CMD ["./main"]
