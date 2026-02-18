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

/* ========== SmallInt Basic ========== */

static void test_add_small(void) {
  TEST("add: small + small");
  OrgValue r = org_add(arena, ORG_TAG_SMALL_INT(3), ORG_TAG_SMALL_INT(4));
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

/* ========== Overflow / BigInt ========== */

static void test_add_overflow(void) {
  TEST("add: SmallInt overflow → BigInt");
  OrgValue max = ORG_TAG_SMALL_INT(ORG_SMALL_MAX);
  OrgValue r = org_add(arena, max, ORG_TAG_SMALL_INT(1));
  ASSERT(ORG_IS_PTR(r));
  ASSERT(org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init_set_si(expected, (long)ORG_SMALL_MAX);
  mpz_add_ui(expected, expected, 1);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_sub_overflow(void) {
  TEST("sub: SmallInt overflow → BigInt");
  OrgValue min = ORG_TAG_SMALL_INT(ORG_SMALL_MIN);
  OrgValue r = org_sub(arena, min, ORG_TAG_SMALL_INT(1));
  ASSERT(ORG_IS_PTR(r));
  ASSERT(org_get_type(r) == ORG_TYPE_BIGINT);
  PASS();
}

static void test_mul_overflow(void) {
  TEST("mul: SmallInt overflow → BigInt");
  OrgValue big = ORG_TAG_SMALL_INT(ORG_SMALL_MAX);
  OrgValue r = org_mul(arena, big, ORG_TAG_SMALL_INT(2));
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  PASS();
}

static void test_bigint_add(void) {
  TEST("add: BigInt + BigInt");
  OrgValue a = org_make_bigint_str(arena, "99999999999999999999");
  OrgValue b = org_make_bigint_str(arena, "1");
  OrgValue r = org_add(arena, a, b);
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
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

static void test_normalize_non_bigint(void) {
  TEST("normalize: non-BigInt returns unchanged");
  OrgValue s = ORG_TAG_SMALL_INT(42);
  ASSERT(org_normalize_int(s) == s);
  OrgValue t = ORG_TRUE;
  ASSERT(org_normalize_int(t) == t);
  PASS();
}

/* ========== Sub: all type paths ========== */

static void test_sub_bigint(void) {
  TEST("sub: BigInt - SmallInt");
  OrgValue a = org_make_bigint_str(arena, "100000000000000000000");
  OrgValue r = org_sub(arena, a, ORG_TAG_SMALL_INT(1));
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init_set_str(expected, "99999999999999999999", 10);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_sub_rational(void) {
  TEST("sub: Rational - Rational");
  OrgValue a = org_make_rational_str(arena, "5", "6");
  OrgValue b = org_make_rational_str(arena, "1", "3");
  OrgValue r = org_sub(arena, a, b);
  /* 5/6 - 1/3 = 1/2 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 1) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

static void test_sub_decimal(void) {
  TEST("sub: Decimal - Decimal");
  OrgValue a = org_make_decimal_str(arena, "5.5");
  OrgValue b = org_make_decimal_str(arena, "2.3");
  OrgValue r = org_sub(arena, a, b);
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 32, 10);
  mpq_canonicalize(expected);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_sub_int_decimal(void) {
  TEST("sub: Integer - Decimal → Decimal");
  OrgValue r =
      org_sub(arena, ORG_TAG_SMALL_INT(3), org_make_decimal_str(arena, "1.5"));
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 3, 2);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_sub_int_rational(void) {
  TEST("sub: Integer - Rational → Rational");
  OrgValue r = org_sub(arena, ORG_TAG_SMALL_INT(2),
                       org_make_rational_str(arena, "1", "3"));
  /* 2 - 1/3 = 5/3 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 5) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 3) == 0);
  PASS();
}

static void test_sub_error(void) {
  TEST("sub: error propagation");
  ASSERT(ORG_IS_ERROR(org_sub(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_sub(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  PASS();
}

static void test_sub_non_numeric(void) {
  TEST("sub: non-numeric → Error");
  OrgValue s = org_make_string(arena, "hi", 2);
  ASSERT(ORG_IS_ERROR(org_sub(arena, s, ORG_TAG_SMALL_INT(1))));
  PASS();
}

/* ========== Mul: all type paths ========== */

static void test_mul_bigint(void) {
  TEST("mul: BigInt * BigInt");
  OrgValue a = org_make_bigint_str(arena, "99999999999999999999");
  OrgValue b = org_make_bigint_str(arena, "2");
  OrgValue r = org_mul(arena, a, b);
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init_set_str(expected, "199999999999999999998", 10);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_mul_rational(void) {
  TEST("mul: Rational * Rational");
  OrgValue a = org_make_rational_str(arena, "2", "3");
  OrgValue b = org_make_rational_str(arena, "3", "4");
  OrgValue r = org_mul(arena, a, b);
  /* 2/3 * 3/4 = 1/2 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 1) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

static void test_mul_decimal(void) {
  TEST("mul: Decimal * Decimal");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue b = org_make_decimal_str(arena, "2.0");
  OrgValue r = org_mul(arena, a, b);
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 3, 1);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_mul_int_rational(void) {
  TEST("mul: Integer * Rational → Rational");
  OrgValue r = org_mul(arena, ORG_TAG_SMALL_INT(3),
                       org_make_rational_str(arena, "1", "2"));
  /* 3 * 1/2 = 3/2 */
  ASSERT(org_is_rational(r));
  PASS();
}

static void test_mul_int_decimal(void) {
  TEST("mul: Integer * Decimal → Decimal");
  OrgValue r =
      org_mul(arena, ORG_TAG_SMALL_INT(2), org_make_decimal_str(arena, "1.5"));
  ASSERT(org_is_decimal(r));
  PASS();
}

static void test_mul_error(void) {
  TEST("mul: error propagation");
  ASSERT(ORG_IS_ERROR(org_mul(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_mul(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  PASS();
}

static void test_mul_non_numeric(void) {
  TEST("mul: non-numeric → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_mul(arena, ORG_TAG_SMALL_INT(1), s)));
  PASS();
}

/* ========== Div: all type paths ========== */

static void test_div_bigint(void) {
  TEST("div: BigInt / BigInt (exact)");
  OrgValue a = org_make_bigint_str(arena, "100000000000000000000");
  OrgValue b = org_make_bigint_str(arena, "2");
  OrgValue r = org_div(arena, a, b);
  mpz_t expected;
  mpz_init_set_str(expected, "50000000000000000000", 10);
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_div_bigint_inexact(void) {
  TEST("div: BigInt / BigInt (inexact → Rational)");
  OrgValue a = org_make_bigint_str(arena, "100000000000000000000");
  OrgValue b = org_make_bigint_str(arena, "3");
  OrgValue r = org_div(arena, a, b);
  ASSERT(org_is_rational(r));
  PASS();
}

static void test_div_bigint_zero(void) {
  TEST("div: BigInt / BigInt(0) → Error");
  OrgValue a = org_make_bigint_str(arena, "123");
  OrgValue b = org_make_bigint_si(arena, 0);
  OrgValue r = org_div(arena, a, b);
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

static void test_div_rational(void) {
  TEST("div: Rational / Rational");
  OrgValue a = org_make_rational_str(arena, "1", "2");
  OrgValue b = org_make_rational_str(arena, "1", "3");
  OrgValue r = org_div(arena, a, b);
  /* (1/2) / (1/3) = 3/2 */
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 3) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 2) == 0);
  PASS();
}

static void test_div_decimal(void) {
  TEST("div: Decimal / Decimal");
  OrgValue a = org_make_decimal_str(arena, "7.5");
  OrgValue b = org_make_decimal_str(arena, "2.5");
  OrgValue r = org_div(arena, a, b);
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 3, 1);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_div_int_decimal(void) {
  TEST("div: Integer / Decimal → Decimal");
  OrgValue r =
      org_div(arena, ORG_TAG_SMALL_INT(3), org_make_decimal_str(arena, "1.5"));
  ASSERT(org_is_decimal(r));
  PASS();
}

static void test_div_rational_zero(void) {
  TEST("div: Rational / 0 → Error");
  OrgValue a = org_make_rational_str(arena, "1", "2");
  OrgValue b = org_make_rational_str(arena, "0", "1");
  OrgValue r = org_div(arena, a, b);
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

static void test_div_decimal_zero(void) {
  TEST("div: Decimal / 0.0 → Error");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue b = org_make_decimal_str(arena, "0.0");
  OrgValue r = org_div(arena, a, b);
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

static void test_div_error(void) {
  TEST("div: error propagation");
  ASSERT(ORG_IS_ERROR(org_div(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_div(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  PASS();
}

static void test_div_non_numeric(void) {
  TEST("div: non-numeric → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_div(arena, s, ORG_TAG_SMALL_INT(1))));
  PASS();
}

/* ========== Mod: all paths ========== */

static void test_mod_bigint(void) {
  TEST("mod: BigInt % SmallInt");
  OrgValue a = org_make_bigint_str(arena, "100000000000000000003");
  OrgValue r = org_mod(arena, a, ORG_TAG_SMALL_INT(10));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 3);
  PASS();
}

static void test_mod_zero(void) {
  TEST("mod: x % 0 → Error");
  ASSERT(ORG_IS_ERROR(
      org_mod(arena, ORG_TAG_SMALL_INT(10), ORG_TAG_SMALL_INT(0))));
  PASS();
}

static void test_mod_rational_error(void) {
  TEST("mod: Rational % Integer → Error");
  OrgValue a = org_make_rational_str(arena, "1", "2");
  ASSERT(ORG_IS_ERROR(org_mod(arena, a, ORG_TAG_SMALL_INT(1))));
  PASS();
}

static void test_mod_decimal_error(void) {
  TEST("mod: Decimal % Integer → Error");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  ASSERT(ORG_IS_ERROR(org_mod(arena, a, ORG_TAG_SMALL_INT(1))));
  PASS();
}

static void test_mod_error(void) {
  TEST("mod: error propagation");
  ASSERT(ORG_IS_ERROR(org_mod(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_mod(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  PASS();
}

/* ========== Neg: all type paths ========== */

static void test_neg_bigint(void) {
  TEST("neg: -BigInt");
  OrgValue a = org_make_bigint_str(arena, "99999999999999999999");
  OrgValue r = org_neg(arena, a);
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  mpz_t expected;
  mpz_init_set_str(expected, "-99999999999999999999", 10);
  ASSERT(mpz_cmp(*org_get_bigint(r), expected) == 0);
  mpz_clear(expected);
  PASS();
}

static void test_neg_rational(void) {
  TEST("neg: -Rational");
  OrgValue a = org_make_rational_str(arena, "3", "4");
  OrgValue r = org_neg(arena, a);
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), -3) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 4) == 0);
  PASS();
}

static void test_neg_decimal(void) {
  TEST("neg: -Decimal");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue r = org_neg(arena, a);
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, -3, 2);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_neg_error(void) {
  TEST("neg: Error → Error");
  ASSERT(ORG_IS_ERROR(org_neg(arena, ORG_ERROR)));
  PASS();
}

static void test_neg_non_numeric(void) {
  TEST("neg: non-numeric → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_neg(arena, s)));
  PASS();
}

/* ========== Rational Arithmetic ========== */

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

/* ========== Decimal Arithmetic ========== */

static void test_decimal_add(void) {
  TEST("add: Decimal + Decimal");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue b = org_make_decimal_str(arena, "2.3");
  OrgValue r = org_add(arena, a, b);
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
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 3, 2);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_decimal_rational_promotion(void) {
  TEST("add: Rational + Decimal → Decimal");
  OrgValue a = org_make_rational_str(arena, "1", "3");
  OrgValue b = org_make_decimal_str(arena, "0.5");
  OrgValue r = org_add(arena, a, b);
  ASSERT(org_is_decimal(r));
  PASS();
}

/* ========== Power ========== */

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
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
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

static void test_pow_rational(void) {
  TEST("pow: (1/2) ** 3 = 1/8");
  OrgValue base = org_make_rational_str(arena, "1", "2");
  OrgValue r = org_pow(arena, base, ORG_TAG_SMALL_INT(3));
  ASSERT(org_is_rational(r));
  mpq_t *q = org_get_rational(r);
  ASSERT(mpz_cmp_si(mpq_numref(*q), 1) == 0);
  ASSERT(mpz_cmp_si(mpq_denref(*q), 8) == 0);
  PASS();
}

static void test_pow_decimal(void) {
  TEST("pow: 1.5 ** 2 = 2.25");
  OrgValue base = org_make_decimal_str(arena, "1.5");
  OrgValue r = org_pow(arena, base, ORG_TAG_SMALL_INT(2));
  ASSERT(org_is_decimal(r));
  mpq_t expected;
  mpq_init(expected);
  mpq_set_si(expected, 9, 4);
  ASSERT(mpq_equal(*org_get_decimal(r), expected));
  mpq_clear(expected);
  PASS();
}

static void test_pow_zero(void) {
  TEST("pow: x ** 0 = 1");
  OrgValue r = org_pow(arena, ORG_TAG_SMALL_INT(999), ORG_TAG_SMALL_INT(0));
  ASSERT(ORG_IS_SMALL(r));
  ASSERT(ORG_UNTAG_SMALL_INT(r) == 1);
  PASS();
}

static void test_pow_bigint_base(void) {
  TEST("pow: BigInt ** 2");
  OrgValue base = org_make_bigint_str(arena, "99999999999999999999");
  OrgValue r = org_pow(arena, base, ORG_TAG_SMALL_INT(2));
  ASSERT(ORG_IS_PTR(r) && org_get_type(r) == ORG_TYPE_BIGINT);
  PASS();
}

static void test_pow_non_numeric_base(void) {
  TEST("pow: non-numeric base → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_pow(arena, s, ORG_TAG_SMALL_INT(2))));
  PASS();
}

static void test_pow_non_int_exp(void) {
  TEST("pow: non-integer exponent → Error");
  OrgValue r = org_pow(arena, ORG_TAG_SMALL_INT(2),
                       org_make_rational_str(arena, "1", "2"));
  ASSERT(ORG_IS_ERROR(r));
  PASS();
}

static void test_pow_error(void) {
  TEST("pow: error propagation");
  ASSERT(ORG_IS_ERROR(org_pow(arena, ORG_ERROR, ORG_TAG_SMALL_INT(2))));
  ASSERT(ORG_IS_ERROR(org_pow(arena, ORG_TAG_SMALL_INT(2), ORG_ERROR)));
  PASS();
}

/* ========== Comparison: all operators and types ========== */

static void test_eq_small(void) {
  TEST("eq: 42 = 42");
  ASSERT(
      ORG_IS_TRUE(org_eq(arena, ORG_TAG_SMALL_INT(42), ORG_TAG_SMALL_INT(42))));
  ASSERT(ORG_IS_FALSE(
      org_eq(arena, ORG_TAG_SMALL_INT(42), ORG_TAG_SMALL_INT(43))));
  PASS();
}

static void test_ne_small(void) {
  TEST("ne: 1 <> 2");
  ASSERT(
      ORG_IS_TRUE(org_ne(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_FALSE(org_ne(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(1))));
  PASS();
}

static void test_lt_small(void) {
  TEST("lt: 1 < 2");
  ASSERT(
      ORG_IS_TRUE(org_lt(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_FALSE(org_lt(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(1))));
  PASS();
}

static void test_le_small(void) {
  TEST("le: 2 <= 2 and 1 <= 2");
  ASSERT(
      ORG_IS_TRUE(org_le(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_TRUE(org_le(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_FALSE(org_le(arena, ORG_TAG_SMALL_INT(3), ORG_TAG_SMALL_INT(2))));
  PASS();
}

static void test_gt_small(void) {
  TEST("gt: 3 > 2");
  ASSERT(
      ORG_IS_TRUE(org_gt(arena, ORG_TAG_SMALL_INT(3), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_FALSE(org_gt(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2))));
  PASS();
}

static void test_ge_small(void) {
  TEST("ge: 2 >= 2 and 3 >= 2");
  ASSERT(
      ORG_IS_TRUE(org_ge(arena, ORG_TAG_SMALL_INT(2), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_TRUE(org_ge(arena, ORG_TAG_SMALL_INT(3), ORG_TAG_SMALL_INT(2))));
  ASSERT(
      ORG_IS_FALSE(org_ge(arena, ORG_TAG_SMALL_INT(1), ORG_TAG_SMALL_INT(2))));
  PASS();
}

static void test_lt_bigint(void) {
  TEST("lt: BigInt comparison");
  OrgValue a = org_make_bigint_str(arena, "99999999999999999998");
  OrgValue b = org_make_bigint_str(arena, "99999999999999999999");
  ASSERT(ORG_IS_TRUE(org_lt(arena, a, b)));
  ASSERT(ORG_IS_FALSE(org_lt(arena, b, a)));
  PASS();
}

static void test_eq_rational(void) {
  TEST("eq: Rational(2/4) = Rational(1/2) (canonicalized)");
  OrgValue a = org_make_rational_str(arena, "2", "4");
  OrgValue b = org_make_rational_str(arena, "1", "2");
  ASSERT(ORG_IS_TRUE(org_eq(arena, a, b)));
  PASS();
}

static void test_lt_rational(void) {
  TEST("lt: Rational(1/3) < Rational(1/2)");
  OrgValue a = org_make_rational_str(arena, "1", "3");
  OrgValue b = org_make_rational_str(arena, "1", "2");
  ASSERT(ORG_IS_TRUE(org_lt(arena, a, b)));
  PASS();
}

static void test_lt_decimal(void) {
  TEST("lt: Decimal(1.5) < Decimal(2.5)");
  OrgValue a = org_make_decimal_str(arena, "1.5");
  OrgValue b = org_make_decimal_str(arena, "2.5");
  ASSERT(ORG_IS_TRUE(org_lt(arena, a, b)));
  PASS();
}

static void test_eq_cross_type(void) {
  TEST("eq: SmallInt(6) = Rational(6/1)");
  OrgValue a = ORG_TAG_SMALL_INT(6);
  OrgValue b = org_make_rational_str(arena, "6", "1");
  ASSERT(ORG_IS_TRUE(org_eq(arena, a, b)));
  PASS();
}

static void test_eq_decimal_int(void) {
  TEST("eq: Decimal(2.0) = SmallInt(2)");
  OrgValue a = org_make_decimal_str(arena, "2.0");
  OrgValue b = ORG_TAG_SMALL_INT(2);
  ASSERT(ORG_IS_TRUE(org_eq(arena, a, b)));
  PASS();
}

static void test_eq_non_numeric_identity(void) {
  TEST("eq: non-numeric identity comparison");
  OrgValue s1 = org_make_string(arena, "hi", 2);
  OrgValue s2 = org_make_string(arena, "hi", 2);
  /* Different pointers → not equal by identity */
  ASSERT(ORG_IS_FALSE(org_eq(arena, s1, s2)));
  /* Same value → equal by identity */
  ASSERT(ORG_IS_TRUE(org_eq(arena, s1, s1)));
  PASS();
}

static void test_ne_non_numeric_identity(void) {
  TEST("ne: non-numeric identity comparison");
  OrgValue s1 = org_make_string(arena, "hi", 2);
  OrgValue s2 = org_make_string(arena, "bye", 3);
  ASSERT(ORG_IS_TRUE(org_ne(arena, s1, s2)));
  ASSERT(ORG_IS_FALSE(org_ne(arena, s1, s1)));
  PASS();
}

static void test_lt_non_numeric_error(void) {
  TEST("lt/le/gt/ge: non-numeric → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_lt(arena, s, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_le(arena, s, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_gt(arena, s, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_ge(arena, s, ORG_TAG_SMALL_INT(1))));
  PASS();
}

static void test_cmp_error_propagation(void) {
  TEST("cmp: error propagation in all comparisons");
  ASSERT(ORG_IS_ERROR(org_eq(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_lt(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  ASSERT(ORG_IS_ERROR(org_le(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_gt(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  ASSERT(ORG_IS_ERROR(org_ge(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_ne(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  PASS();
}

/* ========== Error propagation ========== */

static void test_error_propagation(void) {
  TEST("error propagation through arithmetic");
  ASSERT(ORG_IS_ERROR(org_add(arena, ORG_ERROR, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_mul(arena, ORG_TAG_SMALL_INT(1), ORG_ERROR)));
  PASS();
}

static void test_add_non_numeric(void) {
  TEST("add: non-numeric → Error");
  OrgValue s = org_make_string(arena, "x", 1);
  ASSERT(ORG_IS_ERROR(org_add(arena, s, ORG_TAG_SMALL_INT(1))));
  ASSERT(ORG_IS_ERROR(org_add(arena, ORG_TAG_SMALL_INT(1), s)));
  PASS();
}

int main(void) {
  printf("=== Ops Tests ===\n");
  setup();

  /* SmallInt basic */
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
  test_sub_overflow();
  test_mul_overflow();
  test_bigint_add();
  test_bigint_normalize();
  test_normalize_non_bigint();

  /* Sub: all types */
  test_sub_bigint();
  test_sub_rational();
  test_sub_decimal();
  test_sub_int_decimal();
  test_sub_int_rational();
  test_sub_error();
  test_sub_non_numeric();

  /* Mul: all types */
  test_mul_bigint();
  test_mul_rational();
  test_mul_decimal();
  test_mul_int_rational();
  test_mul_int_decimal();
  test_mul_error();
  test_mul_non_numeric();

  /* Div: all types */
  test_div_bigint();
  test_div_bigint_inexact();
  test_div_bigint_zero();
  test_div_rational();
  test_div_decimal();
  test_div_int_decimal();
  test_div_rational_zero();
  test_div_decimal_zero();
  test_div_error();
  test_div_non_numeric();

  /* Mod: all types */
  test_mod_bigint();
  test_mod_zero();
  test_mod_rational_error();
  test_mod_decimal_error();
  test_mod_error();

  /* Neg: all types */
  test_neg_bigint();
  test_neg_rational();
  test_neg_decimal();
  test_neg_error();
  test_neg_non_numeric();

  /* Rational */
  test_rational_add();
  test_rational_to_integer();
  test_int_rational_promotion();

  /* Decimal */
  test_decimal_add();
  test_decimal_promotion();
  test_decimal_rational_promotion();

  /* Power */
  test_pow_small();
  test_pow_big();
  test_pow_negative_exp();
  test_pow_rational();
  test_pow_decimal();
  test_pow_zero();
  test_pow_bigint_base();
  test_pow_non_numeric_base();
  test_pow_non_int_exp();
  test_pow_error();

  /* Comparison */
  test_eq_small();
  test_ne_small();
  test_lt_small();
  test_le_small();
  test_gt_small();
  test_ge_small();
  test_lt_bigint();
  test_eq_rational();
  test_lt_rational();
  test_lt_decimal();
  test_eq_cross_type();
  test_eq_decimal_int();
  test_eq_non_numeric_identity();
  test_ne_non_numeric_identity();
  test_lt_non_numeric_error();
  test_cmp_error_propagation();

  /* Error */
  test_error_propagation();
  test_add_non_numeric();

  teardown();
  printf("\n%d/%d tests passed\n", tests_passed, tests_run);
  return tests_passed == tests_run ? 0 : 1;
}
