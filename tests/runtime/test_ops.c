/*
 * test_ops.c — Unit tests for arithmetic dispatch.
 *
 * Compile:
 *   clang -Wall -Wextra -g -Ipkg/runtime -o test_ops \
 *       tests/runtime/test_ops.c \
 *       pkg/runtime/core/arena.c pkg/runtime/core/values.c \
 *       pkg/runtime/gmp/gmp_glue.c pkg/runtime/ops/ops.c -lgmp
 */
#include "../../pkg/runtime/gmp/gmp_glue.h"
#include "../../pkg/runtime/ops/ops.h"
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

/* ---- SmallInt Arithmetic ---- */

static void test_add_small(void) {
  TEST("add: small + small");
  OrgValue a = ORG_TAG_SMALL_INT(3);
  OrgValue b = ORG_TAG_SMALL_INT(4);
  OrgValue r = org_add(arena, a, b);
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 7);
  PASS();
}

static void test_sub_small(void) {
  TEST("sub: small - small");
  OrgValue r = org_sub(arena, ORG_TAG_SMALL_INT(10), ORG_TAG_SMALL_INT(3));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 7);
  PASS();
}

static void test_mul_small(void) {
  TEST("mul: small * small");
  OrgValue r = org_mul(arena, ORG_TAG_SMALL_INT(6), ORG_TAG_SMALL_INT(7));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 42);
  PASS();
}

static void test_div_exact(void) {
  TEST("div: exact integer division");
  OrgValue r = org_div(arena, ORG_TAG_SMALL_INT(10), ORG_TAG_SMALL_INT(2));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 5);
  PASS();
}

static void test_div_inexact(void) {
  TEST("div: inexact → Rational");
  OrgValue r = org_div(arena, ORG_TAG_SMALL_INT(3), ORG_TAG_SMALL_INT(2));
  ASSERT(org_is_rational(r));
  /* 3/2 */
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 3) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

static void test_div_zero(void) {
  TEST("div: division by zero → Error");
  OrgValue r = org_div(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(0));
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

static void test_mod_small(void) {
  TEST("mod: small % small");
  OrgValue r = org_mod(arena, ORG_TAG_SMALL_INT(10), ORG_TAG_SMALL_INT(3));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 1);
  PASS();
}

static void test_neg_small(void) {
  TEST("neg: -small");
  OrgValue r = org_neg(arena, ORG_TAG_SMALL_INT(42));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == -42);
  PASS();
}

/* ---- Overflow to BigInt ---- */

