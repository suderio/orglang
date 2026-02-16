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
typedef struct Arena Arena;
typedef struct OrgValue OrgValue;
struct OrgResourceInstance; // Forward decl

static OrgValue *org_int_from_str(Arena *a, const char *s);
static OrgValue *org_dec_from_str(Arena *a, const char *s);
static OrgValue *org_string_from_c(Arena *a, const char *s);
static OrgValue *org_error_make(Arena *a);
static long long org_value_to_long(OrgValue *v);
static char *org_value_to_cstring(Arena *a, OrgValue *v);

// Tiny Arena Allocator
typedef struct Arena {
  char *data;
  size_t size;
  size_t offset;
  struct OrgResourceInstance *resources_head;
  void *context; // Points to OrgContext
} Arena;

static Arena *arena_create(size_t size) {
  Arena *a = (Arena *)malloc(sizeof(Arena));
  a->data = (char *)malloc(size);
  a->size = size;
  a->offset = 0;
  a->resources_head = NULL;
  a->context = NULL;
  return a;
}

// Implemented later to handle type dependencies
static void arena_free(Arena *a);

static void arena_resource_register(Arena *a, struct OrgResourceInstance *res);

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
  ORG_SCOPED_ITERATOR_DATA,
  ORG_ERROR_TYPE
} OrgType;

typedef struct OrgValue OrgValue; // Forward declaration

// Function Pointer Type: (Arena, FunctionObject, Left, Right)
typedef OrgValue *(*OrgFuncPtr)(Arena *a, OrgValue *func, OrgValue *left,
                                OrgValue *right);

typedef struct OrgList {
  OrgValue **items;
  size_t capacity;
  size_t size;
} OrgList;

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
  struct OrgResourceInstance *next_resource;
} OrgResourceInstance;

typedef struct OrgScopedIterator {
  OrgIterator *upstream;
  OrgResource *def;
  OrgValue *context; // The Arena handle
} OrgScopedIterator;

// Function Pointer Type: (Arena, FunctionObject, Left, Right)
typedef OrgValue *(*OrgFuncPtr)(Arena *a, OrgValue *func, OrgValue *left,
                                OrgValue *right);

struct OrgValue {
  OrgType type;
  char *str_val;     // For INT, DEC, STR
  OrgList *list_val; // For LIST
  OrgFunction *func_val;
  OrgResource *resource_val;
  OrgResourceInstance *instance_val;
  OrgIterator *iterator_val;
  struct OrgScopedIterator *scoped_val;
  OrgValue *err_val; // For ERROR
};

// -- Scheduler Types --

typedef struct OrgContext OrgContext;
typedef struct OrgFiber OrgFiber;

// Fiber Function: (Fiber, Context)
typedef void (*OrgFiberFunc)(OrgFiber *fiber, OrgContext *ctx);

struct OrgFiber {
  int id;
  OrgFiberFunc resume;
  OrgValue *state;  // For closure/continuation state
  OrgValue *result; // For passing results between steps
  OrgFiber *next;   // Queue link
  OrgFiber *parent; // Who spawned me (optional, for join)
  Arena *arena;     // Sub-arena (optional)
};

typedef struct OrgScheduler {
  OrgFiber *ready_head;
  OrgFiber *ready_tail;
  int fiber_id_counter;
} OrgScheduler;

struct OrgContext {
  Arena *global_arena;
  OrgScheduler scheduler;
};

// ... (helpers) ...

static OrgValue *org_func_create(Arena *a, OrgFuncPtr func) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_FUNC_TYPE;
  v->func_val = (OrgFunction *)arena_alloc(a, sizeof(OrgFunction));
  v->func_val->func = func;
  return v;
}

static OrgValue *org_call(Arena *a, OrgValue *fn, OrgValue *left,
                          OrgValue *right) {
  if (!fn || fn->type != ORG_FUNC_TYPE) {
    printf("Runtime Error: Attempt to call non-function\n");
    return NULL;
  }
  if (left == NULL)
    left = org_error_make(a);
  if (right == NULL)
    right = org_error_make(a);
  return fn->func_val->func(a, fn, left, right);
}

static OrgValue *org_error_make(Arena *a) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_ERROR_TYPE;
  return v;
}

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

static void arena_resource_register(Arena *a, struct OrgResourceInstance *res) {
  if (!a || !res)
    return;
  res->next_resource = a->resources_head;
  a->resources_head = res;
}

