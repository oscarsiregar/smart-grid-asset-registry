.PHONY: run build test cover lint fmt vet mocks tidy check

MODULE := smart-grid-asset-registry
# Same exclusions as .gitlab-ci.yml go_test.
TEST_PKGS = $(shell go list ./... | grep -vE '^$(MODULE)/(cmd(/.*)?|internal/bootstrap$$|internal/routes$$|internal/mocks(/.*)?|internal/repository/psql/model$$|domain/(repository|service|dto|entity|errorcode)$$)')

run:            ## run the API locally (reads .env)
	go run ./cmd/api

build:          ## build the binary into bin/
	CGO_ENABLED=0 go build -trimpath -o bin/api ./cmd/api

test:           ## run unit tests
	go test $(TEST_PKGS)

cover:          ## usecase coverage (CI gate: >= 80%)
	go test ./internal/usecase/... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

lint:           ## golangci-lint v2
	golangci-lint run --timeout=10m

fmt:            ## format code
	gofmt -w .

vet:
	go vet ./...

mocks:          ## regenerate mocks from domain interfaces (mockery v2.53+)
	mockery

tidy:
	go mod tidy

check: fmt vet test   ## run before every push
