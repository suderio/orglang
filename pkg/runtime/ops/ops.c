#include "ops.h"
#include <string.h>

/*
 * Internal helpers to convert any numeric value to mpz/mpq
 * for the slow-path GMP operations.
 */

/* Numeric category for promotion dispatch */
typedef enum {
  NUM_SMALL,
  NUM_BIGINT,
  NUM_RATIONAL,
  NUM_DECIMAL,
  NUM_NONE,
} NumCat;

static NumCat num_category(OrgValue v) {
  if (ORG_IS_SMALL(v))
    return NUM_SMALL;
  if (!ORG_IS_PTR(v))
    return NUM_NONE;
  switch (org_get_type(v)) {
  case ORG_TYPE_BIGINT:
    return NUM_BIGINT;
  case ORG_TYPE_RATIONAL:
    return NUM_RATIONAL;
  case ORG_TYPE_DECIMAL:
    return NUM_DECIMAL;
  default:
    return NUM_NONE;
  }
}

/* Convert any integer (small or big) to mpz_t. Caller must mpz_clear. */
static void to_mpz(OrgValue v, mpz_t out) {
  if (ORG_IS_SMALL(v)) {
    mpz_init_set_si(out, (long)ORG_UNTAG_SMALL_INT(v));
  } else {
    mpz_init_set(out, *org_get_bigint(v));
  }
}

/* Convert any numeric to mpq_t. Caller must mpq_clear. */
static void to_mpq(OrgValue v, mpq_t out) {
  mpq_init(out);
  if (ORG_IS_SMALL(v)) {
    mpq_set_si(out, (long)ORG_UNTAG_SMALL_INT(v), 1);
  } else {
    switch (org_get_type(v)) {
    case ORG_TYPE_BIGINT:
      mpq_set_z(out, *org_get_bigint(v));
      break;
    case ORG_TYPE_RATIONAL:
      mpq_set(out, *org_get_rational(v));
      break;
    case ORG_TYPE_DECIMAL:
      mpq_set(out, *org_get_decimal(v));
      break;
    default:
      /* Should not happen — caller checked category */
      break;
    }
  }
}

/*
 * Try to normalize a BigInt back to SmallInt if it fits.
 */
OrgValue org_normalize_int(OrgValue v) {
  if (!ORG_IS_PTR(v) || org_get_type(v) != ORG_TYPE_BIGINT)
    return v;
  mpz_t *z = org_get_bigint(v);
  if (mpz_fits_slong_p(*z)) {
    long val = mpz_get_si(*z);
    if (org_small_fits((int64_t)val)) {
      return ORG_TAG_SMALL_INT(val);
    }
  }
  return v;
}

/* Wrap an mpz result: try SmallInt first, otherwise BigInt. */
static OrgValue wrap_mpz(Arena *arena, const mpz_t z) {
  if (mpz_fits_slong_p(z)) {
    long val = mpz_get_si(z);
    if (org_small_fits((int64_t)val)) {
      return ORG_TAG_SMALL_INT(val);
    }
  }
  OrgBigInt *b = (OrgBigInt *)arena_alloc(arena, sizeof(OrgBigInt), 8);
  if (!b)
    return ORG_ERROR;
  b->header.type = ORG_TYPE_BIGINT;
  b->header.flags = 0;
  b->header._pad = 0;
  b->header.size = (uint32_t)sizeof(OrgBigInt);
  mpz_init_set(b->value, z);
  return ORG_TAG_PTR_VAL(b);
}

/* Wrap an mpq result as Rational. */
static OrgValue wrap_mpq_rational(Arena *arena, const mpq_t q) {
  /* If denominator is 1, return as integer */
  if (mpz_cmp_ui(mpq_denref(q), 1) == 0) {
    return wrap_mpz(arena, mpq_numref(q));
  }
  OrgRational *r = (OrgRational *)arena_alloc(arena, sizeof(OrgRational), 8);
  if (!r)
    return ORG_ERROR;
  r->header.type = ORG_TYPE_RATIONAL;
  r->header.flags = 0;
  r->header._pad = 0;
  r->header.size = (uint32_t)sizeof(OrgRational);
  mpq_init(r->value);
  mpq_set(r->value, q);
  return ORG_TAG_PTR_VAL(r);
}

