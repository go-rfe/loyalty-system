.ONESHELL:
SHELL = /bin/bash
.PHONY: build clean update compile update clean get build clean-cache tidy

SERVER_BINNAME=server

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

SERVER_SOURCE=$(GOBASE)/

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## compile: Compile the binary.
build:
	$(MAKE) -s compile

## clean: Clean build files. Runs `go clean` internally.
clean:
	$(MAKE) go-clean

## update: Update modules
update:
	$(MAKE) go-update

test: go-test go-statictest go-vet

compile: go-clean go-get-server build-server

go-update: go-clean-cache go-tidy go-migrate

go-clean:
	@echo "  >  Cleaning build cache"
	@GOBIN=$(GOBIN) go clean
	@rm -rf $(GOBIN)

go-get-server:
	@echo "  >  Checking if there is any missing dependencies..."
	@cd $(SERVER_SOURCE); GOBIN=$(GOBIN) go get $(get)

build-server:
	@echo "  >  Building server binaries..."
	@cd $(SERVER_SOURCE); go build -o $(GOBIN)/$(SERVER_BINNAME) $(GOFILES)

go-clean-cache:
	@echo "  >  Clean modules cache..."
	@go clean -modcache

go-tidy:
	@echo "  >  Update modules..."
	@go mod tidy

go-test:
	@echo "  >  Test project..."
	@go test ./...

go-container:
	@echo "  >  Building docker image..."
	@docker build -q -t go-loyalty-server .
	@docker run --name go-loyalty-server go-loyalty-server

go-statictest: go-container
	@echo " > Static test project..."
	@docker logs go-loyalty-server

go-vet:
	@echo "  >  Vet project..."
	@go vet ./...

go-migrate:
	@echo "  >  Update migrations..."
	@go get -u github.com/go-bindata/go-bindata/...
	@cd internal/db/migrations; go-bindata -pkg migrations .