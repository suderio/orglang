# Design: Args Resource

## Goal

Implement the `@args` resource to allow OrgLang programs to interact with command-line arguments and environment variables, as described in the specification.

## Specification (from README)

- `@args` is a built-in resource.
- It pulls data from command line arguments and environment variables.
- It sends a **Table** containing these values.
- Usage example: `@args -> ...`

## Implementation Details

### 1. C Runtime (`orglang.h`)

We need to capture `argc` and `argv` from the C `main` function and make them available to the OrgLang runtime.

**Changes:**

- Update `main` entry point to `int main(int argc, char **argv)`.
- Store `argc` and `argv` globally or pass them to a setup function.
- Define `org_resource_main_setup`, `org_resource_main_next`, etc.

**Data Structure:**
The `next` function of the `@main` resource should return a **Table** (OrgList) containing:

- Positional elements: Command line arguments (`argv[0]`...`argv[argc-1]`).
- Bindings (optional/future): Environment variables (e.g., `["PATH": "/bin", ...]`). For the first iteration, we can stick to arguments or just add env vars if easy (`environ` in C).

**Resource Lifecycle:**

- **Setup**: No-op or initialize internal state (boolean "emitted").
- **Next**:
  - If not emitted: Construct the Table from `argc`/`argv` (and `env`), mark as emitted, return the Table.
  - If emitted: Return `NULL` (End of Stream).
- **Teardown**: No-op.

### 2. Code Generation (`c_emitter.go`)

**Template Update:**
Change `int main() {` to `int main(int argc, char **argv) {`.

**Globals:**
Add a global `org_args_table` (or similar) or just access `argv` if we keep it in the simplified single-file transpilation model. Since `orglang.h` is included, we can declare a global `static int org_argc; static char **org_argv;` in `orglang_header.h` and populate it in `main` before executing the user code.

**Compiler Hook:**
In `c_emitter.go`, when encountering `@main`:

- The parser sees `@` (Prefix) and `main` (Identifier).
- `emitExpression` for `PrefixExpression` with operator `@`:
  - Currently checks for `sys`, `mem`, `org`.
  - Add case for `main`.
  - Emit `org_resource_main_create(arena)`.

### 3. Testing Strategy

**Sanity Tests:**

- Create `test/sanity/resource_main.org`.
- Content:

  ```rust
  # Expecting to run with some args
  @main -> {
    # Print first arg (program name)
    (("Arg 0: " + left.0) -> @stdout);
    # Print second arg
    (("Arg 1: " + left.1) -> @stdout);
  };
  ```

- Run this test with arguments using the `org run` command (need to ensure `org run` passes args to the binary).
  - `cmd/org/main.go` `runRun` might need update to forward extra args.

**Integration Tests:**

- Use the existing test runner (if any) or manual verification.

## Potential Problems

- `org run` argument forwarding: The current `cobra` command structure might consume arguments. We need to check `args` handling in `runCmd`.
- Environment Variables: C `environ` access requires declaration.

## Roadmap

1. **Modify `orglang_header.h`**: Add globals for usage, implement `org_resource_main_*`.
2. **Modify `c_emitter.go`**: Update template, implement `@main` emission.
3. **Update `main.go`**: Ensure `org run` forwards arguments (optional for this specific task but good for testing).
4. **Verify**: Run `sanity/resource_main.org`.