/* Wrap an mpq result as Decimal (preserving scale from context). */
static OrgValue wrap_mpq_decimal(Arena *arena, const mpq_t q, int32_t scale) {
  OrgDecimal *d = (OrgDecimal *)arena_alloc(arena, sizeof(OrgDecimal), 8);
  if (!d)
    return ORG_ERROR;
  d->header.type = ORG_TYPE_DECIMAL;
  d->header.flags = 0;
  d->header._pad = 0;
  d->header.size = (uint32_t)sizeof(OrgDecimal);
  d->scale = scale;
  d->_pad2 = 0;
  mpq_init(d->value);
  mpq_set(d->value, q);
  return ORG_TAG_PTR_VAL(d);
}

/* Get scale from a value (only meaningful for decimals) */
static int32_t get_scale(OrgValue v) {
  if (ORG_IS_PTR(v) && org_get_type(v) == ORG_TYPE_DECIMAL) {
    return org_get_decimal_scale(v);
  }
  return 0;
}

/* ========== Arithmetic Operations ========== */

OrgValue org_add(Arena *arena, OrgValue a, OrgValue b) {
  /* Fast path: both small integers */
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t sb = ORG_UNTAG_SMALL_INT(b);
    int64_t result;
    if (!__builtin_add_overflow(sa, sb, &result) && org_small_fits(result)) {
      return ORG_TAG_SMALL_INT(result);
    }
    /* Overflow → BigInt */
    mpz_t za, zb, zr;
    mpz_init_set_si(za, (long)sa);
    mpz_init_set_si(zb, (long)sb);
    mpz_init(zr);
    mpz_add(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  /* Error propagation */
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;

  NumCat ca = num_category(a), cb = num_category(b);
  if (ca == NUM_NONE || cb == NUM_NONE)
    return ORG_ERROR;

  /* Integer + Integer (at least one BigInt) */
  if ((ca == NUM_SMALL || ca == NUM_BIGINT) &&
      (cb == NUM_SMALL || cb == NUM_BIGINT)) {
    mpz_t za, zb, zr;
    to_mpz(a, za);
    to_mpz(b, zb);
    mpz_init(zr);
    mpz_add(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  /* Decimal involved → result is Decimal */
  if (ca == NUM_DECIMAL || cb == NUM_DECIMAL) {
    mpq_t qa, qb, qr;
    to_mpq(a, qa);
    to_mpq(b, qb);
    mpq_init(qr);
    mpq_add(qr, qa, qb);
    int32_t scale = get_scale(a);
    int32_t sb_scale = get_scale(b);
    if (sb_scale > scale)
      scale = sb_scale;
    OrgValue rv = wrap_mpq_decimal(arena, qr, scale);
    mpq_clear(qa);
    mpq_clear(qb);
    mpq_clear(qr);
    return rv;
  }

  /* Rational path */
  mpq_t qa, qb, qr;
  to_mpq(a, qa);
  to_mpq(b, qb);
  mpq_init(qr);
  mpq_add(qr, qa, qb);
  OrgValue rv = wrap_mpq_rational(arena, qr);
  mpq_clear(qa);
  mpq_clear(qb);
  mpq_clear(qr);
  return rv;
}

OrgValue org_sub(Arena *arena, OrgValue a, OrgValue b) {
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t sb = ORG_UNTAG_SMALL_INT(b);
    int64_t result;
    if (!__builtin_sub_overflow(sa, sb, &result) && org_small_fits(result)) {
      return ORG_TAG_SMALL_INT(result);
    }
    mpz_t za, zb, zr;
    mpz_init_set_si(za, (long)sa);
    mpz_init_set_si(zb, (long)sb);
    mpz_init(zr);
    mpz_sub(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;

  NumCat ca = num_category(a), cb = num_category(b);
  if (ca == NUM_NONE || cb == NUM_NONE)
    return ORG_ERROR;

  if ((ca == NUM_SMALL || ca == NUM_BIGINT) &&
      (cb == NUM_SMALL || cb == NUM_BIGINT)) {
    mpz_t za, zb, zr;
    to_mpz(a, za);
    to_mpz(b, zb);
    mpz_init(zr);
    mpz_sub(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  if (ca == NUM_DECIMAL || cb == NUM_DECIMAL) {
    mpq_t qa, qb, qr;
    to_mpq(a, qa);
    to_mpq(b, qb);
    mpq_init(qr);
    mpq_sub(qr, qa, qb);
    int32_t scale = get_scale(a);
    int32_t sb_scale = get_scale(b);
    if (sb_scale > scale)
      scale = sb_scale;
    OrgValue rv = wrap_mpq_decimal(arena, qr, scale);
    mpq_clear(qa);
    mpq_clear(qb);
    mpq_clear(qr);
    return rv;
  }

  mpq_t qa, qb, qr;
  to_mpq(a, qa);
  to_mpq(b, qb);
  mpq_init(qr);
  mpq_sub(qr, qa, qb);
  OrgValue rv = wrap_mpq_rational(arena, qr);
  mpq_clear(qa);
  mpq_clear(qb);
  mpq_clear(qr);
  return rv;
}

OrgValue org_mul(Arena *arena, OrgValue a, OrgValue b) {
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t sb = ORG_UNTAG_SMALL_INT(b);
    int64_t result;
    if (!__builtin_mul_overflow(sa, sb, &result) && org_small_fits(result)) {
      return ORG_TAG_SMALL_INT(result);
    }
    mpz_t za, zb, zr;
    mpz_init_set_si(za, (long)sa);
    mpz_init_set_si(zb, (long)sb);
    mpz_init(zr);
    mpz_mul(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;

  NumCat ca = num_category(a), cb = num_category(b);
  if (ca == NUM_NONE || cb == NUM_NONE)
    return ORG_ERROR;

  if ((ca == NUM_SMALL || ca == NUM_BIGINT) &&
      (cb == NUM_SMALL || cb == NUM_BIGINT)) {
    mpz_t za, zb, zr;
    to_mpz(a, za);
    to_mpz(b, zb);
    mpz_init(zr);
    mpz_mul(zr, za, zb);
    OrgValue rv = wrap_mpz(arena, zr);
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(zr);
    return rv;
  }

  if (ca == NUM_DECIMAL || cb == NUM_DECIMAL) {
    mpq_t qa, qb, qr;
    to_mpq(a, qa);
    to_mpq(b, qb);
    mpq_init(qr);
    mpq_mul(qr, qa, qb);
    int32_t scale = get_scale(a) + get_scale(b);
    OrgValue rv = wrap_mpq_decimal(arena, qr, scale);
    mpq_clear(qa);
    mpq_clear(qb);
    mpq_clear(qr);
    return rv;
  }

  mpq_t qa, qb, qr;
  to_mpq(a, qa);
  to_mpq(b, qb);
  mpq_init(qr);
  mpq_mul(qr, qa, qb);
  OrgValue rv = wrap_mpq_rational(arena, qr);
  mpq_clear(qa);
  mpq_clear(qb);
  mpq_clear(qr);
  return rv;
}

/*
 * Division:
 * - Integer / Integer → Integer if exact, Rational if not
 * - Any Decimal involved → Decimal
 * - Otherwise → Rational
 */
OrgValue org_div(Arena *arena, OrgValue a, OrgValue b) {
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;

  /* Division by zero check */
  if (ORG_IS_SMALL(b) && ORG_UNTAG_SMALL_INT(b) == 0)
    return ORG_ERROR;
  if (ORG_IS_PTR(b) && org_get_type(b) == ORG_TYPE_BIGINT &&
      mpz_sgn(*org_get_bigint(b)) == 0)
    return ORG_ERROR;

  NumCat ca = num_category(a), cb = num_category(b);
  if (ca == NUM_NONE || cb == NUM_NONE)
    return ORG_ERROR;

  /* Integer / Integer → try exact, else Rational */
  if ((ca == NUM_SMALL || ca == NUM_BIGINT) &&
      (cb == NUM_SMALL || cb == NUM_BIGINT)) {
    mpz_t za, zb, quo, rem;
    to_mpz(a, za);
    to_mpz(b, zb);
    mpz_init(quo);
    mpz_init(rem);
    mpz_tdiv_qr(quo, rem, za, zb);
    OrgValue rv;
    if (mpz_sgn(rem) == 0) {
      /* Exact division → Integer */
      rv = wrap_mpz(arena, quo);
    } else {
      /* Inexact → Rational */
      mpq_t q;
      mpq_init(q);
      mpq_set_z(q, za);
      mpq_t denom;
      mpq_init(denom);
      mpq_set_z(denom, zb);
      mpq_div(q, q, denom);
      mpq_canonicalize(q);
      rv = wrap_mpq_rational(arena, q);
      mpq_clear(q);
      mpq_clear(denom);
    }
    mpz_clear(za);
    mpz_clear(zb);
    mpz_clear(quo);
    mpz_clear(rem);
    return rv;
  }

  /* Decimal involved */
  if (ca == NUM_DECIMAL || cb == NUM_DECIMAL) {
    mpq_t qa, qb, qr;
    to_mpq(a, qa);
    to_mpq(b, qb);
    if (mpq_sgn(qb) == 0) {
      mpq_clear(qa);
      mpq_clear(qb);
      return ORG_ERROR;
    }
    mpq_init(qr);
    mpq_div(qr, qa, qb);
    int32_t scale = get_scale(a);
    if (scale == 0)
      scale = get_scale(b);
    if (scale == 0)
      scale = 1; /* Default scale for decimal division */
    OrgValue rv = wrap_mpq_decimal(arena, qr, scale);
    mpq_clear(qa);
    mpq_clear(qb);
    mpq_clear(qr);
    return rv;
  }

  /* Rational path */
  mpq_t qa, qb, qr;
  to_mpq(a, qa);
  to_mpq(b, qb);
  if (mpq_sgn(qb) == 0) {
    mpq_clear(qa);
    mpq_clear(qb);
    return ORG_ERROR;
  }
  mpq_init(qr);
  mpq_div(qr, qa, qb);
  OrgValue rv = wrap_mpq_rational(arena, qr);
  mpq_clear(qa);
  mpq_clear(qb);
  mpq_clear(qr);
  return rv;
}

OrgValue org_mod(Arena *arena, OrgValue a, OrgValue b) {
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;

  /* Modulo only defined for integers */
  NumCat ca = num_category(a), cb = num_category(b);
  if (!((ca == NUM_SMALL || ca == NUM_BIGINT) &&
        (cb == NUM_SMALL || cb == NUM_BIGINT))) {
    return ORG_ERROR;
  }

  /* Division by zero */
  if (ORG_IS_SMALL(b) && ORG_UNTAG_SMALL_INT(b) == 0)
    return ORG_ERROR;

  /* Fast path */
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t sb = ORG_UNTAG_SMALL_INT(b);
    return ORG_TAG_SMALL_INT(sa % sb);
  }

  mpz_t za, zb, zr;
  to_mpz(a, za);
  to_mpz(b, zb);
  mpz_init(zr);
  mpz_mod(zr, za, zb);
  OrgValue rv = wrap_mpz(arena, zr);
  mpz_clear(za);
  mpz_clear(zb);
  mpz_clear(zr);
  return rv;
}

OrgValue org_neg(Arena *arena, OrgValue a) {
  if (ORG_IS_ERROR(a))
    return ORG_ERROR;

  if (ORG_IS_SMALL(a)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t result;
    if (!__builtin_sub_overflow((int64_t)0, sa, &result) &&
        org_small_fits(result)) {
      return ORG_TAG_SMALL_INT(result);
    }
    mpz_t z;
    mpz_init_set_si(z, (long)sa);
    mpz_neg(z, z);
    OrgValue rv = wrap_mpz(arena, z);
    mpz_clear(z);
    return rv;
  }

  NumCat c = num_category(a);
  if (c == NUM_NONE)
    return ORG_ERROR;

  if (c == NUM_BIGINT) {
    mpz_t z;
    to_mpz(a, z);
    mpz_neg(z, z);
    OrgValue rv = wrap_mpz(arena, z);
    mpz_clear(z);
    return rv;
  }

  if (c == NUM_DECIMAL) {
    mpq_t q;
    to_mpq(a, q);
    mpq_neg(q, q);
    OrgValue rv = wrap_mpq_decimal(arena, q, get_scale(a));
    mpq_clear(q);
    return rv;
  }

  /* Rational */
  mpq_t q;
  to_mpq(a, q);
  mpq_neg(q, q);
  OrgValue rv = wrap_mpq_rational(arena, q);
  mpq_clear(q);
  return rv;
}

OrgValue org_pow(Arena *arena, OrgValue base, OrgValue exp) {
  if (ORG_IS_ERROR(base) || ORG_IS_ERROR(exp))
    return ORG_ERROR;

  /* Exponent must be a non-negative integer */
  if (!org_is_integer(exp))
    return ORG_ERROR;

  unsigned long e;
  if (ORG_IS_SMALL(exp)) {
    int64_t se = ORG_UNTAG_SMALL_INT(exp);
    if (se < 0)
      return ORG_ERROR;
    e = (unsigned long)se;
  } else {
    if (mpz_sgn(*org_get_bigint(exp)) < 0)
      return ORG_ERROR;
    if (!mpz_fits_ulong_p(*org_get_bigint(exp)))
      return ORG_ERROR;
    e = mpz_get_ui(*org_get_bigint(exp));
  }

  NumCat cb = num_category(base);
  if (cb == NUM_NONE)
    return ORG_ERROR;

  if (cb == NUM_SMALL || cb == NUM_BIGINT) {
    mpz_t z;
    to_mpz(base, z);
    mpz_pow_ui(z, z, e);
    OrgValue rv = wrap_mpz(arena, z);
    mpz_clear(z);
    return rv;
  }

  /* Rational or Decimal: (p/q)^n = p^n / q^n */
  mpq_t q;
  to_mpq(base, q);
  mpz_pow_ui(mpq_numref(q), mpq_numref(q), e);
  mpz_pow_ui(mpq_denref(q), mpq_denref(q), e);
  mpq_canonicalize(q);

  OrgValue rv;
  if (cb == NUM_DECIMAL) {
    rv = wrap_mpq_decimal(arena, q, get_scale(base) * (int32_t)e);
  } else {
    rv = wrap_mpq_rational(arena, q);
  }
  mpq_clear(q);
  return rv;
}

/* ========== Comparison Operations ========== */

/* Internal: compare two numeric values. Returns -1, 0, or 1. */
static int org_cmp_internal(OrgValue a, OrgValue b) {
  /* Both small */
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b)) {
    int64_t sa = ORG_UNTAG_SMALL_INT(a);
    int64_t sb = ORG_UNTAG_SMALL_INT(b);
    return (sa > sb) - (sa < sb);
  }

  /* Convert to rationals for universal comparison */
  mpq_t qa, qb;
  to_mpq(a, qa);
  to_mpq(b, qb);
  int result = mpq_cmp(qa, qb);
  mpq_clear(qa);
  mpq_clear(qb);
  return (result > 0) - (result < 0);
}

OrgValue org_eq(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b)) {
    return ORG_BOOL(a == b); /* Identity comparison for non-numerics */
  }
  return ORG_BOOL(org_cmp_internal(a, b) == 0);
}

OrgValue org_lt(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b))
    return ORG_ERROR;
  return ORG_BOOL(org_cmp_internal(a, b) < 0);
}

OrgValue org_le(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b))
    return ORG_ERROR;
  return ORG_BOOL(org_cmp_internal(a, b) <= 0);
}

OrgValue org_gt(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b))
    return ORG_ERROR;
  return ORG_BOOL(org_cmp_internal(a, b) > 0);
}

OrgValue org_ge(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b))
    return ORG_ERROR;
  return ORG_BOOL(org_cmp_internal(a, b) >= 0);
}

OrgValue org_ne(Arena *arena, OrgValue a, OrgValue b) {
  (void)arena;
  if (ORG_IS_ERROR(a) || ORG_IS_ERROR(b))
    return ORG_ERROR;
  if (!org_is_numeric(a) || !org_is_numeric(b)) {
    return ORG_BOOL(a != b);
  }
  return ORG_BOOL(org_cmp_internal(a, b) != 0);
}
