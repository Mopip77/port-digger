.PHONY: build clean test run

APP_NAME = PortDigger

build:
	go build -ldflags="-s -w" -o $(APP_NAME) .

build-release:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(APP_NAME) .

test:
	go test ./... -v

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME)
	go clean

install: build
	mkdir -p ~/Applications
	cp $(APP_NAME) ~/Applications/

.DEFAULT_GOAL := build
