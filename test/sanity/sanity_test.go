package sanity_test

import (
	"fmt"
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

	// 1.1 Create main.org content
	var mainContent strings.Builder
	var modules []string

	for _, file := range files {
		base := filepath.Base(file)
		if base == "main.org" {
			continue
		}

		modName := strings.TrimSuffix(base, ".org")
		modules = append(modules, modName)

		// <file_name> : "file_name.org" @ org
		mainContent.WriteString(fmt.Sprintf("%s : \"%s\" @ org;\n", modName, base))
		// Validate results:
		// Iterate over module results. If any result is falsy (false, Error, 0, null), print "FAIL".
		// We use Elvis ?:. If item is truthy, it returns item. If falsy, it returns "FAIL".
		// We then pipe to a check function.
		// Since we don't have convenient "if" yet, we flow to stdout if it matches FAIL.
		// Or simpler: item ?: "FAIL" -> @stdout;
		// But this prints "true" for passing tests (if item is true).
		// That's acceptable for verbose output, and we grep for FAIL.
		mainContent.WriteString(fmt.Sprintf("%s -> { right ?: \"FAIL %s\" -> @stdout } -> @stdout;\n", modName, modName))
	}

	mainPath := filepath.Join(wd, "main.org")
	if err := os.WriteFile(mainPath, []byte(mainContent.String()), 0644); err != nil {
		t.Fatalf("failed to write main.org: %v", err)
	}
	defer os.Remove(mainPath)

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
	if err != nil {
		t.Fatalf("failed to run main.org: %v\n%s", err, output)
	}

	// 2.3 Check expected result
	// "each imported Table must return a Table with only true values"
	// Since we can't easily parse the internal table state from here without the runtime printing it,
	// and the request says "compile" and "check", we assume implicit success if it runs
	// for now, UNLESS we expect formatting.
	// But logically, if the @org resource imported them, the 'run' command should have succeeded.
	// We'll trust the exit code 0 for success in this test harness.
	// The user explicitly expects FAILURE here because @org is not implemented.

	// 2.3 Check expected result
	outStr := string(output)
	t.Logf("Output: %s", output)

	if strings.Contains(outStr, "FAIL") {
		t.Fatalf("Test failed: output contains FAIL")
	}
}
