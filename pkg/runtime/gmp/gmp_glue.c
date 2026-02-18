#include "gmp_glue.h"
#include <gmp.h>
#include <string.h>

/*
 * Thread-local arena pointer. The scheduler sets this every time
 * it resumes a fiber, so all GMP operations allocate from the
 * correct per-fiber arena.
 */
static __thread Arena *current_fiber_arena = NULL;

void org_gmp_set_arena(Arena *arena) { current_fiber_arena = arena; }

Arena *org_gmp_get_arena(void) { return current_fiber_arena; }

/*
 * GMP custom allocator: allocate from the thread-local arena.
 */
static void *org_gmp_alloc(size_t size) {
  return arena_alloc(current_fiber_arena, size, 8);
}

/*
 * GMP custom free: no-op. Arena reclaims memory in bulk.
 */
static void org_gmp_free(void *ptr, size_t size) {
  (void)ptr;
  (void)size;
  /* No-op: Arena reclaims in bulk via arena_restore/arena_destroy */
}

/*
 * GMP custom realloc: try to extend in-place if the pointer is
 * the last allocation on the current page. Otherwise allocate new
 * space and copy.
 */
static void *org_gmp_realloc(void *ptr, size_t old_size, size_t new_size) {
  if (new_size <= old_size) {
    return ptr; /* Shrinking: keep same pointer */
  }

  /* Try in-place extension: if ptr is at the end of the current page */
  ArenaPage *page = current_fiber_arena->current;
  uint8_t *end_of_alloc = (uint8_t *)ptr + old_size;
  uint8_t *end_of_page = page->data + page->used;

  if (end_of_alloc == end_of_page) {
    size_t extra = new_size - old_size;
    if (page->used + extra <= page->capacity) {
      page->used += extra;
      return ptr; /* Extended in-place */
    }
  }

  /* Cannot extend in-place: allocate new, copy, abandon old */
  void *new_ptr = arena_alloc(current_fiber_arena, new_size, 8);
  if (!new_ptr)
    return NULL;
  memcpy(new_ptr, ptr, old_size);
  return new_ptr;
}

void org_gmp_init(void) {
  mp_set_memory_functions(org_gmp_alloc, org_gmp_realloc, org_gmp_free);
}
