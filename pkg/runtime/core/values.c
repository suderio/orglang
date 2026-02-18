#include "values.h"
#include <stdlib.h>
#include <string.h>

/* ---- UTF-8 codepoint counting ---- */

static uint32_t count_codepoints(const char *data, size_t byte_len) {
  uint32_t count = 0;
  for (size_t i = 0; i < byte_len;) {
    uint8_t b = (uint8_t)data[i];
    if (b < 0x80)
      i += 1;
    else if ((b & 0xE0) == 0xC0)
      i += 2;
    else if ((b & 0xF0) == 0xE0)
      i += 3;
    else
      i += 4;
    count++;
  }
  return count;
}

/* ---- String ---- */

OrgValue org_make_string(Arena *arena, const char *str, size_t byte_len) {
  size_t total = sizeof(OrgString) + byte_len;
  OrgString *s = (OrgString *)arena_alloc(arena, total, 8);
  if (!s)
    return ORG_ERROR;

  s->header.type = ORG_TYPE_STRING;
  s->header.flags = 0;
  s->header._pad = 0;
  s->header.size = (uint32_t)total;
  s->byte_len = (uint32_t)byte_len;
  s->codepoint_len = count_codepoints(str, byte_len);
  memcpy(s->data, str, byte_len);

  return ORG_TAG_PTR_VAL(s);
}

const char *org_string_data(OrgValue v) {
  return ((OrgString *)ORG_GET_PTR(v))->data;
}

uint32_t org_string_byte_len(OrgValue v) {
  return ((OrgString *)ORG_GET_PTR(v))->byte_len;
}

uint32_t org_string_codepoint_len(OrgValue v) {
  return ((OrgString *)ORG_GET_PTR(v))->codepoint_len;
}

/* ---- BigInt ---- */

OrgValue org_make_bigint_str(Arena *arena, const char *str) {
  OrgBigInt *b = (OrgBigInt *)arena_alloc(arena, sizeof(OrgBigInt), 8);
  if (!b)
    return ORG_ERROR;

  b->header.type = ORG_TYPE_BIGINT;
  b->header.flags = 0;
  b->header._pad = 0;
  b->header.size = (uint32_t)sizeof(OrgBigInt);

  mpz_init(b->value);
  if (mpz_set_str(b->value, str, 10) != 0) {
    /* Invalid string — return error */
    return ORG_ERROR;
  }
  return ORG_TAG_PTR_VAL(b);
}

OrgValue org_make_bigint_si(Arena *arena, int64_t n) {
  OrgBigInt *b = (OrgBigInt *)arena_alloc(arena, sizeof(OrgBigInt), 8);
  if (!b)
    return ORG_ERROR;

  b->header.type = ORG_TYPE_BIGINT;
  b->header.flags = 0;
  b->header._pad = 0;
  b->header.size = (uint32_t)sizeof(OrgBigInt);

  mpz_init_set_si(b->value, (long)n);
  return ORG_TAG_PTR_VAL(b);
}

/* ---- Rational ---- */

OrgValue org_make_rational_str(Arena *arena, const char *num, const char *den) {
  OrgRational *r = (OrgRational *)arena_alloc(arena, sizeof(OrgRational), 8);
  if (!r)
    return ORG_ERROR;

  r->header.type = ORG_TYPE_RATIONAL;
  r->header.flags = 0;
  r->header._pad = 0;
  r->header.size = (uint32_t)sizeof(OrgRational);

  mpq_init(r->value);
  mpz_set_str(mpq_numref(r->value), num, 10);
  mpz_set_str(mpq_denref(r->value), den, 10);
  mpq_canonicalize(r->value);
  return ORG_TAG_PTR_VAL(r);
}

OrgValue org_make_rational_mpz(Arena *arena, const mpz_t num, const mpz_t den) {
  OrgRational *r = (OrgRational *)arena_alloc(arena, sizeof(OrgRational), 8);
  if (!r)
    return ORG_ERROR;

  r->header.type = ORG_TYPE_RATIONAL;
  r->header.flags = 0;
  r->header._pad = 0;
  r->header.size = (uint32_t)sizeof(OrgRational);

  mpq_init(r->value);
  mpz_set(mpq_numref(r->value), num);
  mpz_set(mpq_denref(r->value), den);
  mpq_canonicalize(r->value);
  return ORG_TAG_PTR_VAL(r);
}

/* ---- Decimal ---- */

OrgValue org_make_decimal_str(Arena *arena, const char *str) {
  OrgDecimal *d = (OrgDecimal *)arena_alloc(arena, sizeof(OrgDecimal), 8);
  if (!d)
    return ORG_ERROR;

  d->header.type = ORG_TYPE_DECIMAL;
  d->header.flags = 0;
  d->header._pad = 0;
  d->header.size = (uint32_t)sizeof(OrgDecimal);
  d->_pad2 = 0;

  /* Find the decimal point to determine scale */
  const char *dot = strchr(str, '.');
  if (!dot) {
    /* No decimal point: scale = 0, value = str/1 */
    d->scale = 0;
    mpq_init(d->value);
    mpz_set_str(mpq_numref(d->value), str, 10);
    mpz_set_ui(mpq_denref(d->value), 1);
  } else {
    d->scale = (int32_t)(strlen(dot + 1));

    /* Build numerator by removing the dot */
    size_t len = strlen(str);
    char *num_str =
        (char *)arena_alloc(arena, len, 1); /* temp, no alignment needed */
    size_t before_dot = (size_t)(dot - str);
    memcpy(num_str, str, before_dot);
    memcpy(num_str + before_dot, dot + 1, len - before_dot - 1);
    num_str[len - 1] = '\0';

    mpq_init(d->value);
    mpz_set_str(mpq_numref(d->value), num_str, 10);

    /* Denominator = 10^scale */
    mpz_t denom;
    mpz_init(denom);
    mpz_ui_pow_ui(denom, 10, (unsigned long)d->scale);
    mpz_set(mpq_denref(d->value), denom);
    mpz_clear(denom); /* clear the temp (arena makes this a no-op anyway) */
  }
  mpq_canonicalize(d->value);
  return ORG_TAG_PTR_VAL(d);
}

/* ---- Type name ---- */

const char *org_type_name(OrgValue v) {
  if (ORG_IS_SMALL(v))
    return "SmallInt";
  if (ORG_IS_TRUE(v))
    return "Boolean(true)";
  if (ORG_IS_FALSE(v))
    return "Boolean(false)";
  if (ORG_IS_ERROR(v))
    return "Error";
  if (ORG_IS_UNUSED(v))
    return "Unused";

  if (ORG_IS_PTR(v)) {
    switch (org_get_type(v)) {
    case ORG_TYPE_BIGINT:
      return "BigInt";
    case ORG_TYPE_RATIONAL:
      return "Rational";
    case ORG_TYPE_DECIMAL:
      return "Decimal";
    case ORG_TYPE_STRING:
      return "String";
    case ORG_TYPE_TABLE:
      return "Table";
    case ORG_TYPE_CLOSURE:
      return "Closure";
    case ORG_TYPE_RESOURCE:
      return "Resource";
    case ORG_TYPE_ERROR_OBJ:
      return "ErrorObj";
    }
  }
  return "Unknown";
}
