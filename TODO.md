# OrgLang TODO

Summary of pending issues and feature enhancements discovered during sanity test refinements.

## Refinement Required
- [ ] **Division equality**: `(6 / 3) = 2` fails because division returns a `Decimal` (`2.0`), which fails strict equality against Integer `2`. Need to implement type coercion for comparison operators or unify numeric types.
- [ ] **Lazy Iterator Indexing**: Operations like `([1 2 3] -> { right + 1 }).0` fail because the pipeline returns a `MapIterator` (lazy), not a Table.
    - [ ] Option A: Support index access directly on Iterators by driving them until the requested index.
    - [ ] Option B: Implement an "eager collection" operator (e.g., `!`) to convert Iterators to Tables (e.g., `(list -> map)! . 0`).
- [ ] **Operator Orthogonality Review**: Review other non-short-circuit operators (`&`, `|`, `^`) to distinguish between bitwise and logical semantics, similar to the `!` vs `~` separation.
- [ ] **Extended Assignment Operators**: Implement `:+`, `:-`, `:*`, `:/` etc., in the parser and runtime.
- [ ] **Standard Library Expansion**: Add more built-in resources for file I/O, networking, and string manipulation.

## Technical Debt
- [ ] Refactor `sanity_test.go` to handle large outputs more gracefully and avoid potential deadlocks in its own pipe-to-stdout logic.
- [ ] Improved error reporting in the C runtime (more descriptive signals than just "FAIL").