static void arena_free(Arena *a) {
  if (!a)
    return;
  // Teardown resources
  OrgResourceInstance *curr = a->resources_head;
  while (curr) {
    if (curr->def && curr->def->teardown &&
        curr->def->teardown->type == ORG_FUNC_TYPE) {
      // Call teardown(this=teardown_func, left=state, right=NULL)
      OrgValue *err = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
      err->type = ORG_ERROR_TYPE;
      curr->def->teardown->func_val->func(a, curr->def->teardown, curr->state,
                                          NULL);
    }
    curr = curr->next_resource;
  }
  if (a->data)
    free(a->data);
  free(a);
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
  } else if (strcmp(syscall_name, "arena_create") == 0) {
    Arena *new_arena = arena_create(1024 * 1024); // Default 1MB
    char buf[64];
    sprintf(buf, "%lld", (long long)new_arena);
    return org_int_from_str(a, buf);
  } else if (strcmp(syscall_name, "arena_release") == 0) {
    if (l->size < 2)
      return NULL;
    long long ptr = org_value_to_long(l->items[1]);
    Arena *target = (Arena *)ptr;
    arena_free(target);
    return NULL;
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
  if (val)
    return org_int_from_str(a, "1");
  return org_int_from_str(a, "0");
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

  // Call next(this=instance, left=Error, right=NULL)
  OrgValue *err = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  err->type = ORG_ERROR_TYPE;
  return next_func->func_val->func(a, instance_val, err, NULL);
}

static OrgValue *list_iterator_next(Arena *a, OrgIterator *it) {
  OrgList *state_list = it->state->list_val;
  OrgValue *source_val = state_list->items[0];
  OrgValue *index_val = state_list->items[1];

  long long index = org_value_to_long(index_val);
  OrgList *source = source_val->list_val;

  if (index >= source->size) {
    return NULL;
  }

  OrgValue *item = source->items[index];

  // Update state logic: Update index
  char buf[32];
  sprintf(buf, "%lld", index + 1);
  state_list->items[1] = org_int_from_str(a, buf);

  return item;
}

static OrgValue *org_list_iterator(Arena *a, OrgValue *list) {
  OrgValue *idx = org_int_from_str(a, "0");
  OrgValue *state = org_list_make(a, 2, list, idx);
  return org_iterator_create(a, list_iterator_next, state);
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

  OrgValue *err = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  err->type = ORG_ERROR_TYPE;

  if (transform->type == ORG_FUNC_TYPE) {
    return transform->func_val->func(a, transform, err, val);
  }

  // If transform is Resource Instance (e.g. detailed step)
  if (transform->type == ORG_RESOURCE_INSTANCE_TYPE) {
    OrgValue *step = transform->instance_val->def->step;
    if (step && step->type == ORG_FUNC_TYPE) {
      return step->func_val->func(a, transform, err, val);
    }
  }
  return val;
}

static OrgValue *scoped_iterator_next(Arena *a, OrgIterator *it) {
  if (!it->state->scoped_val)
    return NULL;
  OrgScopedIterator *scoped = it->state->scoped_val;

  // 1. Lazy Setup
  if (!scoped->context) {
    OrgValue *err = org_error_make(a);
    if (scoped->def->setup && scoped->def->setup->type == ORG_FUNC_TYPE) {
      scoped->context =
          scoped->def->setup->func_val->func(a, scoped->def->setup, err, NULL);
    }
  }

  // 2. Context Switch (Arena)
  Arena *target_arena = a;
  if (scoped->context && scoped->context->type == ORG_INT_TYPE) {
    long long ptr = atoll(scoped->context->str_val);
    if (ptr != 0)
      target_arena = (Arena *)ptr;
  }

  // 3. Upstream Next
  OrgValue *val = scoped->upstream->next(target_arena, scoped->upstream);

  // 4. Teardown on End/Error
  if (!val || val->type == ORG_ERROR_TYPE) {
    if (scoped->def->teardown && scoped->def->teardown->type == ORG_FUNC_TYPE) {
      OrgValue *err = org_error_make(a);
      scoped->def->teardown->func_val->func(a, scoped->def->teardown,
                                            scoped->context, NULL);
    }
    return val;
  }
  return val;
}

