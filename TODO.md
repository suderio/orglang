# OrgLang TODO

Summary of pending issues and feature enhancements discovered during sanity test refinements.

## Refinement Required
- [ ] **Arbitrary Precision Numerics**: Replace current `long long` and `double` implementations with arbitrary-precision types (BigInt and BigDecimal) to support the language's design goals.
- [ ] **Rationals and Division**: Implement actual `Rational` type support. Update division logic: `Integer / Integer` returns an `Integer` if exact, otherwise a `Rational` (e.g., `3 / 2 = 3/2`).
- [ ] **Division equality**: `(6 / 3) = 2` currently fails due to type mismatch (Decimal vs Integer). This will be resolved by the new numeric unification/division logic.
- [ ] **Lazy Iterator Indexing**: Operations like `([1 2 3] -> { right + 1 }).0` fail because the pipeline returns a `MapIterator` (lazy), not a Table.
    - [ ] Option A: Support index access directly on Iterators by driving them until the requested index.
    - [ ] Option B: Implement an "eager collection" operator (e.g., `!`) to convert Iterators to Tables (e.g., `(list -> map)! . 0`).
- [ ] **Operator Orthogonality Review**: Review other non-short-circuit operators (`&`, `|`, `^`) to distinguish between bitwise and logical semantics, similar to the `!` vs `~` separation.
- [ ] **Extended Assignment Operators**: Implement `:+`, `:-`, `:*`, `:/` etc., in the parser and runtime.
- [ ] **Standard Library Expansion**: Add more built-in resources for file I/O, networking, and string manipulation.

## Implementation Gaps (Specification Sync)

- [ ] **Higher-Order Operators**: Implement `o` (Compose) and `|>` (Partial Application) in parser, codegen, and runtime.
- [ ] **Short-Circuiting**: Logical `&&` and `||` must not evaluate the right operand if the result is determined by the left.
- [ ] **Advanced Flow**: Implement `-<` (Balanced Dispatch) and `-<>` (Barrier Join) in the runtime.
- [ ] **Table Thunks and Eval**:
    - [ ] Implement actual lazy thunks for table elements.
    - [ ] Differentiate `.` (Table Access - returns thunk) from `?` (Selection/Eval - evaluates).
- [ ] **Comparison Chaining**: Refactor to support `x < y < z` returning the last comparison result as per spec.
- [ ] **String Enhancements**:
    - [ ] Implement `$N` (positional) and `$var` (variable) interpolation.
    - [ ] Ensure strings are semantically Tables indexed by integers.
- [ ] **Execution Model**:
    - [ ] Implement proper `main` entry point lookup and execution.
    - [ ] Support implicit table creation for the entire source file.
- [ ] **Resource Lifecycle**: Ensure full `setup`, `step`, and `teardown` coordination in the C runtime for all resource interactions.
- [ ] **Atoms in Tables Tests**: Add test cases to verify greedy binding and space-separation logic in table literals (e.g., `[a: 1 + 1]` vs `[a: (1 + 1)]`).

## Technical Debt
- [ ] Refactor `sanity_test.go` to handle large outputs more gracefully and avoid potential deadlocks in its own pipe-to-stdout logic.
- [ ] Improved error reporting in the C runtime (more descriptive signals than just "FAIL").
- [ ] Review mutated state in `,` operator (Persistence vs Mutation).
