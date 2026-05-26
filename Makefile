GO ?= go
BINARY_NAME ?= unifi-bootstrapper
BUILD_OUTPUT ?= /tmp/$(BINARY_NAME)
COVERAGE_OUTPUT ?= /tmp/coverage.out
VERSION ?= dev

.PHONY: build
build:
	$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_OUTPUT) ./cmd/bootstrapper/

.PHONY: test
test:
	$(GO) test -v -race ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: coverage
coverage:
	$(GO) test -coverprofile=$(COVERAGE_OUTPUT) ./...
	$(GO) tool cover -func=$(COVERAGE_OUTPUT)

.PHONY: ci-checks
ci-checks: lint test coverage

.PHONY: ci-all
ci-all: ci-checks

.PHONY: clean
clean:
	rm -f $(BUILD_OUTPUT) $(COVERAGE_OUTPUT)
