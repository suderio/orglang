package integration

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update .golden files")

func TestIntegration(t *testing.T) {
	// Locate the compiler binary. Assumes it is built in the project root.
	// For robustness, we could build it to a temp dir here.
	compilerPath, err := filepath.Abs("../../org")
	if err != nil {
		t.Fatalf("failed to resolve compiler path: %v", err)
	}

	// Check if binary exists, attempt build if not
	if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
		t.Log("Compiler binary not found, attempting to build...")
		buildCmd := exec.Command("go", "build", "-o", compilerPath, "../../cmd/org/main.go")
		if out, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to build compiler: %v\n%s", err, out)
		}
	}

	files, err := filepath.Glob("testdata/*.org")
	if err != nil {
		t.Fatalf("failed to glob testdata: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			runTest(t, compilerPath, file)
		})
	}
}

func runTest(t *testing.T, compiler string, sourcePath string) {
	// Assume sourcePath is relative to test execution dir?
	// `go test ./test/integration` runs in `test/integration`.
	// sourcePath from glob is `testdata/file.org`.
	absSource, _ := filepath.Abs(sourcePath)

	// Command: ./org run <file>
	cmd := exec.Command(compiler, "run", absSource)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("compiler execution failed: %v\nStderr: %s", err, stderr.String())
	}

	actual := stdout.Bytes()
	goldenPath := sourcePath[:len(sourcePath)-len(".org")] + ".golden"

	if *update {
		if err := ioutil.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
	}

	expected, err := ioutil.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Normalize line endings if needed (optional)
	if !bytes.Equal(actual, expected) {
		t.Errorf("output mismatch for %s:\nExpected:\n%s\nActual:\n%s",
			sourcePath, string(expected), string(actual))
	}
}
