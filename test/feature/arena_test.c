#include "orglang.h"
#include <stdio.h>

// Globals
static OrgValue *org_var_Error = NULL;
static OrgValue *org_var_stdout = NULL;
static OrgValue *org_var_print = NULL;
static OrgValue *org_var_buf = NULL;
static OrgValue *org_var_n = NULL;
static OrgValue *org_var_Arena = NULL;
static OrgValue *org_var_Tracked = NULL;
static OrgValue *org_var_stdin = NULL;
static OrgValue *org_var_ptr = NULL;
static OrgValue *org_var_val = NULL;


// Auxiliary Functions
static OrgValue *org_fn_0(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), right, org_int_from_str(arena, "-1")));
}

static OrgValue *org_fn_1(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return org_op_infix(arena, "->", right, org_op_prefix(arena, "@", org_var_stdout));
}

static OrgValue *org_fn_2(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return NULL;
}

static OrgValue *org_fn_3(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
(org_var_buf = org_malloc(arena, org_value_to_long(org_int_from_str(arena, "64"))), org_pair_make(arena, org_string_from_c(arena, "buf"), org_var_buf));
(org_var_n = org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "read"), org_int_from_str(arena, "0"), org_var_buf, org_int_from_str(arena, "64"))), org_pair_make(arena, org_string_from_c(arena, "n"), org_var_n));
return org_value_evaluate(arena, org_table_get(arena, org_list_make(arena, 2, org_pair_make(arena, org_string_from_c(arena, "true"), org_var_buf), org_pair_make(arena, org_string_from_c(arena, "false"), org_var_Error)), org_op_infix(arena, ">", org_var_n, org_int_from_str(arena, "0"))));
}

static OrgValue *org_fn_4(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
(org_var_ptr = org_syscall(arena, org_list_make(arena, 1, org_string_from_c(arena, "arena_create"))), org_pair_make(arena, org_string_from_c(arena, "ptr"), org_var_ptr));
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "[ARENA SETUP] Created Arena: "), org_int_from_str(arena, "-1")));
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_var_ptr, org_int_from_str(arena, "-1")));
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "\n"), org_int_from_str(arena, "-1")));
return org_var_ptr;
}

static OrgValue *org_fn_5(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "[ARENA TEARDOWN] Freeing Arena: "), org_int_from_str(arena, "-1")));
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), left, org_int_from_str(arena, "-1")));
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "\n"), org_int_from_str(arena, "-1")));
return org_syscall(arena, org_list_make(arena, 2, org_string_from_c(arena, "arena_release"), left));
}

static OrgValue *org_fn_6(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return right;
}

static OrgValue *org_fn_7(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "[TRACKED SETUP]\n"), org_int_from_str(arena, "-1")));
return org_string_from_c(arena, "TrackedState");
}

static OrgValue *org_fn_8(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "[TRACKED TEARDOWN]\n"), org_int_from_str(arena, "-1")));
}

static OrgValue *org_fn_9(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
return right;
}

static OrgValue *org_fn_10(Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right) {
(org_var_val = org_op_prefix(arena, "@", org_var_Tracked), org_pair_make(arena, org_string_from_c(arena, "val"), org_var_val));
return right;
}


int main() {
    Arena *arena = arena_create(1024 * 1024);
    
    // Program start
    (org_var_Error = org_string_from_c(arena, "Error"), org_pair_make(arena, org_string_from_c(arena, "Error"), org_var_Error));
    (org_var_stdout = org_resource_create(arena, NULL, org_func_create(arena, org_fn_0), NULL, NULL), org_pair_make(arena, org_string_from_c(arena, "stdout"), org_var_stdout));
    (org_var_print = org_func_create(arena, org_fn_1), org_pair_make(arena, org_string_from_c(arena, "print"), org_var_print));
    (org_var_stdin = org_resource_create(arena, org_func_create(arena, org_fn_2), NULL, NULL, org_func_create(arena, org_fn_3)), org_pair_make(arena, org_string_from_c(arena, "stdin"), org_var_stdin));
    (org_var_Arena = org_resource_create(arena, org_func_create(arena, org_fn_4), NULL, org_func_create(arena, org_fn_5), org_func_create(arena, org_fn_6)), org_pair_make(arena, org_string_from_c(arena, "Arena"), org_var_Arena));
    (org_var_Tracked = org_resource_create(arena, org_func_create(arena, org_fn_7), NULL, org_func_create(arena, org_fn_8), org_func_create(arena, org_fn_9)), org_pair_make(arena, org_string_from_c(arena, "Tracked"), org_var_Tracked));
    org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "--- START TEST 1 (Middleware) ---\n"), org_int_from_str(arena, "-1")));
    org_op_infix(arena, "->", org_op_infix(arena, "->", org_op_infix(arena, "->", org_list_make(arena, 1, org_int_from_str(arena, "1")), org_var_Tracked), org_var_Arena), org_op_prefix(arena, "@", org_var_stdout));
    org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), org_string_from_c(arena, "\n--- START TEST 2 (Leak Cleanup) ---\n"), org_int_from_str(arena, "-1")));
    org_op_infix(arena, "->", org_op_infix(arena, "->", org_op_infix(arena, "->", org_list_make(arena, 1, org_int_from_str(arena, "1")), org_func_create(arena, org_fn_10)), org_var_Arena), org_op_prefix(arena, "@", org_var_stdout));

    // Program end
    
    arena_free(arena);
    return 0;
}
