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

# Install as a proper macOS app bundle
install-app: build
	@echo "Creating $(APP_NAME).app bundle..."
	mkdir -p ~/Applications/$(APP_NAME).app/Contents/MacOS
	mkdir -p ~/Applications/$(APP_NAME).app/Contents/Resources
	cp $(APP_NAME) ~/Applications/$(APP_NAME).app/Contents/MacOS/
	cp Info.plist ~/Applications/$(APP_NAME).app/Contents/
	@echo "✅ Installed to ~/Applications/$(APP_NAME).app"
	@echo "You can now find 'Port Digger' in Spotlight/Launcher"

.DEFAULT_GOAL := build