static OrgValue *org_scoped_iterator_create(Arena *a, OrgValue *upstream_iter,
                                            OrgValue *resource_def) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_ITERATOR_TYPE;
  v->iterator_val = (OrgIterator *)arena_alloc(a, sizeof(OrgIterator));
  v->iterator_val->next = scoped_iterator_next;

  // State holds the Scoped Data
  OrgValue *state = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  state->type = ORG_SCOPED_ITERATOR_DATA; // Mock type or just use opaque
  state->scoped_val =
      (OrgScopedIterator *)arena_alloc(a, sizeof(OrgScopedIterator));
  state->scoped_val->upstream = upstream_iter->iterator_val;
  state->scoped_val->def = resource_def->resource_val;
  state->scoped_val->context = NULL;

  v->iterator_val->state = state;
  return v;
}

// -- Args Resource --

static int org_argc;
static char **org_argv;

static OrgValue *org_resource_args_next(Arena *a, OrgIterator *it) {
  // state is int (0 = not emitted, 1 = emitted)
  OrgValue *state = it->state;
  long long emitted = org_value_to_long(state);

  if (emitted) {
    return NULL; // End of Stream
  }

  // Emit Arguments Table
  char emitted_buf[2];
  sprintf(emitted_buf, "1");
  state->str_val = (char *)arena_alloc(a, 2); // Simple mutable update
  strcpy(state->str_val, "1");

  // Create List from argv
  OrgValue *args_list = org_list_create(a, org_argc);
  for (int i = 0; i < org_argc; i++) {
    org_list_append(a, args_list, org_string_from_c(a, org_argv[i]));
  }
  return args_list;
}

// Special Args Iterator
// We will define `org_resource_args_next_func` compatible with OrgFuncPtr.

static OrgValue *org_resource_args_next_func(Arena *a, OrgValue *func,
                                             OrgValue *left, OrgValue *right) {
  // func is Instance (passed by resource_iterator_next)
  if (!func || func->type != ORG_RESOURCE_INSTANCE_TYPE)
    return NULL;

  OrgResourceInstance *inst = func->instance_val;
  // State: 0 (pending), 1 (done)
  OrgValue *state = inst->state;
  long long emitted = org_value_to_long(state);

  if (emitted)
    return NULL;

  // Mark done
  state->str_val = (char *)arena_alloc(a, 2);
  strcpy(state->str_val, "1");

  // Return Args
  OrgValue *args_list = org_list_create(a, org_argc);
  for (int i = 0; i < org_argc; i++) {
    org_list_append(a, args_list, org_string_from_c(a, org_argv[i]));
  }
  return args_list;
}

static OrgValue *org_resource_args_create_wrap(Arena *a) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_RESOURCE_INSTANCE_TYPE;
  v->instance_val =
      (OrgResourceInstance *)arena_alloc(a, sizeof(OrgResourceInstance));

  // Mock Def
  v->instance_val->def = (OrgResource *)arena_alloc(a, sizeof(OrgResource));
  v->instance_val->def->next = org_func_create(a, org_resource_args_next_func);
  // setup/step/teardown NULL

  // Initial State: "0"
  v->instance_val->state = org_int_from_str(a, "0");

  arena_resource_register(a, v->instance_val);
  return v;
}

// -- Stdout Resource --

static OrgValue *org_resource_stdout_step(Arena *a, OrgValue *func,
                                          OrgValue *left, OrgValue *right) {
  // func is Instance (this).
  // left is Error (or upstream source info).
  // right is the Value to print.

  org_print(a, right);
  return right;
}

static OrgValue *org_resource_stdout_create_wrap(Arena *a) {
  OrgValue *v = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
  v->type = ORG_RESOURCE_INSTANCE_TYPE;
  v->instance_val =
      (OrgResourceInstance *)arena_alloc(a, sizeof(OrgResourceInstance));

  v->instance_val->def = (OrgResource *)arena_alloc(a, sizeof(OrgResource));
  v->instance_val->def->step = org_func_create(a, org_resource_stdout_step);
  v->instance_val->def->setup = NULL;
  v->instance_val->def->teardown = NULL;
  v->instance_val->def->next = NULL;

  v->instance_val->state = NULL;

  arena_resource_register(a, v->instance_val);
  return v;
}

// -- Scheduler Implementation --

static void org_sched_init(OrgContext *ctx, Arena *global_arena) {
  ctx->global_arena = global_arena;
  global_arena->context = ctx;
  ctx->scheduler.ready_head = NULL;
  ctx->scheduler.ready_tail = NULL;
  ctx->scheduler.fiber_id_counter = 1;
}

