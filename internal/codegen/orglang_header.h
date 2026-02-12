#ifndef ORGLANG_H
#define ORGLANG_H

#include <fcntl.h>
#include <math.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// Forward declarations
typedef struct Arena Arena;
typedef struct OrgValue OrgValue;
static OrgValue *org_int_from_str(Arena *a, const char *s);
static long long org_value_to_long(OrgValue *v);
static char *org_value_to_cstring(Arena *a, OrgValue *v);

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
  ORG_LIST_TYPE,
  ORG_PAIR_TYPE,
  ORG_FUNC_TYPE,
  ORG_RESOURCE_TYPE,
  ORG_RESOURCE_INSTANCE_TYPE,
  ORG_ITERATOR_TYPE,
  ORG_ERROR_TYPE
} OrgType;

typedef struct OrgValue OrgValue; // Forward declaration

typedef struct OrgList {
  OrgValue **items;
  size_t capacity;
  size_t size;
} OrgList;

typedef OrgValue *(*OrgFuncPtr)(Arena *a, OrgValue *this_val, OrgValue *args);

typedef struct OrgFunction {
  OrgFuncPtr func;
  // Closure support? For now, no captured env
} OrgFunction;

typedef struct OrgIterator OrgIterator;
typedef OrgValue *(*OrgNextFunc)(Arena *a, OrgIterator *it);

struct OrgIterator {
  OrgNextFunc next;
  OrgValue *state;
};

// Resource Definition (Static)
typedef struct OrgResource {
  OrgValue *setup;
  OrgValue *step;
  OrgValue *teardown;
  OrgValue *next;
} OrgResource;

// Resource Instance (Dynamic State)
typedef struct OrgResourceInstance {
  OrgResource *def;
  OrgValue *state;
} OrgResourceInstance;

struct OrgValue {
  OrgType type;
  char *str_val;     // For INT, DEC, STR
  OrgList *list_val; // For LIST
  OrgFunction *func_val;
  OrgResource *resource_val;
  OrgResourceInstance *instance_val;
  OrgIterator *iterator_val;
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

static void org_list_append(Arena *a, OrgValue *list, OrgValue *item);

static OrgValue *org_pair_make(Arena *a, OrgValue *key, OrgValue *val);

static OrgValue *org_func_create(Arena *a, OrgFuncPtr func) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_FUNC_TYPE;
  v->func_val = (OrgFunction *)arena_alloc(a, sizeof(OrgFunction));
  v->func_val->func = func;
  return v;
}

static OrgValue *org_call(Arena *a, OrgValue *fn, OrgValue *args) {
  if (!fn || fn->type != ORG_FUNC_TYPE) {
    printf("Runtime Error: Attempt to call non-function\n");
    return NULL;
  }
  return fn->func_val->func(a, fn, args);
}

static OrgValue *org_resource_create(Arena *a, OrgValue *setup, OrgValue *step,
                                     OrgValue *teardown, OrgValue *next) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_RESOURCE_TYPE;
  v->resource_val = (OrgResource *)arena_alloc(a, sizeof(OrgResource));
  v->resource_val->setup = setup;
  v->resource_val->step = step;
  v->resource_val->teardown = teardown;
  v->resource_val->next = next;

  return v;
}

static OrgValue *org_iterator_create(Arena *a, OrgNextFunc next,
                                     OrgValue *state) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_ITERATOR_TYPE;
  v->iterator_val = (OrgIterator *)arena_alloc(a, sizeof(OrgIterator));
  v->iterator_val->next = next;
  v->iterator_val->state = state;
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

static OrgValue *org_pair_make(Arena *a, OrgValue *key, OrgValue *val) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_PAIR_TYPE;
  v->list_val = org_list_make(a, 2, key, val)->list_val;
  return v;
}

static OrgValue *org_malloc(Arena *a, long long size) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_STR_TYPE;
  v->str_val = (char *)arena_alloc(a, size + 1);
  memset(v->str_val, 0, size + 1);
  return v;
}

