SERVICE_NAME=celeste
-include .env
export

.PHONY: setup
setup: ## Get linting stuffs
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	go get golang.org/x/tools/cmd/goimports

.PHONY: build
build: lint ## Build the app
	go build -ldflags "-w -s -X github.com/bugfixes/${SEVICE_NAME}/internal/app.version=`git describe --tags --dirty` -X github.com/bugfixes/${SERVICE_NAME}/internal/app.commitHash=`git rev-parse HEAD`" -race -o ./bin/${SERVICE_NAME} -v ./cmd/${SERVICE_NAME}/${SEVICE_NAME}.go

.PHONY: test
test: lint ## Test the app
	go test -v -race -bench=./... -benchmem -timeout=120s -cover -coverprofile=./test/coverage.txt -bench=./... ./...

.PHONY: run
run: build ## Build and run
	bin/${SERVICE_NAME}

.PHONY: lambda
lambda: ## Run the lambda version
	go build ./cmd/main

.PHONY: mocks
mocks: ## Generate the mocks
	go generate ./...

.PHONY: full
full: clean build fmt lint test ## Clean, build, make sure its formatted, linted, and test it

.PHONY: docker-up
docker-up: ## Start docker
	docker-compose -p ${SERVICE_NAME} --project-directory=docker -f docker/docker-compose.yml up -d
	sleep 60
	go run ./docker/docker.go

.PHONY: docker-down
docker-down: ## Stop docker
	docker-compose -p ${SERVICE_NAME} --project-directory=docker -f docker/docker-compose.yml down

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker

.PHONY: lint
lint: ## Lint
	golangci-lint run --config configs/golangci.yml

.PHONY: fmt
fmt: ## Formatting
	gofmt -w -s .
	goimports -w .
	go clean ./...

.PHONY: pre-commit
pre-commit: fmt lint ## Do formatting and linting

.PHONY: clean
clean: ## Clean
	go clean ./...
	rm -rf bin/${SERVICE_NAME}