static OrgFiber *org_sched_spawn(OrgContext *ctx, OrgFiberFunc func,
                                 OrgValue *state) {
  // Use global arena for now for Fiber structs.
  Arena *a = ctx->global_arena;

  OrgFiber *f = (OrgFiber *)arena_alloc(a, sizeof(OrgFiber));
  f->id = ctx->scheduler.fiber_id_counter++;
  f->resume = func;
  f->state = state;
  f->result = NULL;
  f->next = NULL;
  f->parent = NULL;
  f->arena = a; // Share global arena for now

  // Enqueue
  if (ctx->scheduler.ready_tail) {
    ctx->scheduler.ready_tail->next = f;
    ctx->scheduler.ready_tail = f;
  } else {
    ctx->scheduler.ready_head = f;
    ctx->scheduler.ready_tail = f;
  }

  return f;
}

static void org_sched_run(OrgContext *ctx) {
  // Event Loop
  while (ctx->scheduler.ready_head) {
    // 1. Pop
    OrgFiber *f = ctx->scheduler.ready_head;
    ctx->scheduler.ready_head = f->next;
    if (!ctx->scheduler.ready_head) {
      ctx->scheduler.ready_tail = NULL;
    }

    // 2. Run
    if (f->resume) {
      f->resume(f, ctx);
    }

    // 3. IO Poll (Mock)
    // if (io_queue_has_completions) process();
  }
}

// -- Scheduler Tasks --

static void org_sink_task(OrgFiber *fiber, OrgContext *ctx) {
  OrgValue *state = fiber->state;
  // state = [Item, Sink]
  if (state->type != ORG_LIST_TYPE || state->list_val->size < 2)
    return;

  OrgValue *item = state->list_val->items[0];
  OrgValue *sink = state->list_val->items[1];
  Arena *a = fiber->arena;

  // Logic from org_op_infix default push behavior

  // 1. If Sink is Function
  if (sink->type == ORG_FUNC_TYPE) {
    OrgValue *err = org_error_make(a);
    sink->func_val->func(a, sink, err, item);
    return;
  }

  // 2. If Sink is Resource Instance
  if (sink->type == ORG_RESOURCE_INSTANCE_TYPE) {
    OrgValue *step = sink->instance_val->def->step;
    if (step && step->type == ORG_FUNC_TYPE) {
      OrgValue *err = org_error_make(a);
      step->func_val->func(a, sink, err, item);
    }
    return;
  }

  // 3. Fallback / TODO: Broadcast lists etc.
}

static void org_pump_task(OrgFiber *fiber, OrgContext *ctx) {
  OrgValue *state = fiber->state;
  // state = [Iterator, Sink]
  if (state->type != ORG_LIST_TYPE || state->list_val->size < 2)
    return;

  OrgValue *iter = state->list_val->items[0];
  OrgValue *sink = state->list_val->items[1];
  Arena *a = fiber->arena;

  // Get Next Item
  // Iterator Next might need Arena. Use Fiber's arena.
  if (iter->type != ORG_ITERATOR_TYPE)
    return;

  OrgValue *item = iter->iterator_val->next(a, iter->iterator_val);

  // Check End
  if (!item)
    return; // Null means end
  if (item->type == ORG_STR_TYPE && strcmp(item->str_val, "Error") == 0)
    return;

  // Spawn Sink Task
  OrgValue *task_state = org_list_make(a, 2, item, sink);
  org_sched_spawn(ctx, org_sink_task, task_state);

  // Requeue Self
  if (ctx->scheduler.ready_tail) {
    ctx->scheduler.ready_tail->next = fiber;
    ctx->scheduler.ready_tail = fiber;
  } else {
    ctx->scheduler.ready_head = fiber;
    ctx->scheduler.ready_tail = fiber;
  }
  fiber->next = NULL;
}

