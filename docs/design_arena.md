# Design: `@arena` Resource and Managed Flow Lifecycle

## Objective
Implement the `@arena` library resource to allow heap allocation in an Arena model during pull workflows. The setup and teardown of the arena must be entirely managed by the language, ensuring that memory management is tied to the lifecycle of the data flow.

## Problem Statement
Currently, resources in OrgLang (like `@stdout`) are manually instantiated or global buffers. The `->` operator handles data flow but does not automatically manage the lifecycle (`setup`/`teardown`) of resources passed into the flow.
For `@arena`, we need a mechanism where:
1.  An Arena is created when the flow starts.
2.  Allocations within the flow use this new Arena.
3.  The Arena is potentially freed when the flow ends.

## Proposed Solution

### 1. Resource Middleware Pattern
We introduce the concept of using a **Resource Definition** (not just an Instance) as a middleware in a flow chain.

Syntax:
```rust
source -> @arena -> sink
```
*(Note: Assuming `@arena` refers to the definition or a factory, consistent with "library resources").*

When the runtime encounters `Iterator -> ResourceDefinition` (or a specific "Context Resource"):
1. It creates a **Scoped Iterator**.
2. This iterator manages the lifecycle of the resource instance.

### 2. Runtime Implementation (`orglang_header.h`)

#### Arena Mechanics & Teardown
The Arena will be enhanced to track resources that require converting/cleanup (specifically OS handles) to ensure deterministic finalization.

```c
// Forward declaration
typedef struct OrgResourceInstance OrgResourceInstance;

typedef struct Arena {
  char *data;
  size_t size;
  size_t offset;
  // List of active resource instances allocated in this arena that have teardowns
  struct OrgResourceInstance *resources_head; 
} Arena;
```

**Allocation**:
- Standard values (Int, String) just bump `offset`.
- **Resource Instances** (created via `@`) are linked into `resources_head`.

**Teardown (`arena_free`)**:
1. Iterate `resources_head`.
2. For each instance, call its `def->teardown` (if present) with its `state`.
   - Example: `@file` teardown calls `@sys("close", fd)`.
3. Free/Unmap the `data` block.

#### New Type: `OrgScopedIterator`
A new iterator type that wraps an upstream iterator and manages the Arena context.

```c
typedef struct OrgScopedIterator {
    OrgIterator *upstream;
    OrgResource *def;     // The @arena definition
    OrgValue *context;    // The Arena handle (OrgValue wrapping Arena*)
} OrgScopedIterator;
```

**Lifecycle in `next()`**:
1.  **Context Switching**:
    - The `context` (Arena) is passed to `upstream->next(arena_ptr, ...)`.
    - All allocations in `upstream` happen in this Arena.
2.  **End of Stream / Error**:
    - `arena_free(arena_ptr)` is called.
    - This triggers the "List Walk" described above, closing all FDs.
    - The Arena memory is released.

### 3. The `@arena` Resource Definition
The `@arena` resource acts as a factory for these scopes.

```rust
Arena : resource [
    setup: { "mem" @ sys }    # Returns a new Arena
    # Teardown is handled implicitly by the ScopedIterator calling arena_free?
    # Or proper: 
    teardown: { left @ "free_arena" } 
];
```

*Refinement*: To keep it "managed by language", the `ScopedIterator` created by `-> @arena` logic will be essentially hardcoded to manage an `Arena` type context, or `@arena` uses a native setup that returns a special "Arena/Context" value.

### 4. Integration with `->` Operator
Modify `org_op_infix` in `orglang_header.h`:
Checks if `right` is the `@arena` resource (or any Resource Definition).
- Creates `OrgScopedIterator`.
- `OrgScopedIterator`'s `next` function handles the `arena_create` / call upstream / `arena_free` cycle.

### 5. Safety and Semantics
- **Deterministic Finalization**: Scope end = Arena Free = Handles Closed.
- **Concurrency**: Each flow branch (`-> @arena`) gets its own isolated Arena and Handle List.
- **No GC**: Memory is reclaimed in bulk. Handles are closed precisely.

## Tasks
1.  Add `ORG_SCOPED_ITERATOR_TYPE` to `OrgType` enum in `orglang_header.h`.
2.  Implement `org_scoped_iterator_create` and its `next` function.
3.  Update `org_op_infix` to handle `Iterator -> ResourceDefinition`.
4.  Implement the native `setup`/`teardown` logic for Arena (may require `arena_create_sub` or similar).

