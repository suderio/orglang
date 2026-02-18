/*
 * test_table.c — Unit tests for OrgTable.
 *
 * Compile:
 *   clang -Wall -Wextra -g -Ipkg/runtime -o test_table \
 *       tests/runtime/test_table.c \
 *       pkg/runtime/core/arena.c pkg/runtime/core/values.c \
 *       pkg/runtime/gmp/gmp_glue.c pkg/runtime/table/table.c -lgmp
 */
#include "../../pkg/runtime/gmp/gmp_glue.h"
#include "../../pkg/runtime/table/table.h"
#include <stdio.h>
#include <string.h>

static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name)                                                             \
  do {                                                                         \
    tests_run++;                                                               \
    printf("  %-55s", name);                                                   \
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

static Arena *arena;

static void setup(void) {
  arena = arena_new(65536);
  org_gmp_init();
  org_gmp_set_arena(arena);
}

static void teardown(void) { arena_destroy(arena); }

/* ========== Construction ========== */

static void test_new_table(void) {
  TEST("table: new creates empty table");
  OrgValue t = org_table_new(arena);
  ASSERT(ORG_IS_PTR(t));
  ASSERT(org_get_type(t) == ORG_TYPE_TABLE);
  ASSERT(org_table_count(t) == 0);
  PASS();
}

static void test_new_sized(void) {
  TEST("table: new_sized with capacity hint");
  OrgValue t = org_table_new_sized(arena, 100);
  ASSERT(ORG_IS_PTR(t));
  ASSERT(org_table_count(t) == 0);
  PASS();
}

/* ========== String Key Operations ========== */

static void test_set_get_string(void) {
  TEST("table: set/get with string key");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "hello", 5);
  OrgValue val = ORG_TAG_SMALL_INT(42);

  org_table_set(arena, t, key, val);
  ASSERT(org_table_count(t) == 1);

  OrgValue got = org_table_get(t, key);
  ASSERT(ORG_IS_SMALL(got));
  ASSERT(ORG_UNTAG_SMALL_INT(got) == 42);
  PASS();
}

static void test_get_cstr(void) {
  TEST("table: get_cstr convenience lookup");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "name", 4);
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(99));

  OrgValue got = org_table_get_cstr(t, "name");
  ASSERT(ORG_IS_SMALL(got));
  ASSERT(ORG_UNTAG_SMALL_INT(got) == 99);
  PASS();
}

static void test_get_cstr_not_found(void) {
  TEST("table: get_cstr not found → Error");
  OrgValue t = org_table_new(arena);
  OrgValue got = org_table_get_cstr(t, "missing");
  ASSERT(ORG_IS_ERROR(got));
  PASS();
}

static void test_set_overwrites(void) {
  TEST("table: set overwrites existing key");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "x", 1);
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(1));
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(2));
  ASSERT(org_table_count(t) == 1);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, key)) == 2);
  PASS();
}

static void test_string_key_comparison(void) {
  TEST("table: different string objects, same content");
  OrgValue t = org_table_new(arena);
  OrgValue k1 = org_make_string(arena, "abc", 3);
  OrgValue k2 = org_make_string(arena, "abc", 3);
  /* k1 and k2 are different pointers but same content */
  ASSERT(k1 != k2);

  org_table_set(arena, t, k1, ORG_TAG_SMALL_INT(10));
  OrgValue got = org_table_get(t, k2);
  ASSERT(ORG_IS_SMALL(got));
  ASSERT(ORG_UNTAG_SMALL_INT(got) == 10);
  PASS();
}

/* ========== Integer Key Operations ========== */

static void test_set_get_int(void) {
  TEST("table: set/get with integer key");
  OrgValue t = org_table_new(arena);
  OrgValue key = ORG_TAG_SMALL_INT(0);
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(100));

  OrgValue got = org_table_get(t, key);
  ASSERT(ORG_IS_SMALL(got));
  ASSERT(ORG_UNTAG_SMALL_INT(got) == 100);
  PASS();
}

static void test_negative_int_key(void) {
  TEST("table: negative integer key works");
  OrgValue t = org_table_new(arena);
  OrgValue key = ORG_TAG_SMALL_INT(-5);
  org_table_set(arena, t, key, ORG_TRUE);
  ASSERT(ORG_IS_TRUE(org_table_get(t, key)));
  PASS();
}

/* ========== Push (Auto-Index) ========== */

