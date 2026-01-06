.PHONY: build install test clean lint fmt release

VERSION ?= 0.1.0
BINARY_NAME = envault
BUILD_DIR = bin
RELEASE_DIR = releases
LDFLAGS = -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envault

install:
	go install ./cmd/envault

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)/ $(RELEASE_DIR)/

lint:
	golangci-lint run

fmt:
	go fmt ./...

# Build for all platforms and create release archives
release: clean
	mkdir -p $(RELEASE_DIR)
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envault
	tar -czf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Linux_x86_64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envault
	tar -czf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Linux_arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envault
	tar -czf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Darwin_x86_64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envault
	tar -czf $(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Darwin_arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/envault
	cd $(BUILD_DIR) && zip ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Windows_x86_64.zip $(BINARY_NAME).exe
	# Windows arm64
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/envault
	cd $(BUILD_DIR) && zip ../$(RELEASE_DIR)/$(BINARY_NAME)_$(VERSION)_Windows_arm64.zip $(BINARY_NAME).exe
	@echo "Release archives created in $(RELEASE_DIR)/"
	@ls -la $(RELEASE_DIR)/

# Create GitHub release (requires gh CLI)
github-release: release
	gh release create v$(VERSION) $(RELEASE_DIR)/* --title "v$(VERSION)" --notes "Release v$(VERSION)"
