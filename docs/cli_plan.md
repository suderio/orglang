# CLI Plan

This document outlines the command-line interface for the OrgLang compiler (`org`).

## Core Principles

- **POSIX Compliance**: Standard flag parsing (`-f`, `--flag`).
- **Subcommand Structure**: similar to `go`, `cargo`, `git`.
- **Default Action**: `org <file.org>` behaves like `org run <file.org>`.

## Design Philosophy

### **"Distinct, yet Sober"**

The CLI should feel modern and professional, avoiding excessive emojis or neon colors, but using typography and subtle colors to guide the eye.

- **Palette**: Monochromatic base with single accent color (e.g., specific shade of blue or purple) for headers/success.
- **Typography**: Clean indentation, bold for emphasis, dim for secondary info.
- **Libraries**:
  - `github.com/spf13/cobra`: Command structure.
  - `github.com/charmbracelet/lipgloss`: Styling and layout.
  - `github.com/muesli/termenv`: Color profile detection.

## Dependencies

- **Framework**: `github.com/spf13/cobra`
- **Styling**: `github.com/charmbracelet/lipgloss`

## Commands

### `build`

Compiles OrgLang source code into an executable or bytecode.

**Usage**: `org build [flags] <input>`

**Flags**:

- `-o, --output <file>`: Output file name (default: input file name without extension).
- `-t, --target <arch>`: Target architecture (e.g., `linux/amd64`, `wasm`). *Future*.
- `-O, --optimize <level>`: Optimization level (`0`, `1`, `2`, `3`). Default `1`.
- `--static`: Link statically (for C output).
- `--debug`: Include debug information.
- `-v, --verbose`: Verbose output during compilation.

**Status**: TBD (Stub implementation)

### `run`

Compiles and executes an OrgLang program immediately.

**Usage**: `org run [flags] <input> [args...]`

**Flags**:

- `-a, --args <args>`: Pass arguments to the program (alternative to `[args...]`).
- `--debug`: Run in debug mode (e.g., debugger attached).

**Status**: TBD (Stub implementation)

### `repl`

Starts an interactive Read-Eval-Print Loop.

**Usage**: `org repl [flags]`

**Flags**:

- `--banner/--no-banner`: Show/Hide welcome message.
- `--history <file>`: Path to history file.

**Status**: TBD (Stub implementation)

### `test`

Runs tests defined in OrgLang files (standard testing framework).

**Usage**: `org test [flags] [files...]`

**Flags**:

- `-v, --verbose`: Verbose test output.
- `--filter <regex>`: Run only tests matching regex.
- `--coverage`: Generate coverage report.

**Status**: TBD (Stub implementation)

### `version`

Prints the current version of the OrgLang compiler.

**Usage**: `org version [flags]`

**Flags**:

- `--short`: Print only the version number.
- `--json`: Print version info in JSON format.

**Status**: **To Be Implemented Now**

### `help`

Help about any command.

**Usage**: `org help [command]`

**Status**: **To Be Implemented Now** (Cobra provides this automatically).

## Recommended Additional Commands

### `check` / `vet`

Performs static analysis without Compiling/Running. Useful for CI/CD and editor integration.

**Usage**: `org check <input>`

**Status**: TBD

### `fmt`

Formats OrgLang source files to standard style.

**Usage**: `org fmt [flags] <files...>`

**Flags**:

- `-w, --write`: Write result to file instead of stdout.
- `--check`: checks if file is formatted (exit code 1 if not).

**Status**: TBD (Critical for modern languages)

### `doc`

Generates documentation from docstrings.

**Usage**: `org doc [flags] <input>`

**Flags**:

- `--html`: Output HTML.
- `--json`: Output JSON.

**Status**: TBD

### `clean`

Removes build artifacts.

**Usage**: `org clean`

**Status**: TBD

### Versioning Strategy

To synchronize the CLI version with Git tags (like GoReleaser), we have two main approaches:

#### 1. LDFLAGS Injection (Recommended Standard)

This is the standard Go way. Variables in `pkg/cmd/version.go` are overwritten at build time.

- **Source of Truth**: Git Tags (`v1.0.0`).
- **Mechanism**: `go build -ldflags "-X orglang/pkg/cmd.Version=$(git describe --tags)"`
- **GoReleaser Config**:

  ```yaml
  builds:
    - ldflags:
      - -X orglang/pkg/cmd.Version={{.Version}}
      - -X orglang/pkg/cmd.Commit={{.Commit}}
      - -X orglang/pkg/cmd.BuildDate={{.Date}}
  ```

#### 2. Embedded Version File (Alternative)

If a file-based source of truth is preferred (e.g. `VERSION` file in root).

- **Source of Truth**: `VERSION` file.
- **Mechanism**:
  - Create `VERSION` file containing `1.0.0`.
  - In `pkg/cmd/version.go`:

    ```go
    //go:embed ../../VERSION
    var versionFile string
    var Version = strings.TrimSpace(versionFile)
    ```

- **Pros**: Easy to read programmatically without build flags.
- **Cons**: Requires manual bumping or extra tooling to sync with Git tags.

**Recommendation**: Use **LDFLAGS** as it guarantees the binary version matches the git tag/commit exactly without manual file management.

**Selected Strategy**: Option 1 (LDFLAGS Injection).

## Implementation Phase 1

Focus on skeleton and `version`/`help`.

1. Initialize Cobra structure.
2. Implement `version` command.
3. Stub `build`, `run`, `test`, `repl` with "TBD" output.
