#include "arena.h"
#include <stdlib.h>
#include <string.h>

/*
 * Align `n` up to the next multiple of `align`.
 * align must be a power of 2.
 */
static inline size_t align_up(size_t n, size_t align) {
  return (n + align - 1) & ~(align - 1);
}

/*
 * Allocate a new ArenaPage with at least `capacity` usable bytes.
 * The page struct and its data[] are allocated in a single malloc.
 */
static ArenaPage *page_new(size_t capacity) {
  ArenaPage *p = (ArenaPage *)malloc(sizeof(ArenaPage) + capacity);
  if (!p)
    return NULL;
  p->prev = NULL;
  p->capacity = capacity;
  p->used = 0;
  return p;
}

Arena *arena_new(size_t page_size) {
  Arena *a = (Arena *)malloc(sizeof(Arena));
  if (!a)
    return NULL;

  a->default_page_size = page_size < 64 ? 64 : page_size;
  a->current = page_new(a->default_page_size);
  if (!a->current) {
    free(a);
    return NULL;
  }
  return a;
}

void *arena_alloc(Arena *arena, size_t size, size_t align) {
  ArenaPage *page = arena->current;

  /*
   * Compute aligned pointer relative to the actual memory address,
   * not just the offset. This ensures correct alignment even when
   * data[] base isn't aligned to the requested boundary.
   */
  uintptr_t base = (uintptr_t)(page->data + page->used);
  uintptr_t aligned = (base + align - 1) & ~(align - 1);
  size_t padding = aligned - (uintptr_t)(page->data);
  /* padding is the required offset from data[] start */

  /* Fast path: fits in current page */
  if (padding + size <= page->capacity) {
    void *ptr = (void *)aligned;
    page->used = padding + size;
    return ptr;
  }

  /*
   * Slow path: need a new page.
   * Large objects (> half the default page size) get their own
   * dedicated page to avoid wasting space.
   */
  size_t new_capacity = arena->default_page_size;
  if (size > new_capacity / 2) {
    new_capacity = align_up(size, align);
  }

  ArenaPage *new_page = page_new(new_capacity);
  if (!new_page)
    return NULL;

  new_page->prev = arena->current;
  arena->current = new_page;

  /* First allocation in fresh page — align from data[] base */
  uintptr_t new_base = (uintptr_t)(new_page->data);
  uintptr_t new_aligned = (new_base + align - 1) & ~(align - 1);
  size_t new_padding = new_aligned - new_base;
  void *ptr = (void *)new_aligned;
  new_page->used = new_padding + size;
  return ptr;
}

ArenaCheckpoint arena_save(Arena *arena) {
  ArenaCheckpoint cp;
  cp.page = arena->current;
  cp.used = arena->current->used;
  return cp;
}

void arena_restore(Arena *arena, ArenaCheckpoint checkpoint) {
  /* Free all pages allocated after the checkpoint page */
  while (arena->current != checkpoint.page) {
    ArenaPage *prev = arena->current->prev;
    free(arena->current);
    arena->current = prev;
  }
  /* Reset the checkpoint page's used offset */
  arena->current->used = checkpoint.used;
}

void arena_destroy(Arena *arena) {
  ArenaPage *page = arena->current;
  while (page) {
    ArenaPage *prev = page->prev;
    free(page);
    page = prev;
  }
  free(arena);
}
