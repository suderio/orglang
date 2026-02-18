#ifndef ORG_VALUES_H
#define ORG_VALUES_H

#include "arena.h"
#include <gmp.h>
#include <stddef.h>
#include <stdint.h>

/*
 * OrgValue — Tagged 64-bit value.
 *
 * Lower 2 bits encode the type:
 *   00 → Pointer to heap object (Arena-allocated, 8-byte aligned)
 *   01 → Small Integer (62-bit signed, shifted left by 2)
 *   10 → Special: Boolean (True/False), Error, Unused
 *   11 → Reserved
 */
typedef uint64_t OrgValue;

/* ---- Tag checks ---- */
#define ORG_TAG_MASK 3
#define ORG_TAG_PTR 0
#define ORG_TAG_SMALL 1
#define ORG_TAG_SPECIAL 2
#define ORG_TAG_RESERVED 3

#define ORG_IS_PTR(v) (((v) & ORG_TAG_MASK) == ORG_TAG_PTR)
#define ORG_IS_SMALL(v) (((v) & ORG_TAG_MASK) == ORG_TAG_SMALL)
#define ORG_IS_SPECIAL(v) (((v) & ORG_TAG_MASK) == ORG_TAG_SPECIAL)

/* ---- Small Integers (62-bit signed) ---- */
#define ORG_SMALL_MAX ((int64_t)((UINT64_C(1) << 61) - 1))
#define ORG_SMALL_MIN (-(int64_t)(UINT64_C(1) << 61))

#define ORG_TAG_SMALL_INT(n)                                                   \
  ((OrgValue)(((uint64_t)(int64_t)(n) << 2) | ORG_TAG_SMALL))
#define ORG_UNTAG_SMALL_INT(v)                                                 \
  ((int64_t)(v) >> 2) /* arithmetic shift preserves sign */

static inline int org_small_fits(int64_t n) {
  return n >= ORG_SMALL_MIN && n <= ORG_SMALL_MAX;
}

/* ---- Special values ---- */
#define ORG_TRUE ((OrgValue)0x06)   /* 0b...0110 */
#define ORG_FALSE ((OrgValue)0x02)  /* 0b...0010 */
#define ORG_ERROR ((OrgValue)0x0A)  /* 0b...1010 */
#define ORG_UNUSED ((OrgValue)0x0E) /* 0b...1110 (internal: absent operand) */

#define ORG_IS_TRUE(v) ((v) == ORG_TRUE)
#define ORG_IS_FALSE(v) ((v) == ORG_FALSE)
#define ORG_IS_ERROR(v) ((v) == ORG_ERROR)
#define ORG_IS_UNUSED(v) ((v) == ORG_UNUSED)
#define ORG_IS_BOOL(v) (ORG_IS_TRUE(v) || ORG_IS_FALSE(v))

/* Boolean from C int (0 = false, nonzero = true) */
#define ORG_BOOL(cond) ((cond) ? ORG_TRUE : ORG_FALSE)

/* ---- Heap object header ---- */
typedef enum OrgType {
  ORG_TYPE_BIGINT,
  ORG_TYPE_RATIONAL,
  ORG_TYPE_DECIMAL,
  ORG_TYPE_STRING,
  ORG_TYPE_TABLE,
  ORG_TYPE_CLOSURE,
  ORG_TYPE_RESOURCE,
  ORG_TYPE_ERROR_OBJ,
} OrgType;

/*
 * Every Arena-allocated object starts with this header.
 * The header is 8 bytes, keeping subsequent fields aligned.
 */
typedef struct OrgObject {
  uint8_t type;  /* OrgType enum */
  uint8_t flags; /* Reserved for future use */
  uint16_t _pad;
  uint32_t size; /* Total object size in bytes (including header) */
} OrgObject;

/* Extract pointer from a tagged value (caller must check ORG_IS_PTR first) */
#define ORG_GET_PTR(v) ((OrgObject *)(uintptr_t)(v))

/* Create a tagged pointer from an OrgObject* */
#define ORG_TAG_PTR_VAL(ptr) ((OrgValue)(uintptr_t)(ptr))

/* Get the OrgType from a pointer-tagged value */
static inline OrgType org_get_type(OrgValue v) {
  return (OrgType)ORG_GET_PTR(v)->type;
}

/* ---- Numeric type checks ---- */

static inline int org_is_integer(OrgValue v) {
  return ORG_IS_SMALL(v) ||
         (ORG_IS_PTR(v) && org_get_type(v) == ORG_TYPE_BIGINT);
}

static inline int org_is_rational(OrgValue v) {
  return ORG_IS_PTR(v) && org_get_type(v) == ORG_TYPE_RATIONAL;
}

static inline int org_is_decimal(OrgValue v) {
  return ORG_IS_PTR(v) && org_get_type(v) == ORG_TYPE_DECIMAL;
}

static inline int org_is_numeric(OrgValue v) {
  return org_is_integer(v) || org_is_rational(v) || org_is_decimal(v);
}

/* ---- BigInt representation ---- */
typedef struct OrgBigInt {
  OrgObject header;
  mpz_t value;
} OrgBigInt;

/* Create BigInt from decimal string (e.g., "12345678901234567890") */
OrgValue org_make_bigint_str(Arena *arena, const char *str);

/* Create BigInt from a C int64 */
OrgValue org_make_bigint_si(Arena *arena, int64_t n);

/* Get the mpz_t from a BigInt value */
static inline mpz_t *org_get_bigint(OrgValue v) {
  return &((OrgBigInt *)ORG_GET_PTR(v))->value;
}

/* ---- Rational representation ---- */
typedef struct OrgRational {
  OrgObject header;
  mpq_t value;
} OrgRational;

/* Create Rational from numerator/denominator strings (auto-canonicalizes) */
OrgValue org_make_rational_str(Arena *arena, const char *num, const char *den);

/* Create Rational from two mpz_t values */
OrgValue org_make_rational_mpz(Arena *arena, const mpz_t num, const mpz_t den);

/* Get the mpq_t from a Rational value */
static inline mpq_t *org_get_rational(OrgValue v) {
  return &((OrgRational *)ORG_GET_PTR(v))->value;
}

/* ---- Decimal representation (rational + display scale) ---- */
typedef struct OrgDecimal {
  OrgObject header;
  mpq_t value;   /* Exact rational representation */
  int32_t scale; /* Digits after decimal point (display hint) */
  int32_t _pad2;
} OrgDecimal;

/* Create Decimal from string (e.g., "3.14" → 314/100, scale=2) */
OrgValue org_make_decimal_str(Arena *arena, const char *str);

/* Get the mpq_t from a Decimal value */
static inline mpq_t *org_get_decimal(OrgValue v) {
  return &((OrgDecimal *)ORG_GET_PTR(v))->value;
}

static inline int32_t org_get_decimal_scale(OrgValue v) {
  return ((OrgDecimal *)ORG_GET_PTR(v))->scale;
}

/* ---- String representation ---- */
typedef struct OrgString {
  OrgObject header;
  uint32_t byte_len;      /* Length in bytes (not codepoints) */
  uint32_t codepoint_len; /* Length in Unicode codepoints */
  char data[];            /* UTF-8 encoded, NOT null-terminated */
} OrgString;

OrgValue org_make_string(Arena *arena, const char *str, size_t byte_len);
const char *org_string_data(OrgValue v);
uint32_t org_string_byte_len(OrgValue v);
uint32_t org_string_codepoint_len(OrgValue v);

/* ---- Type query ---- */
const char *org_type_name(OrgValue v);

#endif /* ORG_VALUES_H */
