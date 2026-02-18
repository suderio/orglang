#include "values.h"
#include <string.h>

/*
 * Count Unicode codepoints in a UTF-8 byte sequence.
 * Continuation bytes (10xxxxxx) are not counted.
 */
static uint32_t count_codepoints(const char *data, size_t byte_len) {
  uint32_t count = 0;
  for (size_t i = 0; i < byte_len;) {
    uint8_t b = (uint8_t)data[i];
    if (b < 0x80) {
      i += 1;
    } else if ((b & 0xE0) == 0xC0) {
      i += 2;
    } else if ((b & 0xF0) == 0xE0) {
      i += 3;
    } else {
      i += 4;
    }
    count++;
  }
  return count;
}

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
  OrgString *s = (OrgString *)ORG_GET_PTR(v);
  return s->data;
}

uint32_t org_string_byte_len(OrgValue v) {
  OrgString *s = (OrgString *)ORG_GET_PTR(v);
  return s->byte_len;
}

uint32_t org_string_codepoint_len(OrgValue v) {
  OrgString *s = (OrgString *)ORG_GET_PTR(v);
  return s->codepoint_len;
}

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
