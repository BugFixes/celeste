SERVICE_NAME=celeste
STACK_TIME=$(shell date "+%y-%m-%d_%H-%M")
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
docker-up: docker-start sleepy ## Start docker

docker-start: ## Docker Start
	docker compose -p ${SERVICE_NAME} --project-directory=docker -f docker-compose.yml up -d

docker-stop: ## Docker Stop
	docker compose -p ${SERVICE_NAME} --project-directory=docker -f docker-compose.yml down

.PHONY: docker-down
docker-down: docker-stop ## Stop docker

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker

.PHONY: docker-logs
docker-logs: ## Follow the logs
	docker logs -f ${SERVICE_NAME}_localstack_1

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

sleepy: ## Sleepy
	sleep 60

.PHONY: cloud-up
cloud-up: docker-start sleepy stack-create ## CloudFormation Up

.PHONY: cloud-restart
cloud-restart: docker-down cloud-up

.PHONY: stack-create
stack-create: # Create the stack
	aws cloudformation create-stack \
  		--template-body file://docker/cloudformation.yaml \
  		--stack-name ${SERVICE_NAME}-$(STACK_TIME) \
  		--endpoint https://localhost.localstack.cloud:4566 \
  		--region us-east-1 \
  		--parameters \
  		  ParameterKey=GithubKey,ParameterValue=${GITHUB_CLIENT_ID} \
  		  ParameterKey=GithubSecret,ParameterValue=${GITHUB_CLIENT_SECERT} \
  		1> /dev/null

.PHONY: stack-delete
stack-delete: # Delete the stack
	aws cloudformation delete-stack \
		--stack-name ${SERVICE_NAME}-$(STACK_TIME) \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1

.PHONY: injectData
injectData: # Inject Agent
	export $(egrep -v '^#' .env | xargs)

	aws dynamodb put-item \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1 \
		--table-name ticketing \
		--item '{"access_token":{"S":"${GITHUB_ACCESS_TOKEN}"},"id":{"S":"github_token"},"system":{"S":"github"},"agent_id":{"S":"${BUGFIXES_GITHUB_AGENT_ID}"},"ticketing_details":{"M":{"owner":{"S":"bugfixes"},"repo":{"S":"celeste"},"installation_id":{"S":"${GITHUB_INSTALL_ID}"}}}}' 1> /dev/null
	aws dynamodb put-item \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1 \
		--table-name ticketing \
		--item '{"access_token":{"S":"${JIRA_API_TOKEN}"},"id":{"S":"jira_token"},"system":{"S":"jira"},"agent_id":{"S":"${BUGFIXES_JIRA_AGENT_ID}"},"ticketing_details":{"M":{"project_name":{"S":"${JIRA_PROJECT_NAME}"},"username":{"S":"${JIRA_USERNAME}"},"host":{"S":"${JIRA_HOST}"},"project_key":{"S":"${JIRA_PROJECT_KEY}"}}}}' 1>/dev/null
	aws dynamodb put-item \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1 \
		--table-name agents \
		--item '{"id":{"S":"${BUGFIXES_GITHUB_AGENT_ID}"},"name":{"S":"github"},"account_record":{"M":{"id":{"S":"${BUGFIXES_ACCOUNT_ID}"},"name":{"S":"bob"},"email":{"S":"${BUGFIXES_ACCOUNT_EMAIL}"},"level":{"S":"0"},"account_credentials":{"M":{"key":{"S":"${BUGFIXES_ACCOUNT_KEY}"},"secret":{"S":"${BUGFIXES_ACCOUNT_SECRET}"}}}}},"agent_credentials":{"M":{"key":{"S":"${BUGFIXES_GITHUB_AGENT_KEY}"},"secret":{"S":"${BUGFIXES_GITHUB_AGENT_SECRET}"}}}}' 1> /dev/null
	aws dynamodb put-item \
  		--endpoint http://localhost.localstack.cloud:4566 \
  		--region us-east-1 \
  		--table-name agents \
  		--item '{"id":{"S":"${BUGFIXES_JIRA_AGENT_ID}"},"name":{"S":"jira"},"account_record":{"M":{"id":{"S":"${BUGFIXES_ACCOUNT_ID}"},"name":{"S":"bob"},"email":{"S":"${BUGFIXES_ACCOUNT_EMAIL}"},"level":{"S":"0"},"account_credentials":{"M":{"key":{"S":"${BUGFIXES_ACCOUNT_KEY}"},"secret":{"S":"${BUGFIXES_ACCOUNT_SECRET}"}}}}},"agent_credentials":{"M":{"key":{"S":"${BUGFIXES_JIRA_AGENT_KEY}"},"secret":{"S":"${BUGFIXES_JIRA_AGENT_SECRET}"}}}}' 1> /dev/null
	aws dynamodb put-item \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1 \
		--table-name accounts \
		--item '{"id":{"S":"${BUGFIXES_ACCOUNT_ID}"},"name":{"S":"bob"},"email":{"S":"${BUGFIXES_ACCOUNT_EMAIL}"},"level":{"S":"0"},"account_credentials":{"M":{"key":{"S":"${BUGFIXES_ACCOUNT_KEY}"},"secret":{"S":"${BUGFIXES_ACCOUNT_SECRET}"}}}}' 1> /dev/null
	aws dynamodb put-item \
		--endpoint http://localhost.localstack.cloud:4566 \
		--region us-east-1 \
		--table-name comms \
		--item '{"agent_id":{"S":"${BUGFIXES_GITHUB_AGENT_ID}"},"comms_details":{"M":{"channel":{"S":"${DISCORD_TEST_CHANNEL}"}}},"id":{"S":"${BUGFIXES_GITHUB_AGENT_ID}"},"system":{"S":"discord"}}' 1> /dev/null
	aws dynamodb put-item \
  		--endpoint http://localhost.localstack.cloud:4566 \
  		--region us-east-1 \
  		--table-name comms \
  		--item '{"agent_id":{"S":"${BUGFIXES_JIRA_AGENT_ID}"},"comms_details":{"M":{"channel":{"S":"${DISCORD_TEST_CHANNEL}"}}},"id":{"S":"${BUGFIXES_JIRA_AGENT_ID}"},"system":{"S":"discord"}}' 1> /dev/null

.PHONY: bucket-up
bucket-up: bucket-create bucket-upload ## S3 Bucket Up

bucket-create: ## Create the bucket for builds
	aws s3api create-bucket \
		--endpoint https://localhost.localstack.cloud:4566 \
		--bucket celeste \
		--quiet

bucket-upload: build-aws ## Put the build in the bucket
	aws s3 cp bin/celeste-local.zip s3://celeste/celeste-local.zip --endpoint https://localhost.localstack.cloud:4566

build-aws: ## Build for AWS
	GOOS=linux GOARCH=amd64 go build -o bin/celeste ./cmd/main
	zip bin/celeste-local.zip bin/celeste
