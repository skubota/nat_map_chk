BIN := nat_map_chk
GOBIN ?= $(shell go env GOPATH)/bin

NAME    := nat_map_chk
VERSION := 0.1
MINVER  :=$(shell date -u +.%Y%m%d)
BUILD_LDFLAGS := "-X main.Name=$(NAME) -X main.Version=$(VERSION)$(MINVER)" 

.PHONY: all
all: clean build

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) -o $(BIN) 

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) 

.PHONY: deps
deps:
	go get gortc.io/stun

.PHONY: lint
lint: $(GOBIN)/golint
	go fmt
	go vet 
	$(GOBIN)/golint -set_exit_status 

$(GOBIN)/golint:
	go get golang.org/x/lint/golint

.PHONY: clean
clean:
	rm -rf $(BIN)
	go clean
