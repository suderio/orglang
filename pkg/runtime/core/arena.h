#ifndef ORG_ARENA_H
#define ORG_ARENA_H

#include <stddef.h>
#include <stdint.h>

/*
 * Arena Allocator - Chained-page bump-pointer allocator.
 *
 * Memory is allocated by bumping a pointer forward. Individual frees
 * are not supported; memory is reclaimed in bulk via checkpoints or
 * by destroying the entire arena.
 *
 * All allocations are 8-byte aligned (required for tagged pointers).
 */

/* A single page in the arena's linked list. */
typedef struct ArenaPage {
    struct ArenaPage *prev; /* Previous page (linked list) */
    size_t capacity;        /* Total usable bytes in data[] */
    size_t used;            /* Bytes allocated so far */
    uint8_t data[];         /* Flexible array member */
} ArenaPage;

/* The arena handle. */
typedef struct Arena {
    ArenaPage *current;        /* Active page */
    size_t default_page_size;  /* Default data[] capacity for new pages */
} Arena;

/* Checkpoint for save/restore (sub-scope reclamation). */
typedef struct ArenaCheckpoint {
    ArenaPage *page;  /* Page at time of save */
    size_t used;      /* used offset at time of save */
} ArenaCheckpoint;

/*
 * Create a new arena. page_size is the default capacity (in bytes)
 * for each page's data[] region. Typical values: 4096 or 65536.
 * Returns NULL on allocation failure.
 */
Arena *arena_new(size_t page_size);

/*
 * Allocate `size` bytes from the arena, aligned to `align` bytes.
 * align must be a power of 2 (typically 8).
 * Returns NULL only if the system is out of memory.
 */
void *arena_alloc(Arena *arena, size_t size, size_t align);

/*
 * Save the current arena position. Paired with arena_restore()
 * to reclaim all allocations made after the checkpoint.
 */
ArenaCheckpoint arena_save(Arena *arena);

/*
 * Restore the arena to a previously saved checkpoint.
 * All memory allocated after the checkpoint becomes invalid.
 * Pages allocated after the checkpoint are freed back to the OS.
 */
void arena_restore(Arena *arena, ArenaCheckpoint checkpoint);

/*
 * Destroy the arena and release all pages back to the OS.
 * The Arena pointer itself is also freed.
 */
void arena_destroy(Arena *arena);

#endif /* ORG_ARENA_H */
