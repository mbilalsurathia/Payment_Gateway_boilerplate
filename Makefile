.PHONY: build test run clean docker-build docker-run mock init-db swagger-validate swagger-serve help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=payment-gateway
BINARY_DIR=./bin
MAIN_PATH=./cmd/main.go

help:
	@echo "Payment Gateway Integration System"
	@echo "make build           - Build the application"
	@echo "make test            - Run tests"
	@echo "make run             - Run the application locally"
	@echo "make mock            - Run with mock database"
	@echo "make clean           - Remove binary files"
	@echo "make docker-build    - Build docker image"
	@echo "make docker-run      - Run with docker-compose"
	@echo "make docker-stop     - Stop docker containers"
	@echo "make init-db         - Initialize the database"
	@echo "make swagger-validate - Validate OpenAPI specification"
	@echo "make swagger-serve   - Serve Swagger UI locally"

build:
	mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

run: build
	$(BINARY_DIR)/$(BINARY_NAME)

mock:
	USE_MOCK_DB=true $(GOCMD) run $(MAIN_PATH)

clean:
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

docker-build:
	docker build -t payment-gateway .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Validate OpenAPI specification
swagger-validate:
	swagger validate docs/openapi.yaml

# Serve Swagger UI locally
swagger-serve:
	docker run -p 8081:8080 -e SWAGGER_JSON=/openapi.yaml -v $(PWD)/docs/openapi.yaml:/openapi.yaml swaggerapi/swagger-ui

# Install required tools
install-tools:
	$(GOGET) -u github.com/go-swagger/go-swagger/cmd/swagger

# Tidy up dependencies
tidy:
	$(GOMOD) tidy