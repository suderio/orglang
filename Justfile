# Clean build artifacts and test outputs
clean:
    rm -rf dist/
    rm -f org
    rm -f test/sanity/main test/sanity/*.c
    go clean

# Run all tests (after cleaning)
test: clean
    go test ./...

# Build snapshot release
snapshot: clean
    goreleaser release --snapshot --clean

# Build full release
release: clean
    goreleaser release --clean

# Build the binary locally
build:
    go build -o org ./cmd/org