static void test_add_overflow(void) {
  TEST("add: SmallInt overflow → BigInt");
  OrgValue max = ORG_TAG_SMALL_INT(ORG_SMALL_MAX);
  OrgValue one = ORG_TAG_SMALL_INT(1);
  OrgValue r = org_add(arena, max, one);
  /* Should overflow to BigInt */
  ASSERT(ORG_IS_PTR(r));
  ASSERT(org_get_type(r) == ORG_TYPE_BIGINT);
  /* Verify value: ORG_SMALL_MAX + 1 */
  mpz_t expected;
  mpz_init_set_si(expected, (long)ORG_SMALL_MAX);
  mpz_add_ui(expected, expected, 1);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_bigint_add(void) {
  TEST("add: BigInt + BigInt");
  OrgValue a = org_make_bigint_str(arena, "99999999999999999999");
  OrgValue b = org_make_bigint_str(arena, "1");
  OrgValue r = org_add(arena, a, b);
  ASSERT(ORG_IS_PTR(r));
  ASSERT(org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init_set_str(expected, "100000000000000000000", 10);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_bigint_normalize(void) {
  TEST("BigInt normalizes back to SmallInt when fits");
  OrgValue big = org_make_bigint_si(arena, 42);
  OrgValue normal = org_normalize_int(big);
  ASSERT(ORG_IS_SMALL(normal));
  ASSERT(ORG_UNTAG_SMALL_INT(normal) == 42);
  PASS();
}

/* ---- Rational Arithmetic ---- */

static void test_rational_add(void) {
  TEST("add: Rational + Rational");
  OrgValue a = org_make_rational_str(arena, "1", "3");
  OrgValue b = org_make_rational_str(arena, "1", "6");
  OrgValue r = org_add(arena, a, b);
  /* 1/3 + 1/6 = 1/2 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 1) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

static void test_rational_to_integer(void) {
  TEST("add: Rational simplifies to Integer (2/3 + 1/3 = 1)");
  OrgValue a = org_make_rational_str(arena, "2", "3");
  OrgValue b = org_make_rational_str(arena, "1", "3");
  OrgValue r = org_add(arena, a, b);
  /* 2/3 + 1/3 = 1 → should be SmallInt */
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 1);
  PASS();
}

static void test_int_rational_promotion(void) {
  TEST("add: Integer + Rational → Rational");
  OrgValue a = ORG_TAG_SMALL_INT(1);
  OrgValue b = org_make_rational_str(arena, "1", "2");
  OrgValue r = org_add(arena, a, b);
  /* 1 + 1/2 = 3/2 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 3) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

/* ---- Decimal Arithmetic ---- */

static void test_decimal_add(void) {
  TEST("add: Decimal + Decimal");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue b = org_make_decimal_str(arena, "2.3");
  OrgValue r = org_add(arena, a, b);
  /* 1.5 + 2.3 = 3.8 */
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 38, 10);
  mpq_canonicalize(expected);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_decimal_promotion(void) {
  TEST("add: Integer + Decimal → Decimal");
  OrgValue a = ORG_TAG_SMALL_INT(1);
  OrgValue b = org_make_decimal_str(arena, "0.5");
  OrgValue r = org_add(arena, a, b);
  /* 1 + 0.5 = 1.5 → Decimal */
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 3, 2);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

/* ---- Power ---- */

static void test_pow_small(void) {
  TEST("pow: 2 ** 10 = 1024");
  OrgValue r = org_pow(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(10));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 1024);
  PASS();
}

static void test_pow_big(void) {
  TEST("pow: 2 ** 64 → BigInt");
  OrgValue r = org_pow(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(64));
  ASSERT(ORG_IS_PTR(r));
  ASSERT(org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init(expected);
  mpz_ui_pow_ui(expected, 2, 64);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_pow_negative_exp(void) {
  TEST("pow: negative exponent → Error");
  OrgValue r = org_pow(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(-1));
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

/* ---- Comparison ---- */

static void test_eq_small(void) {
  TEST("eq: 42 = 42");
  OrgValue r = org_eq(arena, ORG_TAG_SMALL_INT(42), ORG_TAG_SMALL_INT(42));
  ASSERT(ORG_IS_TRUE(r));
  PASS();
}

static void test_ne_small(void) {
  TEST("ne: 1 <> 2");
  OrgValue r = org_ne(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2));
  ASSERT(ORG_IS_TRUE(r));
  PASS();
}

static void test_lt_small(void) {
  TEST("lt: 1 < 2");
  OrgValue r = org_lt(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2));
  ASSERT(ORG_IS_TRUE(r));
  PASS();
}

static void test_eq_cross_type(void) {
  TEST("eq: SmallInt(6) = Rational(6/1)");
  OrgValue a = ORG_TAG_SMALL_INT(6);
  OrgValue b = org_make_rational_str(arena, "6", "1");
  OrgValue r = org_eq(arena, a, b);
  /* 6/1 canonicalizes to integer → wrap_mpq_rational returns SmallInt */
  ASSERT(ORG_IS_TRUE(r));
  PASS();
}

static void test_eq_decimal_int(void) {
  TEST("eq: Decimal(2.0) = SmallInt(2)");
  OrgValue a = org_make_decimal_str(arena, "2.0");
  OrgValue b = ORG_TAG_SMALL_INT(2);
  OrgValue r = org_eq(arena, a, b);
  ASSERT(ORG_IS_TRUE(r));
  PASS();
}

/* ---- Error propagation ---- */

static void test_error_propagation(void) {
  TEST("error propagation through arithmetic");
  OrgValue r = org_add(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1));
  ASSERT(ORG_IS_ERROR(r));
  r = org_mul(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR);
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

int main(void) {
  printf("=== Ops Tests ===\n");
  setup();

  /* SmallInt */
  test_add_small();
  test_sub_small();
  test_mul_small();
  test_div_exact();
  test_div_inexact();
  test_div_zero();
  test_mod_small();
  test_neg_small();

  /* Overflow / BigInt */
  test_add_overflow();
  test_bigint_add();
  test_bigint_normalize();

  /* Rational */
  test_rational_add();
  test_rational_to_integer();
  test_int_rational_promotion();

  /* Decimal */
  test_decimal_add();
  test_decimal_promotion();

  /* Power */
  test_pow_small();
  test_pow_big();
  test_pow_negative_exp();

  /* Comparison */
  test_eq_small();
  test_ne_small();
  test_lt_small();
  test_eq_cross_type();
  test_eq_decimal_int();

  /* Error */
  test_error_propagation();

  teardown();
  printf("\n%d/%d tests passed\n", tests_passed, tests_run);
  return tests_passed == tests_run ? 0 : 1;
}
