BINARY_NAME=checker
SOURCE_DIR=./src
PKGS=./src/utils
BIN_DIR=./bin

# OS detection
UNAME_S := $(shell uname -s)
CURRENT_OS := unknown
CURRENT_ARCH := $(shell uname -m)

ifeq ($(OS),Windows_NT)
    CURRENT_OS := windows
    BINARY_EXTENSION := .exe
else ifeq ($(UNAME_S),Darwin)
    CURRENT_OS := darwin
else ifeq ($(UNAME_S),Linux)
    CURRENT_OS := linux
endif

# Build configuration
GOARCH := amd64
ifeq ($(CURRENT_ARCH),arm64)
    GOARCH := arm64
endif

# Common tasks
dependencies:
	go get -d ./...

fmt:
	gofmt -s -l -w $(PKGS)

vet:
	go vet $(PKGS)

lint:
	golangci-lint run -c .golangci.yml $(PKGS)

staticcheck:
	staticcheck $(PKGS)

<<<<<<< HEAD
# Generic build function
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME)$(BINARY_EXTENSION) ./main.go
=======
#TODO REPLACE THESE WITH A BUILD SCRIPT
build-linux:
# Don't ask why the indent is haywire. Make just blows up otherwise
ifeq ($(OS), Windows_NT)
	set GOOS=linux
	set GOARCH=amd64
	go build -o ./bin/${BINARY_NAME} ./main.go
else ifeq ($(UNAME_S), Linux)
	GOARCH=amd64 GOOS=linux go build -o ./bin/${BINARY_NAME} ./main.go
else
	@echo "OS not supported."
endif
>>>>>>> 3d170ca (fix: changed the naming of gitlab actions and golanglinter files)

# Platform-specific builds
build-linux: GOOS=linux
build-linux: BINARY_EXTENSION=
build-linux: build

build-macos: GOOS=darwin
build-macos: BINARY_EXTENSION=
build-macos: build

build-windows: GOOS=windows
build-windows: BINARY_EXTENSION=.exe
build-windows: build

dev:
	go run ./main.go

clean:
	go clean
	rm -f $(BIN_DIR)/*

.PHONY: dependencies fmt vet lint staticcheck build build-linux build-macos build-windows dev clean
