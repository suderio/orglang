# CLI Plan

This document outlines the command-line interface for the OrgLang compiler (`org`).

## Core Principles

- **POSIX Compliance**: Standard flag parsing (`-f`, `--flag`).
- **Subcommand Structure**: similar to `go`, `cargo`, `git`.
- **Default Action**: `org <file.org>` behaves like `org run <file.org>`.

## Dependencies

- **Library**: `github.com/spf13/cobra` (Recommended standard for Go CLIs).
- **Fallback**: Standard `flag` package (if external deps strictly avoided, but Cobra preferred for robustness).

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

## Implementation Phase 1

Focus on skeleton and `version`/`help`.

1. Initialize Cobra structure.
2. Implement `version` command.
3. Stub `build`, `run`, `test`, `repl` with "TBD" output.
