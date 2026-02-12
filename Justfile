# Clean build artifacts and test outputs
clean:
	rm -rf dist/
	rm -f org
	rm -f test/sanity/*.c test/sanity/orglang.h
	rm -f test/integration/testdata/*.c test/integration/testdata/orglang.h
	find test/sanity test/integration/testdata -type f ! -name "*.*" -executable -delete 2>/dev/null || true
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
