
# Detect OS and set binary name accordingly
ifeq ($(OS),Windows_NT)
    BINARY_NAME = smartcopy.exe
else
    BINARY_NAME = smartcopy
endif

build: $(BINARY_NAME)
	@echo "Build complete: $(BINARY_NAME)"

$(BINARY_NAME): main.go
	go build -o $(BINARY_NAME) main.go

clean:
	@echo "Cleaning..."
	@go clean
	@echo "Clean complete"

clean-all:
	@echo "Cleaning all binaries..."
	@go clean
	@echo "All binaries cleaned"

run: 
	go run main.go

test: build
	go run test/main.go

# Build for specific platforms
build-windows:
	GOOS=windows GOARCH=amd64 go build -o smartcopy.exe main.go

build-linux:
	GOOS=linux GOARCH=amd64 go build -o smartcopy main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o smartcopy main.go

# Build for all platforms
build-all: build-windows build-linux build-darwin
	@echo "All platform builds complete"

.PHONY: build clean clean-all run test build-windows build-linux build-darwin build-all
