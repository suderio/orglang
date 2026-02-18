/*
 * test_values.c — Unit tests for the OrgValue tagged value system.
 *
 * Compile:
 *   clang -Wall -Wextra -g -o test_values \
 *       test_values.c ../../pkg/runtime/core/values.c
 * ../../pkg/runtime/core/arena.c
 */
#include "../../pkg/runtime/core/values.h"
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

/* ---- Small Integer Tests ---- */

static void test_small_int_zero(void) {
  TEST("small int: zero");
  OrgValue v = ORG_TAG_SMALL_INT(0);
  ASSERT(ORG_IS_SMALL(v));
  ASSERT(!ORG_IS_PTR(v));
  ASSERT(!ORG_IS_SPECIAL(v));
  ASSERT(ORG_UNTAG_SMALL_INT(v) == 0);
  PASS();
}

static void test_small_int_positive(void) {
  TEST("small int: positive");
  OrgValue v = ORG_TAG_SMALL_INT(42);
  ASSERT(ORG_IS_SMALL(v));
  ASSERT(ORG_UNTAG_SMALL_INT(v) == 42);
  PASS();
}

static void test_small_int_negative(void) {
  TEST("small int: negative");
  OrgValue v = ORG_TAG_SMALL_INT(-100);
  ASSERT(ORG_IS_SMALL(v));
  ASSERT(ORG_UNTAG_SMALL_INT(v) == -100);
  PASS();
}

static void test_small_int_max(void) {
  TEST("small int: max value");
  OrgValue v = ORG_TAG_SMALL_INT(ORG_SMALL_MAX);
  ASSERT(ORG_IS_SMALL(v));
  ASSERT(ORG_UNTAG_SMALL_INT(v) == ORG_SMALL_MAX);
  PASS();
}

static void test_small_int_min(void) {
  TEST("small int: min value");
  OrgValue v = ORG_TAG_SMALL_INT(ORG_SMALL_MIN);
  ASSERT(ORG_IS_SMALL(v));
  ASSERT(ORG_UNTAG_SMALL_INT(v) == ORG_SMALL_MIN);
  PASS();
}

static void test_small_int_fits(void) {
  TEST("small int: org_small_fits");
  ASSERT(org_small_fits(0));
  ASSERT(org_small_fits(42));
  ASSERT(org_small_fits(-42));
  ASSERT(org_small_fits(ORG_SMALL_MAX));
  ASSERT(org_small_fits(ORG_SMALL_MIN));
  ASSERT(!org_small_fits(ORG_SMALL_MAX + 1));
  ASSERT(!org_small_fits(ORG_SMALL_MIN - 1));
  PASS();
}

/* ---- Special Value Tests ---- */

static void test_special_true(void) {
  TEST("special: true");
  ASSERT(ORG_IS_SPECIAL(ORG_TRUE));
  ASSERT(ORG_IS_TRUE(ORG_TRUE));
  ASSERT(ORG_IS_BOOL(ORG_TRUE));
  ASSERT(!ORG_IS_FALSE(ORG_TRUE));
  ASSERT(!ORG_IS_ERROR(ORG_TRUE));
  ASSERT(!ORG_IS_SMALL(ORG_TRUE));
  ASSERT(!ORG_IS_PTR(ORG_TRUE));
  PASS();
}

static void test_special_false(void) {
  TEST("special: false");
  ASSERT(ORG_IS_SPECIAL(ORG_FALSE));
  ASSERT(ORG_IS_FALSE(ORG_FALSE));
  ASSERT(ORG_IS_BOOL(ORG_FALSE));
  ASSERT(!ORG_IS_TRUE(ORG_FALSE));
  PASS();
}

static void test_special_error(void) {
  TEST("special: error");
  ASSERT(ORG_IS_SPECIAL(ORG_ERROR));
  ASSERT(ORG_IS_ERROR(ORG_ERROR));
  ASSERT(!ORG_IS_BOOL(ORG_ERROR));
  ASSERT(!ORG_IS_SMALL(ORG_ERROR));
  PASS();
}

static void test_special_unused(void) {
  TEST("special: unused");
  ASSERT(ORG_IS_SPECIAL(ORG_UNUSED));
  ASSERT(ORG_IS_UNUSED(ORG_UNUSED));
  ASSERT(!ORG_IS_ERROR(ORG_UNUSED));
  ASSERT(!ORG_IS_BOOL(ORG_UNUSED));
  PASS();
}

static void test_bool_macro(void) {
  TEST("special: ORG_BOOL macro");
  ASSERT(ORG_BOOL(1) == ORG_TRUE);
  ASSERT(ORG_BOOL(0) == ORG_FALSE);
  ASSERT(ORG_BOOL(42) == ORG_TRUE);
  PASS();
}