static char *org_value_to_cstring(Arena *a, OrgValue *v) {
  if (!v)
    return (char *)"";
  if (v->type == ORG_STR_TYPE)
    return v->str_val;
  if (v->type == ORG_INT_TYPE)
    return v->str_val;
  return (char *)"";
}

static OrgValue *org_syscall(Arena *a, OrgValue *args) {
  if (args->type != ORG_LIST_TYPE) {
    printf("Syscall expects list arguments\n");
    return NULL;
  }
  OrgList *l = args->list_val;
  if (l->size < 1)
    return NULL;

  char *syscall_name = org_value_to_cstring(a, l->items[0]);

  if (strcmp(syscall_name, "read") == 0) {
    // ("read", fd, buffer, size)
    if (l->size < 4)
      return NULL;
    int fd = (int)org_value_to_long(l->items[1]);
    OrgValue *buf_val = l->items[2];
    long size = org_value_to_long(l->items[3]);

    char *ptr = buf_val->str_val;
    ssize_t n = read(fd, ptr, size);

    char res_buf[64];
    sprintf(res_buf, "%ld", (long)n);
    return org_int_from_str(a, res_buf);
  } else if (strcmp(syscall_name, "write") == 0) {
    // ("write", fd, buffer, len)
    if (l->size < 4)
      return NULL;
    int fd = (int)org_value_to_long(l->items[1]);
    OrgValue *buf_val = l->items[2];
    long size = org_value_to_long(l->items[3]);

    char *ptr = org_value_to_cstring(a, buf_val);
    if (size == -1) {
      size = strlen(ptr);
    }
    ssize_t n = write(fd, ptr, size);

    char res_buf[64];
    sprintf(res_buf, "%ld", (long)n);
    return org_int_from_str(a, res_buf);
  }
  return NULL;
}

// Helper for matching keys
static int org_key_match(OrgValue *itemKey, OrgValue *searchKey) {
  if (!itemKey || !searchKey)
    return 0;
  if (itemKey->type != searchKey->type)
    return 0;
  if (itemKey->type == ORG_STR_TYPE || itemKey->type == ORG_INT_TYPE) {
    return strcmp(itemKey->str_val, searchKey->str_val) == 0;
  }
  return 0;
}

static OrgValue *org_table_get(Arena *a, OrgValue *table, OrgValue *key) {
  if (!table || table->type != ORG_LIST_TYPE) {
    return NULL;
  }
  OrgList *l = table->list_val;

  // Try Key-Value lookup first
  for (size_t i = 0; i < l->size; i++) {
    OrgValue *item = l->items[i];
    if (item->type == ORG_PAIR_TYPE) {
      if (org_key_match(item->list_val->items[0], key)) {
        return item->list_val->items[1];
      }
    }
  }

  // Fallback to Index
  long long targetIdx = org_value_to_long(key);
  long long currentPos = 0;

  for (size_t i = 0; i < l->size; i++) {
    OrgValue *item = l->items[i];
    if (item->type != ORG_PAIR_TYPE) {
      if (currentPos == targetIdx)
        return item;
      currentPos++;
    }
  }

  return NULL;
}

// ... inside org_op_infix ... (I need target context for this replacement)
// I will split this into two calls or target org_table_get specifically.
// And another call for org_op_infix fallback.

// End of helpers section (Logic continues)

