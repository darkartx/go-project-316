help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

tidy: ## Tidy up dependencies, format code, and run vet
	go mod tidy
	go fmt ./...
	go vet ./...

.PHONY: test
test: tidy ## Run all tests
	go test -v --race ./...

test_coverage: tidy ## Run all tests with coverage
	go test -v -coverprofile=coverage.out --race ./...

install: ## Install app to system
	go install

lint: ## Lint code
	golangci-lint run ./...

build: ## Build app
	go build -o bin/hexlet-go-crawler ./cmd/hexlet-go-crawler

run: ## Run app
	go run ./cmd/hexlet-go-crawler ${URL} --depth=0