static OrgValue *org_op_infix(Arena *a, const char *op, OrgValue *left,
                              OrgValue *right) {
  if (left == NULL)
    left = org_error_make(a);
  if (right == NULL)
    right = org_error_make(a);

  OrgContext *ctx = (OrgContext *)a->context;

  // Flow Operator -> (Pull / Lazy / Map Support)
  if (strcmp(op, "->") == 0) {
    OrgValue *iter = NULL;

    // 0. Middleware / Scoped Iterator
    if (right->type == ORG_RESOURCE_TYPE) {
      if (left->type == ORG_ITERATOR_TYPE) {
        iter = left;
      } else {
        iter = org_list_iterator(a, left); // Promote
      }
      return org_scoped_iterator_create(a, iter, right);
    }

    // 1. Promote Left to Iterator IF appropriate
    // Logic: If we are piping into a Transform (Func), we want an Iterator
    // (Lazy Map). If we are piping into a Sink (ResourceInstance), we want a
    // Pump Task.

    int right_is_sink = (right->type == ORG_RESOURCE_INSTANCE_TYPE);

    if (left->type == ORG_ITERATOR_TYPE || left->type == ORG_LIST_TYPE ||
        left->type == ORG_PAIR_TYPE ||
        (left->type == ORG_RESOURCE_INSTANCE_TYPE &&
         left->instance_val->def->next)) {
      // Potential Iterator Context
      if (left->type == ORG_ITERATOR_TYPE)
        iter = left;
      else if (left->type == ORG_LIST_TYPE || left->type == ORG_PAIR_TYPE)
        iter = org_list_iterator(a, left);
      else
        iter = org_iterator_create(a, resource_iterator_next, left);

      if (right->type == ORG_FUNC_TYPE) {
        // Lazy Map
        OrgValue *state = org_list_make(a, 2, iter, right);
        return org_iterator_create(a, map_iterator_next, state);
      }

      if (right_is_sink) {
        // Spawn Pump Task
        if (ctx) {
          OrgValue *state = org_list_make(a, 2, iter, right);
          org_sched_spawn(ctx, org_pump_task, state);
        }
        return NULL;
      }
    }

    // Default: Spawn Single Task (Value -> Sink/Func)
    if (ctx) {
      OrgValue *state = org_list_make(a, 2, left, right);
      org_sched_spawn(ctx, org_sink_task, state);
    }
    return left;
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
  } else if (strcmp(op, "&") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val & r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "|") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val | r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "^") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val ^ r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "<<") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val << r_val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, ">>") == 0) {
    char buf[64];
    sprintf(buf, "%lld", l_val >> r_val);
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

  // Logical / Bitwise (Non-short-circuit for simplicity here, short-circuit
  // handled by emit/thunk ideally) For now we implement basic boolean logic.

  if (strcmp(op, "&&") == 0) {
    return org_bool(a, (l_val != 0) && (r_val != 0));
  }
  if (strcmp(op, "||") == 0) {
    return org_bool(a, (l_val != 0) || (r_val != 0));
  }

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

  if (strcmp(op, "-") == 0) {
    // Unary Negation
    // We can reuse infix '-' with 0 - right?
    // Or just implement negation.
    // For now: 0 - right
    OrgValue *zero = org_int_from_str(a, "0");
    return org_op_infix(a, "-", zero, right);
  } else if (strcmp(op, "!") == 0) {
    long long val = org_value_to_long(right);
    return org_bool(a, !val);
  } else if (strcmp(op, "~") == 0) {
    long long val = org_value_to_long(right);
    char buf[64];
    sprintf(buf, "%lld", ~val);
    return org_int_from_str(a, buf);
  } else if (strcmp(op, "++") == 0) {
    // Increment: + 1
    OrgValue *one = org_int_from_str(a, "1");
    return org_op_infix(a, "+", right, one);
  } else if (strcmp(op, "--") == 0) {
    // Decrement: - 1
    OrgValue *one = org_int_from_str(a, "1");
    return org_op_infix(a, "-", right, one);
  } else if (strcmp(op, "@") == 0) {
    if (right->type == ORG_RESOURCE_TYPE) {
      OrgValue *instance = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
      instance->type = ORG_RESOURCE_INSTANCE_TYPE;
      instance->instance_val =
          (OrgResourceInstance *)arena_alloc(a, sizeof(OrgResourceInstance));
      instance->instance_val->def = right->resource_val;
      instance->instance_val->state = NULL;

      OrgValue *setup = right->resource_val->setup;
      if (setup && setup->type == ORG_FUNC_TYPE) {
        OrgValue *err = (OrgValue *)arena_alloc(a, sizeof(OrgValue));
        err->type = ORG_ERROR_TYPE;
        instance->instance_val->state =
            setup->func_val->func(a, right, err, NULL);
      }
      arena_resource_register(a, instance->instance_val);
      return instance;
    }
  }

  return right;
}

#endif
