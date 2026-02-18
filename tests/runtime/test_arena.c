/*
 * test_arena.c — Unit tests for the Arena allocator.
 *
 * Compile:
 *   clang -Wall -Wextra -g -o test_arena \
 *       test_arena.c ../../pkg/runtime/core/arena.c
 */
#include "../../pkg/runtime/core/arena.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name)                                                             \
  do {                                                                         \
    tests_run++;                                                               \
    printf("  %-50s", name);                                                   \
  } while (0)

#define PASS()                                                                 \
  do {                                                                         \
    tests_passed++;                                                            \
    printf("✅\n");                                                            \
  } while (0)

#define ASSERT(cond)                                                           \
  do {                                                                         \
    if (!(cond)) {                                                             \
      printf("❌ FAIL: %s (line %d)\n", #cond, __LINE__);                      \
      return;                                                                  \
    }                                                                          \
  } while (0)

static void test_arena_new_destroy(void) {
  TEST("arena_new / arena_destroy");
  Arena *a = arena_new(4096);
  ASSERT(a != NULL);
  ASSERT(a->current != NULL);
  ASSERT(a->default_page_size == 4096);
  arena_destroy(a);
  PASS();
}

static void test_arena_min_page_size(void) {
  TEST("arena_new clamps small page size to 64");
  Arena *a = arena_new(8);
  ASSERT(a != NULL);
  ASSERT(a->default_page_size == 64);
  arena_destroy(a);
  PASS();
}

static void test_arena_alloc_basic(void) {
  TEST("arena_alloc basic allocation");
  Arena *a = arena_new(4096);
  void *p1 = arena_alloc(a, 16, 8);
  ASSERT(p1 != NULL);
  ASSERT(((uintptr_t)p1 & 7) == 0); /* 8-byte aligned */

  void *p2 = arena_alloc(a, 32, 8);
  ASSERT(p2 != NULL);
  ASSERT(p2 > p1); /* Sequential in same page */
  ASSERT(((uintptr_t)p2 & 7) == 0);

  arena_destroy(a);
  PASS();
}

static void test_arena_alloc_alignment(void) {
  TEST("arena_alloc respects alignment");
  Arena *a = arena_new(4096);

  /* Allocate 1 byte, then request 16-byte alignment */
  arena_alloc(a, 1, 8);
  void *p = arena_alloc(a, 16, 16);
  ASSERT(p != NULL);
  ASSERT(((uintptr_t)p & 15) == 0); /* 16-byte aligned */

  arena_destroy(a);
  PASS();
}

static void test_arena_alloc_overflow_to_new_page(void) {
  TEST("arena_alloc overflows to new page");
  Arena *a = arena_new(64);

  /* Fill first page */
  void *p1 = arena_alloc(a, 64, 8);
  ASSERT(p1 != NULL);
  ArenaPage *first_page = a->current;

  /* Next allocation must go to a new page */
  void *p2 = arena_alloc(a, 16, 8);
  ASSERT(p2 != NULL);
  ASSERT(a->current != first_page); /* New page was allocated */

  arena_destroy(a);
  PASS();
}

static void test_arena_alloc_large_object(void) {
  TEST("arena_alloc large object gets dedicated page");
  Arena *a = arena_new(64);

  /* Request something bigger than half the page size */
  void *p = arena_alloc(a, 128, 8);
  ASSERT(p != NULL);
  /* The page should have capacity >= 128 */
  ASSERT(a->current->capacity >= 128);

  arena_destroy(a);
  PASS();
}

static void test_arena_alloc_write_read(void) {
  TEST("arena_alloc memory is usable (write/read)");
  Arena *a = arena_new(4096);

  char *s = (char *)arena_alloc(a, 12, 8);
  ASSERT(s != NULL);
  memcpy(s, "Hello World", 12);
  ASSERT(strcmp(s, "Hello World") == 0);

  arena_destroy(a);
  PASS();
}

static void test_arena_save_restore(void) {
  TEST("arena_save / arena_restore");
  Arena *a = arena_new(4096);

  /* Allocate some initial data */
  void *p1 = arena_alloc(a, 32, 8);
  ASSERT(p1 != NULL);

  ArenaCheckpoint cp = arena_save(a);

  /* Allocate more after checkpoint */
  void *p2 = arena_alloc(a, 64, 8);
  ASSERT(p2 != NULL);
  ASSERT(a->current->used > cp.used);

  /* Restore */
  arena_restore(a, cp);
  ASSERT(a->current->used == cp.used);

  /* New allocation should reuse the freed space */
  void *p3 = arena_alloc(a, 64, 8);
  ASSERT(p3 == p2); /* Same address, space was reclaimed */

  arena_destroy(a);
  PASS();
}

static void test_arena_save_restore_across_pages(void) {
  TEST("arena_save / arena_restore across pages");
  Arena *a = arena_new(64);

  /* Allocate in first page */
  arena_alloc(a, 32, 8);
  ArenaCheckpoint cp = arena_save(a);
  ArenaPage *saved_page = a->current;

  /* Force new page allocations */
  arena_alloc(a, 64, 8);
  arena_alloc(a, 64, 8);
  ASSERT(a->current != saved_page);

  /* Restore should free the extra pages */
  arena_restore(a, cp);
  ASSERT(a->current == saved_page);
  ASSERT(a->current->used == cp.used);

  arena_destroy(a);
  PASS();
}

static void test_arena_many_small_allocs(void) {
  TEST("arena_alloc many small allocations");
  Arena *a = arena_new(256);

  for (int i = 0; i < 1000; i++) {
    int *p = (int *)arena_alloc(a, sizeof(int), 8);
    ASSERT(p != NULL);
    *p = i;
    ASSERT(*p == i);
  }

  arena_destroy(a);
  PASS();
}

int main(void) {
  printf("=== Arena Tests ===\n");

  test_arena_new_destroy();
  test_arena_min_page_size();
  test_arena_alloc_basic();
  test_arena_alloc_alignment();
  test_arena_alloc_overflow_to_new_page();
  test_arena_alloc_large_object();
  test_arena_alloc_write_read();
  test_arena_save_restore();
  test_arena_save_restore_across_pages();
  test_arena_many_small_allocs();

  printf("\n%d/%d tests passed\n", tests_passed, tests_run);
  return tests_passed == tests_run ? 0 : 1;
}
