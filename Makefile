APP?=web-crawler
TEST_SERVER?=web-server-integration-test

.PHONY: build
## build: builds the application
build: clean
	@echo "Building..."
	@go build -o ${APP} cmd/*.go

.PHONY: run
## run: runs main.go
run:
	go run cmd/main.go

.PHONY: clean
## clean: removes the binary
clean:
	@echo "Cleaning"
	@rm -rf ${APP}

.PHONY: test
## test: runs go test with default values
test: stop serve
	go test -v -count=1 -race -cover ./...

.PHONY: utest
## utest: runs go local unit-tests, no integration test is run
utest:
	go test -v -count=1 -race -cover -short ./...

.PHONY: build-server
## build-server: builds the web-server container in test_data to be used in the integration tests
build-server: stop
	@echo "Building integration tests web-server"
	@docker build -t ${TEST_SERVER} test_data/

.PHONY: serve
## serve: builds and runs the web-server in the test_data folder to allow for integration tests
serve: build-server
	@echo "Starting integration tests web-server"
	@docker run -it -d --rm -p 0.0.0.0:8080:8080 ${TEST_SERVER}

.PHONY: stop
## stop: stops the web-server for integration tests
stop:
	@echo "Stopping container"
	@docker stop $$(docker ps -a -q --filter ancestor=${TEST_SERVER} --format="{{.ID}}") || true

.PHONY: bench
## bench: runs benchmarks
bench:
	go test -v -count=1 -bench=. ./... -run NONE

.PHONY: benchmem
## benchmem: runs processing and memory benchmarks
benchmem:
	go test -v -count=1 -bench=. ./... -benchmem -run NONE

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | sort | column -t -s ':' |  sed -e 's/^/ /'