static void test_push(void) {
  TEST("table: push assigns sequential indices");
  OrgValue t = org_table_new(arena);
  org_table_push(arena, t, ORG_TAG_SMALL_INT(10));
  org_table_push(arena, t, ORG_TAG_SMALL_INT(20));
  org_table_push(arena, t, ORG_TAG_SMALL_INT(30));

  ASSERT(org_table_count(t) == 3);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, ORG_TAG_SMALL_INT(0))) == 10);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, ORG_TAG_SMALL_INT(1))) == 20);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, ORG_TAG_SMALL_INT(2))) == 30);
  PASS();
}

/* ========== Has ========== */

static void test_has(void) {
  TEST("table: has returns true/false");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_FALSE(org_table_has(t, key)));
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(1));
  ASSERT(ORG_IS_TRUE(org_table_has(t, key)));
  PASS();
}

/* ========== Not Found ========== */

static void test_get_not_found(void) {
  TEST("table: get missing key → Error");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "nope", 4);
  OrgValue got = org_table_get(t, key);
  ASSERT(ORG_IS_ERROR(got));
  PASS();
}

/* ========== Invalid Operations ========== */

static void test_invalid_key(void) {
  TEST("table: invalid key type → Error");
  OrgValue t = org_table_new(arena);
  /* Boolean is not a valid table key */
  ASSERT(ORG_IS_ERROR(org_table_set(arena, t, ORG_TRUE, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_table_get(t, ORG_TRUE)));
  PASS();
}

static void test_get_non_table(void) {
  TEST("table: operations on non-table → Error");
  OrgValue not_table = ORG_TAG_SMALL_INT(42);
  ASSERT(ORG_IS_ERROR(org_table_get(not_table, ORG_TAG_SMALL_INT(0))));
  ASSERT(ORG_IS_ERROR(
      org_table_set(arena, not_table, ORG_TAG_SMALL_INT(0), ORG_TRUE)));
  ASSERT(ORG_IS_ERROR(org_table_push(arena, not_table, ORG_TRUE)));
  ASSERT(org_table_count(not_table) == 0);
  PASS();
}

static void test_get_cstr_non_table(void) {
  TEST("table: get_cstr on non-table → Error");
  ASSERT(ORG_IS_ERROR(org_table_get_cstr(ORG_TAG_SMALL_INT(1), "x")));
  PASS();
}

/* ========== Resize/Growth ========== */

static void test_many_entries(void) {
  TEST("table: insert 100 entries (forces resize)");
  OrgValue t = org_table_new(arena);
  for (int i = 0; i < 100; i++) {
    OrgValue key = ORG_TAG_SMALL_INT(i);
    OrgValue val = ORG_TAG_SMALL_INT(i * 10);
    org_table_set(arena, t, key, val);
  }
  ASSERT(org_table_count(t) == 100);

  /* Verify all entries */
  for (int i = 0; i < 100; i++) {
    OrgValue got = org_table_get(t, ORG_TAG_SMALL_INT(i));
    ASSERT(ORG_IS_SMALL(got));
    ASSERT(ORG_UNTAG_SMALL_INT(got) == i * 10);
  }
  PASS();
}

static void test_many_string_keys(void) {
  TEST("table: insert 50 string keys (forces resize)");
  OrgValue t = org_table_new(arena);
  char buf[16];
  for (int i = 0; i < 50; i++) {
    int n = snprintf(buf, sizeof(buf), "key_%d", i);
    OrgValue key = org_make_string(arena, buf, (size_t)n);
    org_table_set(arena, t, key, ORG_TAG_SMALL_INT(i));
  }
  ASSERT(org_table_count(t) == 50);

  /* Verify some entries */
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get_cstr(t, "key_0")) == 0);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get_cstr(t, "key_49")) == 49);
  PASS();
}

/* ========== Mixed Key Types ========== */

static void test_mixed_keys(void) {
  TEST("table: mixed string and integer keys");
  OrgValue t = org_table_new(arena);
  OrgValue skey = org_make_string(arena, "name", 4);
  org_table_set(arena, t, skey, ORG_TAG_SMALL_INT(1));
  org_table_set(arena, t, ORG_TAG_SMALL_INT(0), ORG_TAG_SMALL_INT(2));

  ASSERT(org_table_count(t) == 2);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, skey)) == 1);
  ASSERT(ORG_UNTAG_SMALL_INT(org_table_get(t, ORG_TAG_SMALL_INT(0))) == 2);
  PASS();
}

/* ========== UTF-8 String Keys ========== */

