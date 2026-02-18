# Clean build artifacts and test outputs
clean:
	rm -rf dist/
	rm -rf build/
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

# Build Reference PDF (requires pandoc and wkhtmltopdf/weasyprint)
pdf:
    @echo "🎨 Renderizando PDF com WeasyPrint e CSS..."
    pandoc README.md -o docs/Reference.pdf --css assets/style.css --pdf-engine=weasyprint --embed-resources --standalone --toc --metadata title="OrgLang Reference"
    @echo "✨ PDF gerado"

view: pdf
    xdg-open "docs/Reference.pdf"

# Generate test coverage report
coverage:
    go test -coverprofile=coverage.out -covermode=count ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "📊 Coverage report: coverage.html"

# Run C runtime unit tests
test-c:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "🔨 Building C runtime tests..."
    mkdir -p build/test
    for src in tests/runtime/test_*.c; do
        name=$(basename "$src" .c)
        clang -Wall -Wextra -g -Ipkg/runtime -o "build/test/$name" "$src" \
            pkg/runtime/core/*.c pkg/runtime/gmp/*.c pkg/runtime/ops/*.c \
            pkg/runtime/table/*.c pkg/runtime/closure/*.c \
            pkg/runtime/resource/*.c pkg/runtime/sched/*.c \
            -lgmp 2>/dev/null || clang -Wall -Wextra -g -Ipkg/runtime -o "build/test/$name" "$src" \
            $(find pkg/runtime -name '*.c' 2>/dev/null | head -20) -lgmp 2>/dev/null || \
            echo "⚠️  Skipping $name (missing sources)"
        if [ -f "build/test/$name" ]; then
            echo "  ✅ $name"
            "./build/test/$name"
        fi
    done

# Generate C runtime coverage report
coverage-c:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "📊 Building C tests with coverage instrumentation..."
    mkdir -p build/coverage
    for src in tests/runtime/test_*.c; do
        name=$(basename "$src" .c)
        clang -fprofile-instr-generate -fcoverage-mapping -Wall -g -Ipkg/runtime \
            -o "build/coverage/$name" "$src" \
            $(find pkg/runtime -name '*.c') -lgmp 2>/dev/null || continue
        echo "  Running $name..."
        LLVM_PROFILE_FILE="build/coverage/$name.profraw" "./build/coverage/$name"
    done
    llvm-profdata merge build/coverage/*.profraw -o build/coverage/merged.profdata
    llvm-cov report build/coverage/test_* -instr-profile=build/coverage/merged.profdata
    llvm-cov show build/coverage/test_* -instr-profile=build/coverage/merged.profdata \
        --format=html -output-dir=build/coverage/html
    @echo "📊 C coverage report: build/coverage/html/index.html"

# Run all tests (Go + C)
test-all: test test-c

