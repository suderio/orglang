# Hybrid Scheduler Design

The **OrgLang Hybrid Scheduler** is designed to treat every pulse of data as a task that can be suspended and resumed based on hardware availability. In our model, `@args` is simply the "Seed Pulse" that starts the chain reaction.

## 1. The Global Topology: The Registry

When the Go compiler processes an OrgLang file, it generates a `main()` function that acts as a **Registration Phase**.

Every flow starting with `@args` is effectively registered as a "Root Flow" by executing the expression that spawns the initial fibers.

```c
// Pseudo-generated C code from the Go compiler
// This is just an example, the actual implementation will be different.
void org_init_program(OrgContext* ctx) {
    // Flow 1: @args -> "Hello" -> @stdout
    // This executes effectively as:
    // org_sched_spawn_pump(ctx, main_iter, func_hello_stdout);
    
    // Flow 2: @args -> "World" -> @stderr
    // org_sched_spawn_pump(ctx, main_iter, func_world_stderr);
}
```

## 2. The `io_uring` / IOCP Execution Model (Simulated for Prototype)

To achieve non-blocking parallelism, we use a **Completion-Based** loop.

### The "Pulse" as a Completion Event

In OrgLang, the execution follows these stages:

1. **Submission:** The scheduler pulses `@args`. This pulse travels through the graph. When it hits `@stdout`, the resource doesn't "write" and wait. It submits a `SQE` (Submission Queue Entry) to the kernel/scheduler and **yields** its current execution context (the Fiber).
2. **The Event Loop:** The runtime enters a `wait` state.
3. **Completion:** The kernel/scheduler finishes the simulated IO. It places a `CQE` (Completion Queue Entry) in the ring.
4. **Resumption:** The scheduler sees the `CQE`, looks up the Fiber associated with that specific request, and **resumes** it.

## 3. The Hybrid Part: Fibers

* **M (Fibers):** Every logic branch (the code inside `{ }`) is a Fiber. They are extremely cheap.
* **N (OS Threads):** We run one Event Loop per CPU Core (Simulated as 1 for now).

**How `@args` works in this context:**
When `@args` pulses, it creates separate Fibers. Each Fiber uses its own **Sub-Arena** (or shares one carefully), allowing for isolation.

## 4. Implementation Details

### Data Structures

```c
typedef struct OrgFiber {
    int id;
    OrgContext *ctx;
    Arena *arena; // Sub-arena
    void (*resume)(struct OrgFiber *f, OrgValue *val); // Continuation
    OrgValue *state; // Captured state/closure
    struct OrgFiber *next; // Queue link
} OrgFiber;

typedef struct OrgScheduler {
    OrgFiber *ready_head;
    OrgFiber *ready_tail;
    // IO Queues would go here
} OrgScheduler;

typedef struct OrgContext {
    Arena *global_arena;
    OrgScheduler scheduler;
} OrgContext;
```

### Scheduler Loop

```c
// This is just pseudocode to illustrate the idea.
// The actual implementation will be different.
void org_scheduler_run(OrgContext *ctx) {
    while (ctx->scheduler.ready_head) {
        // Pop Fiber
        OrgFiber *f = ctx->scheduler.ready_head;
        ctx->scheduler.ready_head = f->next;
        if (!ctx->scheduler.ready_head) ctx->scheduler.ready_tail = NULL;

        // Run Fiber
        // The resume function executes the next step.
        // It might spawn new fibers or just finish.
        if (f->resume) {
            f->resume(f, NULL);
        }
        
        // Check for IO completions (Simulate polling)
        org_io_poll(ctx);
    }
}
```

### Operator `->` Refactoring

The `->` operator currently executes recursively. It must be changed to:

1. **If Left is Iterator**: Spawn a `Pump` fiber that iterates `Left` and spawns `Task` fibers for `Right` for each item.
    * This makes `->` returns immediately (registering the pump).
2. **If Left is Value**: Spawn a `Task` fiber that executes `Right(Left)`.

### IO Simulation

For `@stdout`, instead of `printf`:

1. Submit "Write" request to Scheduler IO Queue.
2. Yield Fiber (Return).
3. Scheduler processes IO Queue (prints to real stdout).
4. mark Fiber as ready (if it had a continuation).

For the prototype, `@stdout` can just print and return (synchronous is a special case of async that completes immediately).

## 5. Potential Problems

1. **Memory Management**: Fibers need their own Arenas or careful management of the global Arena to avoid fragmentation if they have different lifecycles.
2. **Stack Variables**: C functions relies on stack. If we "yield", we lose stack.
    * **Solution**: We must use CPS (Continuation Passing Style) where `state` is heap-allocated in the Arena.
    * The `OrgFiber` struct holds the `state`.
    * Every "Step" function must take `OrgFiber*` and cast `state` to its specific struct.