static OrgValue *org_value_evaluate(Arena *a, OrgValue *v) {
  if (!v)
    return NULL;
  // Prototype: Identity.
  // In future: If v is Thunk, execute.
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
  if (v->type == ORG_STR_TYPE || v->type == ORG_INT_TYPE ||
      v->type == ORG_DEC_TYPE) {
    if (v->str_val) {
      printf("%s\n", v->str_val);
    } else {
      if (v->type == ORG_STR_TYPE)
        printf("\"\"\n");
      else
        printf("0\n");
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
  } else if (v->type == ORG_RESOURCE_TYPE) {
    printf("<Resource Definition>\n");
  } else if (v->type == ORG_RESOURCE_INSTANCE_TYPE) {
    printf("<Resource Instance>\n");
  } else if (v->type == ORG_ITERATOR_TYPE) {
    printf("<Iterator>\n");
  } else {
    // Fallback safely
    printf("Unknown Type: %d\n", v->type);
    printf("?\n");
  }
  return v;
}

// Operators (Mock implementation with Coercion)

// Iterator Helpers
static OrgValue *resource_iterator_next(Arena *a, OrgIterator *it) {
  // state is the Instance
  OrgValue *instance_val = it->state;
  if (instance_val->type != ORG_RESOURCE_INSTANCE_TYPE)
    return NULL;

  OrgResourceInstance *inst = instance_val->instance_val;
  OrgValue *next_func = inst->def->next;

  if (!next_func || next_func->type != ORG_FUNC_TYPE)
    return NULL;

  // Call next(this=instance, args=NULL)
  return next_func->func_val->func(a, instance_val, NULL);
}

static OrgValue *map_iterator_next(Arena *a, OrgIterator *it) {
  OrgList *l = it->state->list_val;
  OrgValue *source = l->items[0];
  OrgValue *transform = l->items[1];

  OrgValue *val = source->iterator_val->next(a, source->iterator_val);
  if (!val)
    return NULL; // End of Stream

  // Check for Error/EOF
  if (val->type == ORG_STR_TYPE && strcmp(val->str_val, "Error") == 0) {
    return val;
  }

  if (transform->type == ORG_FUNC_TYPE) {
    return transform->func_val->func(a, transform, val);
  }

  // If transform is Resource Instance (e.g. detailed step)
  if (transform->type == ORG_RESOURCE_INSTANCE_TYPE) {
    OrgValue *step = transform->instance_val->def->step;
    if (step && step->type == ORG_FUNC_TYPE) {
      return step->func_val->func(a, transform, val);
    }
  }
  return val;
}

static OrgValue *org_op_infix(Arena *a, const char *op, OrgValue *left,
                              OrgValue *right) {
  // Flow Operator -> (Pull / Lazy / Map Support)
  if (strcmp(op, "->") == 0) {
    // right->type);
    OrgValue *iter = NULL;

    // 0. Guard: Error on Raw Resource Definitions
    if (left->type == ORG_RESOURCE_TYPE || right->type == ORG_RESOURCE_TYPE) {
      printf("Runtime Error: Resource Definition used directly in Flow. Use "
             "@resource to instantiate.\n");
      return NULL;
    }

    // 1. Promote Left to Iterator if possible
    if (left->type == ORG_ITERATOR_TYPE) {
      iter = left;
    } else if (left->type == ORG_RESOURCE_INSTANCE_TYPE &&
               left->instance_val->def->next) {
      // Create Iterator from Instance
      // We pass the Instance as state to the helper
      iter = org_iterator_create(a, resource_iterator_next, left);
    }

    // 2. If Left is Iterator
    if (iter) {
      // Case A: Right is Function/Transform -> Lazy Map
      if (right->type == ORG_FUNC_TYPE) {
        OrgValue *state = org_list_make(a, 2, iter, right);
        return org_iterator_create(a, map_iterator_next, state);
      }

      // Case B: Right is Sink  // Resource Instantiation (@def)
      if (right->type == ORG_RESOURCE_INSTANCE_TYPE) {
        OrgValue *result = NULL;
        while ((result = iter->iterator_val->next(a, iter->iterator_val)) !=
               NULL) {
          // Check for Error/EOF
          if (result->type == ORG_STR_TYPE &&
              strcmp(result->str_val, "Error") == 0) {
            break;
          }

          OrgValue *step = right->instance_val->def->step;
          if (step && step->type == ORG_FUNC_TYPE) {
            // Pass Instance (as 'this') and Result (as 'args')
            // Note: 'step' expects (this, args)
            step->func_val->func(a, right, result);
          }
        }
        return NULL; // Flow flows into sink, returns nothing (or maybe the sink
                     // state?)
      }

      // Case C: Right is List (Broadcast) - Not impl yet for Pull
      // Fallback to push behavior if Right is not func/resource?
    }

    // Default Push Behavior (Legacy/Value)
    // Left -> Right
    long long l_val = org_value_to_long(left);
    long long r_val = org_value_to_long(right);

    if (right->type == ORG_FUNC_TYPE) {
      return right->func_val->func(a, right, left);
    }
    if (right->type == ORG_RESOURCE_TYPE) {
      OrgValue *step = right->resource_val->step;
      if (step && step->type == ORG_FUNC_TYPE) {
        return step->func_val->func(a, right, left);
      }
    }
    if (right->type == ORG_RESOURCE_INSTANCE_TYPE) {
      OrgValue *step = right->instance_val->def->step;
      if (step && step->type == ORG_FUNC_TYPE) {
        return step->func_val->func(a, right, left);
      }
    }
    return right;
  }

  long long l_val = org_value_to_long(left);
  long long r_val = org_value_to_long(right);

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
  } else if (strcmp(op, "**") == 0) {
    char buf[64];
    sprintf(buf, "%lld", (long long)pow((double)l_val, (double)r_val));
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

  // Table Access . (Left=Table, Right=Key)
  if (strcmp(op, ".") == 0) {
    return org_table_get(a, left, right);
  }

  // Table Eval ? (Left=Key/Cond, Right=Table)
  if (strcmp(op, "?") == 0) {
    return org_table_get(a, right, left);
  }

  // Error Check ?? (Returns right if left is Error, else left)
  if (strcmp(op, "??") == 0) {
    if (left->type == ORG_ERROR_TYPE)
      return right;
    return left;
  }

  // Elvis Operator ?: (Returns right if left is falsy, else left)
  if (strcmp(op, "?:") == 0) {
    int falsy = 0;
    if (left->type == ORG_ERROR_TYPE)
      falsy = 1;
    else if (left->type == ORG_INT_TYPE && org_value_to_long(left) == 0)
      falsy = 1; // Assuming boolean false is 0
    else if (left->type == ORG_STR_TYPE && strlen(left->str_val) == 0)
      falsy = 1;
    else if (left->type == ORG_LIST_TYPE && left->list_val->size == 0)
      falsy = 1;

    if (falsy)
      return right;
    return left;
  }

  // Comma operator ,
  if (strcmp(op, ",") == 0) {
    if (left->type == ORG_LIST_TYPE) {
      org_list_append(a, left, right);
      return left;
    } else {
      return org_list_make(a, 2, left, right);
    }
  }

  // Logical / Bitwise (Non-short-circuit)
  if (strcmp(op, "&") == 0) {
    return org_bool(a, (l_val != 0) && (r_val != 0));
  }
  if (strcmp(op, "|") == 0) {
    return org_bool(a, (l_val != 0) || (r_val != 0));
  }
  if (strcmp(op, "^") == 0) {
    return org_bool(a, (l_val != 0) ^ (r_val != 0));
  }

  // Pair Construction :
  if (strcmp(op, ":") == 0) {
    return org_pair_make(a, left, right);
  }

  else {
    // Default fallback
    printf("Debug: %s %s %s\n", left->str_val ? left->str_val : "List", op,
           right->str_val ? right->str_val : "List");
    return left;
  }
}

static OrgValue *org_op_prefix(Arena *a, const char *op, OrgValue *right) {
  // Resource Instantiation (@def)
  if (strcmp(op, "@") == 0) {
    if (right->type == ORG_RESOURCE_TYPE) {
      // 1. Create Instance
      OrgValue *instance = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
      instance->type = ORG_RESOURCE_INSTANCE_TYPE;
      instance->instance_val =
          (OrgResourceInstance *)arena_alloc(a, sizeof(OrgResourceInstance));
      instance->instance_val->def = right->resource_val;
      instance->instance_val->state = NULL;

      // 2. Call Setup
      OrgValue *setup = right->resource_val->setup;
      if (setup && setup->type == ORG_FUNC_TYPE) {
        // Pass definition as 'this', and maybe some args?
        // For now, setup takes no args.
        instance->instance_val->state = setup->func_val->func(a, right, NULL);
      }
      return instance;
    }
  }

  // Fallback / Other prefixes
  return right;
}

#endif
