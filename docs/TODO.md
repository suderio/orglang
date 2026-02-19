# OrgLang TODO

Summary of pending issues and feature enhancements discovered during sanity test refinements.

## Implementation Gaps (Specification Sync)

- [ ] **Variable Capture (Closures)**: Implement lexical environment capture for operators (currently they are pure functions of inputs and globals).
- [ ] **Higher-Order Operators**: Implement `o` (Compose) and `|>` (Partial Application) in parser, codegen, and runtime.
- [ ] **Advanced Flow**: Implement `-<` (Balanced Dispatch) and `-<>` (Barrier Join) in the runtime.
- [ ] **Table Thunks and Eval**:
  - [ ] Implement actual lazy thunks for table elements.
  - [ ] Differentiate `.` (Table Access - returns thunk) from `?` (Selection/Eval - evaluates).
- [ ] **String Enhancements**:
- [ ] **Import Caching (Runtime)**: Implement a runtime cache for modules to prevent multiple executions of the same file (Singletons), aligning with the specification in the Build Model.
- [ ] **Resource Lifecycle**: Ensure full `setup`, `step`, and `teardown` coordination in the C runtime for all resource interactions.

- [ ] **Execution Model**:
  - [ ] Implement proper `main` entry point lookup and execution.
  - [ ] Support implicit table creation for the entire source file.
- [ ] **Resource Lifecycle**: Ensure full `setup`, `step`, and `teardown` coordination in the C runtime for all resource interactions.
- [ ] **Standard Library Expansion**:
  - [ ] Add more built-in resources for file I/O (`@file`), networking (`@net`), and string manipulation.
  - [ ] Implement string interpolation (`$N`, `$var`).
  - [ ] Ensure strings are semantically Tables indexed by integers.
- [ ] **Short-circuiting Tests**: Add test cases to verify `&&` and `||` short-circuiting (e.g., `false && (1/0)` should not error if short-circuiting works).
- [ ] **Error Flux**: Alternative path for errors in the flux.

## Future Roadmap (Wishlist)

- [ ] **Scientific Notation**: Add support for scientific notation (e.g., `1.2e10`) in decimal literals.
- [ ] **Standard Library Expansion**: Add more built-in resources for file I/O, networking, and string manipulation.
- [ ] **Scheduler: Async IO** — `@stdout.next` currently calls `write()` synchronously. Replace with IO queue submission + fiber yield.
- [ ] **Scheduler: Preemptive Yield** — Fibers currently run to completion. Add cooperative yield points and time-slice preemption.
- [ ] **Scheduler: `io_uring`/`epoll`** — Integrate kernel-level async IO for non-blocking resource operations.
- [ ] **Scheduler: Multi-Thread M:N** — Expand from single OS thread to one event loop per CPU core.
- [ ] **Static Analysis**: Implement a compiler pass for early error detection (undefined variables, type hints).
- [ ] **Pattern Matching**: Implement destructuring for table arguments in functions.
- [ ] **Coroutines**: Add first-class support for suspended execution contexts.
- [ ] **Tooling**:
  - [ ] **REPL**: Interactive environment for experimentation.
  - [ ] **LSP**: Language Server Protocol for IDE integration.
  - [ ] **Package Manager**: Dependency management tool (`org get`).
- [ ] **Optimizations**:
  - [ ] **Tail Call Optimization (TCO)**: For deep recursion safety.
  - [ ] **Bytecode Interpreter**: For faster development cycles.
  - [ ] **Machine Type Specialization**: Optimize arbitrary-precision numbers to `int64`/`float64` when possible.

## Technical Debt

- [ ] Improved error reporting in the C runtime (more descriptive signals than just "FAIL").
- [ ] Review mutated state in `,` operator (Persistence vs Mutation).
- [ ] **Documentation: EBNF grammar outdated** (README.md §Full Grammar). The EBNF does not cover: raw strings (`RAWSTRING`), escape sequences in `STRING`, Unicode identifiers, `\` and `'` as structural/delimiter characters.