/* ---- All specials are distinct ---- */

static void test_specials_distinct(void) {
  TEST("special values are all distinct");
  ASSERT(ORG_TRUE != ORG_FALSE);
  ASSERT(ORG_TRUE != ORG_ERROR);
  ASSERT(ORG_TRUE != ORG_UNUSED);
  ASSERT(ORG_FALSE != ORG_ERROR);
  ASSERT(ORG_FALSE != ORG_UNUSED);
  ASSERT(ORG_ERROR != ORG_UNUSED);
  PASS();
}

/* ---- String Tests ---- */

static void test_string_ascii(void) {
  TEST("string: ASCII");
  Arena *a = arena_new(4096);
  OrgValue v = org_make_string(a, "hello", 5);
  ASSERT(ORG_IS_PTR(v));
  ASSERT(org_get_type(v) == ORG_TYPE_STRING);
  ASSERT(org_string_byte_len(v) == 5);
  ASSERT(org_string_codepoint_len(v) == 5);
  ASSERT(memcmp(org_string_data(v), "hello", 5) == 0);
  arena_destroy(a);
  PASS();
}

static void test_string_utf8_multibyte(void) {
  TEST("string: UTF-8 multibyte (世界)");
  Arena *a = arena_new(4096);
  const char *s = "世界"; /* 2 codepoints, 6 bytes (3 bytes each) */
  OrgValue v = org_make_string(a, s, 6);
  ASSERT(ORG_IS_PTR(v));
  ASSERT(org_string_byte_len(v) == 6);
  ASSERT(org_string_codepoint_len(v) == 2);
  arena_destroy(a);
  PASS();
}

static void test_string_utf8_emoji(void) {
  TEST("string: UTF-8 emoji (🌍💩)");
  Arena *a = arena_new(4096);
  const char *s = "🌍💩"; /* 2 codepoints, 8 bytes (4 bytes each) */
  OrgValue v = org_make_string(a, s, 8);
  ASSERT(ORG_IS_PTR(v));
  ASSERT(org_string_byte_len(v) == 8);
  ASSERT(org_string_codepoint_len(v) == 2);
  arena_destroy(a);
  PASS();
}

static void test_string_empty(void) {
  TEST("string: empty");
  Arena *a = arena_new(4096);
  OrgValue v = org_make_string(a, "", 0);
  ASSERT(ORG_IS_PTR(v));
  ASSERT(org_string_byte_len(v) == 0);
  ASSERT(org_string_codepoint_len(v) == 0);
  arena_destroy(a);
  PASS();
}

/* ---- Type Name Tests ---- */

static void test_type_name(void) {
  TEST("org_type_name");
  ASSERT(strcmp(org_type_name(ORG_TAG_SMALL_INT(1)), "SmallInt") == 0);
  ASSERT(strcmp(org_type_name(ORG_TRUE), "Boolean(true)") == 0);
  ASSERT(strcmp(org_type_name(ORG_FALSE), "Boolean(false)") == 0);
  ASSERT(strcmp(org_type_name(ORG_ERROR), "Error") == 0);
  ASSERT(strcmp(org_type_name(ORG_UNUSED), "Unused") == 0);

  Arena *a = arena_new(4096);
  OrgValue s = org_make_string(a, "test", 4);
  ASSERT(strcmp(org_type_name(s), "String") == 0);
  arena_destroy(a);
  PASS();
}

/* ---- Pointer alignment guarantee ---- */

static void test_pointer_alignment(void) {
  TEST("arena pointers have tag bits clear (aligned)");
  Arena *a = arena_new(4096);
  void *p = arena_alloc(a, 32, 8);
  OrgValue v = ORG_TAG_PTR_VAL(p);
  ASSERT(ORG_IS_PTR(v));
  ASSERT(ORG_GET_PTR(v) == (OrgObject *)p);
  arena_destroy(a);
  PASS();
}

int main(void) {
  printf("=== Values Tests ===\n");

  test_small_int_zero();
  test_small_int_positive();
  test_small_int_negative();
  test_small_int_max();
  test_small_int_min();
  test_small_int_fits();

  test_special_true();
  test_special_false();
  test_special_error();
  test_special_unused();
  test_bool_macro();
  test_specials_distinct();

  test_string_ascii();
  test_string_utf8_multibyte();
  test_string_utf8_emoji();
  test_string_empty();

  test_type_name();
  test_pointer_alignment();

  printf("\n%d/%d tests passed\n", tests_passed, tests_run);
  return tests_passed == tests_run ? 0 : 1;
}