static void test_utf8_key(void) {
  TEST("table: UTF-8 string keys (世界)");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "世界", 6);
  org_table_set(arena, t, key, ORG_TAG_SMALL_INT(42));

  OrgValue got = org_table_get_cstr(t, "世界");
  ASSERT(ORG_IS_SMALL(got));
  ASSERT(ORG_UNTAG_SMALL_INT(got) == 42);
  PASS();
}

/* ========== Hash Function ========== */

static void test_hash_consistency(void) {
  TEST("hash: same content produces same hash");
  OrgValue k1 = org_make_string(arena, "test", 4);
  OrgValue k2 = org_make_string(arena, "test", 4);
  ASSERT(org_hash_value(k1) == org_hash_value(k2));
  PASS();
}

static void test_hash_int(void) {
  TEST("hash: integer keys produce non-zero hashes");
  uint32_t h0 = org_hash_value(ORG_TAG_SMALL_INT(0));
  uint32_t h1 = org_hash_value(ORG_TAG_SMALL_INT(1));
  /* Different ints should (likely) have different hashes */
  ASSERT(h0 != h1);
  PASS();
}

static void test_key_equal_strings(void) {
  TEST("key_equal: same-content strings are equal");
  OrgValue k1 = org_make_string(arena, "xyz", 3);
  OrgValue k2 = org_make_string(arena, "xyz", 3);
  OrgValue k3 = org_make_string(arena, "abc", 3);
  ASSERT(org_key_equal(k1, k2) == 1);
  ASSERT(org_key_equal(k1, k3) == 0);
  PASS();
}

static void test_key_equal_ints(void) {
  TEST("key_equal: same ints are equal, different are not");
  ASSERT(org_key_equal(ORG_TAG_SMALL_INT(5), ORG_TAG_SMALL_INT(5)) == 1);
  ASSERT(org_key_equal(ORG_TAG_SMALL_INT(5), ORG_TAG_SMALL_INT(6)) == 0);
  PASS();
}

static void test_key_equal_cross_type(void) {
  TEST("key_equal: int vs string are not equal");
  OrgValue k1 = ORG_TAG_SMALL_INT(1);
  OrgValue k2 = org_make_string(arena, "1", 1);
  ASSERT(org_key_equal(k1, k2) == 0);
  PASS();
}

/* ========== Table Storing Various Values ========== */

static void test_table_stores_table(void) {
  TEST("table: stores nested table as value");
  OrgValue outer = org_table_new(arena);
  OrgValue inner = org_table_new(arena);
  OrgValue key = org_make_string(arena, "child", 5);
  org_table_set(arena, outer, key, inner);

  OrgValue got = org_table_get(outer, key);
  ASSERT(ORG_IS_PTR(got) && org_get_type(got) == ORG_TYPE_TABLE);
  PASS();
}

static void test_table_stores_string(void) {
  TEST("table: stores string value");
  OrgValue t = org_table_new(arena);
  OrgValue key = org_make_string(arena, "msg", 3);
  OrgValue val = org_make_string(arena, "hello", 5);
  org_table_set(arena, t, key, val);

  OrgValue got = org_table_get(t, key);
  ASSERT(ORG_IS_PTR(got) && org_get_type(got) == ORG_TYPE_STRING);
  OrgString *s = (OrgString *)ORG_GET_PTR(got);
  ASSERT(s->byte_len == 5);
  ASSERT(memcmp(s->data, "hello", 5) == 0);
  PASS();
}

int main(void) {
  printf("=== Table Tests ===\n");
  setup();

  /* Construction */
  test_new_table();
  test_new_sized();

  /* String key ops */
  test_set_get_string();
  test_get_cstr();
  test_get_cstr_not_found();
  test_set_overwrites();
  test_string_key_comparison();

  /* Integer key ops */
  test_set_get_int();
  test_negative_int_key();

  /* Push */
  test_push();

  /* Has */
  test_has();

  /* Not found */
  test_get_not_found();

  /* Invalid operations */
  test_invalid_key();
  test_get_non_table();
  test_get_cstr_non_table();

  /* Growth */
  test_many_entries();
  test_many_string_keys();

  /* Mixed keys */
  test_mixed_keys();

  /* UTF-8 */
  test_utf8_key();

  /* Hash/key_equal */
  test_hash_consistency();
  test_hash_int();
  test_key_equal_strings();
  test_key_equal_ints();
  test_key_equal_cross_type();

  /* Various values */
  test_table_stores_table();
  test_table_stores_string();

  teardown();
  printf("\n%d/%d tests passed\n", tests_passed, tests_run);
  return tests_passed == tests_run ? 0 : 1;
}
