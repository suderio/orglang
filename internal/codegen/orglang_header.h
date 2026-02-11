#ifndef ORGLANG_H
#define ORGLANG_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

// Tiny Arena Allocator
typedef struct Arena {
  char *data;
  size_t size;
  size_t offset;
} Arena;

static Arena *arena_create(size_t size) {
  Arena *a = (Arena *)malloc(sizeof(Arena));
  a->data = (char *)malloc(size);
  a->size = size;
  a->offset = 0;
  return a;
}

static void arena_free(Arena *a) {
  free(a->data);
  free(a);
}

static void *arena_alloc(Arena *a, size_t size) {
  if (a->offset + size > a->size) {
    fprintf(stderr, "Arena out of memory!\n");
    exit(1);
  }
  void *ptr = a->data + a->offset;
  a->offset += size;
  return ptr;
}

// OrgLang Types
typedef enum { ORG_INT_TYPE, ORG_DEC_TYPE, ORG_STR_TYPE } OrgType;

typedef struct OrgValue {
  OrgType type;
  char *str_val; // For now, just store string representation
} OrgValue;

static OrgValue *org_int_from_str(Arena *a, const char *s) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_INT_TYPE;
  size_t len = strlen(s) + 1;
  v->str_val = (char *)arena_alloc(a, len);
  strcpy(v->str_val, s);
  return v;
}

static OrgValue *org_dec_from_str(Arena *a, const char *s) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_DEC_TYPE;
  size_t len = strlen(s) + 1;
  v->str_val = (char *)arena_alloc(a, len);
  strcpy(v->str_val, s);
  return v;
}

static OrgValue *org_string_from_c(Arena *a, const char *s) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_STR_TYPE;
  size_t len = strlen(s) + 1;
  v->str_val = (char *)arena_alloc(a, len);
  strcpy(v->str_val, s);
  return v;
}

static OrgValue *org_bool(Arena *a, int val) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_INT_TYPE; // Reuse int for now

  const char *boolStr = val ? "true" : "false";
  size_t len = strlen(boolStr) + 1;
  v->str_val = (char *)arena_alloc(a, len);
  strcpy(v->str_val, boolStr);
  return v;
}

// Operators (Mock implementation)

static OrgValue *org_op_infix(Arena *a, const char *op, OrgValue *left,
                              OrgValue *right) {
  // Determine type, perform op (using GMP later)
  // For now, just return left for testing connectivity
  printf("Debug: %s %s %s\n", left->str_val, op, right->str_val);
  return left;
}

static OrgValue *org_print(Arena *a, OrgValue *v) {
  if (v->type == ORG_STR_TYPE) {
    printf("%s\n", v->str_val);
  } else {
    printf("%s\n", v->str_val); // All have str_val for now
  }
  return v;
}

static OrgValue *org_op_prefix(Arena *a, const char *op, OrgValue *right) {
  printf("Debug: %s %s\n", op, right->str_val);
  return right;
}

#endif
