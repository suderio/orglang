#ifndef ORGLANG_H
#define ORGLANG_H

#include <stdarg.h>
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
typedef enum {
  ORG_INT_TYPE,
  ORG_DEC_TYPE,
  ORG_STR_TYPE,
  ORG_LIST_TYPE
} OrgType;

typedef struct OrgValue OrgValue; // Forward declaration

typedef struct OrgList {
  OrgValue **items;
  size_t capacity;
  size_t size;
} OrgList;

struct OrgValue {
  OrgType type;
  char *str_val;     // For INT, DEC, STR
  OrgList *list_val; // For LIST
};

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

static OrgValue *org_list_create(Arena *a, size_t capacity) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_LIST_TYPE;

  OrgList *l = (OrgList *)arena_alloc(a, sizeof(OrgList));
  l->capacity = capacity > 0 ? capacity : 4;
  l->size = 0;
  l->items = (OrgValue **)arena_alloc(a, sizeof(OrgValue *) * l->capacity);

  v->list_val = l;
  return v;
}

static void org_list_append(Arena *a, OrgValue *list, OrgValue *item) {
  if (list->type != ORG_LIST_TYPE)
    return;
  OrgList *l = list->list_val;
  if (l->size >= l->capacity) {
    // Reallocate (simple expand in Arena? Arena doesn't support realloc easily
    // without waste) For prototype, we just alloc new larger array and copy.
    size_t new_cap = l->capacity * 2;
    OrgValue **new_items =
        (OrgValue **)arena_alloc(a, sizeof(OrgValue *) * new_cap);
    for (size_t i = 0; i < l->size; i++)
      new_items[i] = l->items[i];
    l->items = new_items;
    l->capacity = new_cap;
  }
  l->items[l->size++] = item;
}

static OrgValue *org_list_make(Arena *a, int count, ...) {
  OrgValue *v = org_list_create(a, count);

  va_list args;
  va_start(args, count);
  for (int i = 0; i < count; i++) {
    OrgValue *item = va_arg(args, OrgValue *);
    org_list_append(a, v, item);
  }
  va_end(args);

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

static long long org_value_to_long(OrgValue *v) {
  if (!v)
    return 0;
  switch (v->type) {
  case ORG_INT_TYPE:
  case ORG_DEC_TYPE:
    return atoll(v->str_val);
  case ORG_STR_TYPE:
    return strlen(v->str_val);
  case ORG_LIST_TYPE:
    return v->list_val->size;
  default:
    return 0;
  }
}

static OrgValue *org_print(Arena *a, OrgValue *v) {
  if (!v) {
    printf("null\n");
    return v;
  }
  if (v->type == ORG_STR_TYPE) {
    if (v->str_val) {
      printf("%s\n", v->str_val);
    } else {
      printf("\"\"\n");
    }
  } else if (v->type == ORG_LIST_TYPE) {
    if (!v->list_val) {
      printf("[]\n");
      return v;
    }
    printf("[");
    for (size_t i = 0; i < v->list_val->size; i++) {
      if (i > 0)
        printf(" ");
      OrgValue *item = v->list_val->items[i];
      if (item->type == ORG_STR_TYPE)
        printf("\"%s\"", item->str_val);
      else if (item->str_val)
        printf("%s", item->str_val);
      else if (item->type == ORG_LIST_TYPE)
        printf("[...]"); // Nested list simplistic print
      else
        printf("?");
    }
    printf("]\n");
  } else {
    printf("%s\n", v->str_val);
  }
  return v;
}

// Operators (Mock implementation with Coercion)

static OrgValue *org_op_infix(Arena *a, const char *op, OrgValue *left,
                              OrgValue *right) {
  long long l_val = org_value_to_long(left);
  long long r_val = org_value_to_long(right);
  long long res = 0;

  // Simple integer arithmetic for prototype
  if (strcmp(op, "+") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val + r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "-") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val - r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "*") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val * r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "/") == 0) {
    char buf[64];
    sprintf(buf, "%lld", r_val != 0 ? l_val / r_val : 0);
    return org_int_from_str(a, buf);
  }

  // Comparisons return Boolean (org_bool)
  if (strcmp(op, ">") == 0)
    return org_bool(a, l_val > r_val);
  else if (strcmp(op, "<") == 0)
    return org_bool(a, l_val < r_val);
  else if (strcmp(op, ">=") == 0)
    return org_bool(a, l_val >= r_val);
  else if (strcmp(op, "<=") == 0)
    return org_bool(a, l_val <= r_val);
  else if (strcmp(op, "=") == 0)
    return org_bool(a, l_val == r_val);
  else if (strcmp(op, "<>") == 0)
    return org_bool(a, l_val != r_val);

  else {
    // Default fallback
    printf("Debug: %s %s %s\n", left->str_val ? left->str_val : "List", op,
           right->str_val ? right->str_val : "List");
    return left;
  }
}

static OrgValue *org_op_prefix(Arena *a, const char *op, OrgValue *right) {
  // Coercion mock
  if (strcmp(op, "@") == 0)
    return right; // Resource tag is identity for now
  printf("Debug: %s %s\n", op, right->str_val ? right->str_val : "List");
  return right;
}

#endif
