# Walkthrough - Arena Resource Implementation

I have implemented the `@arena` resource mechanics and the `ScopedIterator` middleware pattern in the OrgLang C runtime. This allows for language-managed heap allocation scopes with deterministic finalization of resources.

## Changes

### Runtime (`orglang_header.h`)

1.  **Arena Resource Tracking**:
    - Added `resources_head` to `Arena` struct to track a linked list of active `OrgResourceInstance`s.
    - Implemented `arena_resource_register` to add instances to the current arena's list upon creation (`@` operator).

2.  **Deterministic Teardown**:
    - Implemented `arena_free`. It now iterates the `resources_head` list and calls the `teardown` method for each resource before freeing the memory block.
    - This ensures that OS handles (e.g., file descriptors) opened within an arena scope are closed even if they leak contextually.

3.  **Scoped Iterator (Middleware)**:
    - Added `ORG_SCOPED_ITERATOR_TYPE`.
    - Implemented `org_scoped_iterator_create` and `scoped_iterator_next`.
    - Modified `->` operator (`org_op_infix`) to detect if the right operand is a **Resource Definition**.
    - If so, it creates a `ScopedIterator` which:
        - Calls `setup` on the resource to get a context (e.g., Arena Pointer).
        - Switches the active `Arena *` passed to the upstream iterator.
        - Calls `teardown` when the stream ends.

4.  **Syscalls**:
    - Added `arena_create` and `arena_release` to `@sys` to support the implementation of the `@arena` resource in OrgLang.

## Verification

I created a test file `test/feature/arena_test.org` that mocks an Arena using the new syscalls and verifies two scenarios:

### 1. Standard Middleware Flow
```rust
[1] -> Tracked -> Arena -> @stdout;
```
- **Result**: `Tracked` is setup inside `Arena` scope. It tears itself down when the flow ends naturally.

### 2. Leaked Resource Cleanup
```rust
[1] -> { val: @Tracked; right } -> Arena -> @stdout;
```
- **Result**: `@Tracked` is created but "leaked" (not returned to the flow). When `Arena` scope ends, `arena_free` is triggered, which successfully finds the leaked `@Tracked` instance and calls its `teardown`.

### Output Log
```
--- START TEST 2 (Leak Cleanup) ---
[ARENA SETUP] Created Arena: 94175254172480
[TRACKED SETUP]
1[ARENA TEARDOWN] Freeing Arena: 94175254172480
[TRACKED TEARDOWN]
```
The "TRACKED TEARDOWN" appearing after "Freeing Arena" confirms that the Arena correctly cleaned up its resources.
