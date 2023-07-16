.PHONY: ckeck install upload

BUILD_VERSION ?= $(shell cat VERSION)
BUILD_OUTPUT ?= gobrot
GO ?= GO111MODULE=on CGO_ENABLED=0 go
GOOS ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f2)

run:
	@$(GO) run main.go

build:
	@$(GO) build -o $(BUILD_OUTPUT) main.go

clean:
	@echo -n ">> CLEAN"
	@$(GO) clean -i ./...
	@rm -f esimov-*-*
	@rm -rf dist/*
	@printf '%s\n' '$(OK)'


crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-linux-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-windows-amd64
	@printf '%s\n' '$(OK)'

	@echo -n ">> CROSSCOMPILE linux/arm64"
	@GOOS=linux GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-linux-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/arm64"
	@GOOS=darwin GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/arm64"
	@GOOS=windows GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-windows-arm64
	@printf '%s\n' '$(OK)'
