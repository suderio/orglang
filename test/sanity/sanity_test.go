package sanity_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanity(t *testing.T) {
	// 1. Resolve compiler path
	compilerPath, err := filepath.Abs("../../org")
	if err != nil {
		t.Fatalf("failed to resolve compiler absolute path: %v", err)
	}

	// Build the compiler if it doesn't exist
	if _, err := os.Stat(compilerPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", compilerPath, "../../cmd/org/main.go")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed to build compiler: %v\n%s", err, output)
		}
	}

	// 2. Read all .org files in test/sanity/*.org
	wd, _ := os.Getwd()
	files, err := filepath.Glob(filepath.Join(wd, "*.org"))
	if err != nil {
		t.Fatalf("failed to glob files: %v", err)
	}

	// Verify main.org exists
	mainPath := filepath.Join(wd, "main.org")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Fatalf("main.org not found in sanity directory")
	}

	// 2. For each file (sanity check individually)
	// The request says "2. For each file: 2.1 compiles the file"
	// and "2.2 compiles main.org"
	// So we should verify individual compilation too.

	for _, file := range files {
		if filepath.Base(file) == "main.org" {
			continue
		}

		t.Logf("Compiling %s...", filepath.Base(file))
		cmd := exec.Command(compilerPath, "build", file)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("failed to compile %s: %v\n%s", filepath.Base(file), err, output)
		} else {
			t.Logf("Compiled %s successfully", filepath.Base(file))
			// Cleanup artifact
			base := filepath.Base(file)
			name := strings.TrimSuffix(base, ".org")
			os.Remove(name)
			os.Remove(name + ".c")
		}
	}

	// 2.2 Compile main.org
	t.Log("Compiling main.org...")
	cmd := exec.Command(compilerPath, "run", mainPath)
	output, err := cmd.CombinedOutput()

	// Cleanup main artifacts regardless of success
	os.Remove("main")
	os.Remove("main.c")

	if err != nil {
		t.Fatalf("failed to run main.org: %v\n%s", err, output)
	}

	// 2.3 Check expected result
	outStr := string(output)
	t.Logf("Output: %s", output)

	if strings.Contains(outStr, "FAIL") {
		t.Fatalf("Test failed: output contains FAIL")
	}
}
