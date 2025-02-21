BINARY_NAME=checker
SOURCE_DIR=./src


UNAME_S := $(shell uname -s)

dependencies:
	go get -d ./...

fmt:
	gofmt -s -l -w $(pkgs)

vet:
	go vet $(pkgs)

lint:
	golangci-lint run -c .golangci.yml $(pkgs)

staticcheck:
	staticcheck $(pkgs)

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

build-windows:
# Don't ask why the indent is haywire. Make just blows up otherwise
ifeq ($(OS), Windows_NT)
	set GOOS=windows
	set GOARCH=amd64
	go build -o ./bin/${BINARY_NAME}.exe ./main.go
else ifeq ($(UNAME_S), Linux)
	GOARCH=amd64 GOOS=windows go build -o ./bin/${BINARY_NAME}.exe ./main.go
else
	@echo "OS not supported."
endif

dev:
	go run ./main.go

clean:
	go clean
	rm ./bin/*
