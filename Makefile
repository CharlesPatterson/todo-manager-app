GOCMD:=$(shell which go)
GOLINT:=$(shell which golangci-lint)
GOIMPORT:=$(shell which goimports)
GOFMT:=$(shell which gofmt)
DOCKER:=$(shell which docker)
SWAG:=$(shell which swag)
DOCKER_TAG:="CharlesPatterson/todoapp"
GOBUILD:=$(GOCMD) build
GOINSTALL:=$(GOCMD) install
GOCLEAN:=$(GOCMD) clean
GOTEST:=$(GOCMD) test
GOMODTIDY:=$(GOCMD) mod tidy
GOGET:=$(GOCMD) get
GOLIST:=$(GOCMD) list
GOVET:=$(GOCMD) vet
GOSEC:=$(shell which gosec)
GOPATH:=$(shell $(GOCMD) env GOPATH)
u := $(if $(update),-u)

BINARY_NAME:=todos-app
GOFILES:=$(shell find . -name "*.go" -type f)
PACKAGES:=$(shell $(GOLIST)	github.com/CharlesPatterson/todos-app/controller github.com/CharlesPatterson/todos-app/middleware github.com/CharlesPatterson/todos-app/model)

all: build

.PHONY: build
build: deps
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

.PHONY: docs
docs:
	$(SWAG) init -g ./cmd/todos-app/main.go
	$(SWAG) fmt -g ./cmd/todos-app/main.go

.PHONY: docker
docker: deps
	$(DOCKER) build -t $(DOCKER_TAG) .

.PHONY: install
install: deps
	$(GOINSTALL) ./cmd/${BINARY_NAME}

.PHONY: lint
lint:
	$(GOLINT) run

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: deps
deps:
	$(GOMODTIDY)

.PHONY: vet
vet: deps
	$(GOVET) $(PACKAGES)

.PHONY: vet
sec: deps
	$(GOSEC) ./...

.PHONY: fmt
fmt:
	$(GOFMT) -s -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;
