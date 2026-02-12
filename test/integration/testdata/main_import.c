#include "orglang.h"
#include <stdio.h>

// Globals
static OrgValue *org_var_print = NULL;
static OrgValue *org_var_n = NULL;
static OrgValue *org_var_stdin = NULL;
static OrgValue *org_var_add = NULL;
static OrgValue *org_var_sub = NULL;
static OrgValue *org_var_res = NULL;
static OrgValue *org_var_Error = NULL;
static OrgValue *org_var_stdout = NULL;
static OrgValue *org_var_buf = NULL;
static OrgValue *org_var_math = NULL;


// Auxiliary Functions
static OrgValue *org_fn_0(Arena *arena, OrgValue *this_val, OrgValue *args) {
return org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "write"), org_int_from_str(arena, "1"), args, org_int_from_str(arena, "-1")));
}

static OrgValue *org_fn_1(Arena *arena, OrgValue *this_val, OrgValue *args) {
return org_op_infix(arena, "->", args, org_op_prefix(arena, "@", org_var_stdout));
}

static OrgValue *org_fn_2(Arena *arena, OrgValue *this_val, OrgValue *args) {
return NULL;
}

static OrgValue *org_fn_3(Arena *arena, OrgValue *this_val, OrgValue *args) {
(org_var_buf = org_malloc(arena, org_value_to_long(org_int_from_str(arena, "64"))), org_pair_make(arena, org_string_from_c(arena, "buf"), org_var_buf));
(org_var_n = org_syscall(arena, org_list_make(arena, 4, org_string_from_c(arena, "read"), org_int_from_str(arena, "0"), org_var_buf, org_int_from_str(arena, "64"))), org_pair_make(arena, org_string_from_c(arena, "n"), org_var_n));
return org_value_evaluate(arena, org_table_get(arena, org_list_make(arena, 2, org_pair_make(arena, org_string_from_c(arena, "true"), org_var_buf), org_pair_make(arena, org_string_from_c(arena, "false"), org_var_Error)), org_op_infix(arena, ">", org_var_n, org_int_from_str(arena, "0"))));
}

static OrgValue *org_fn_5(Arena *arena, OrgValue *this_val, OrgValue *args) {
return org_op_infix(arena, "+", org_table_get(arena, args, org_int_from_str(arena, "0")), org_table_get(arena, args, org_int_from_str(arena, "1")));
}

static OrgValue *org_fn_6(Arena *arena, OrgValue *this_val, OrgValue *args) {
return org_op_infix(arena, "-", org_table_get(arena, args, org_int_from_str(arena, "0")), org_table_get(arena, args, org_int_from_str(arena, "1")));
}

static OrgValue *org_module_4(Arena *arena, OrgValue *this_val, OrgValue *args) {
    OrgValue *org_var_add = NULL;
    OrgValue *org_var_sub = NULL;
    OrgValue *stmt_0 = (org_var_add = org_func_create(arena, org_fn_5), org_pair_make(arena, org_string_from_c(arena, "add"), org_var_add));
    OrgValue *stmt_1 = (org_var_sub = org_func_create(arena, org_fn_6), org_pair_make(arena, org_string_from_c(arena, "sub"), org_var_sub));
    return org_list_make(arena, 2, stmt_0, stmt_1);
}


int main() {
    Arena *arena = arena_create(1024 * 1024);
    
    // Program start
    (org_var_Error = org_string_from_c(arena, "Error"), org_pair_make(arena, org_string_from_c(arena, "Error"), org_var_Error));
    (org_var_stdout = org_resource_create(arena, NULL, org_func_create(arena, org_fn_0), NULL, NULL), org_pair_make(arena, org_string_from_c(arena, "stdout"), org_var_stdout));
    (org_var_print = org_func_create(arena, org_fn_1), org_pair_make(arena, org_string_from_c(arena, "print"), org_var_print));
    (org_var_stdin = org_resource_create(arena, org_func_create(arena, org_fn_2), NULL, NULL, org_func_create(arena, org_fn_3)), org_pair_make(arena, org_string_from_c(arena, "stdin"), org_var_stdin));
    (org_var_math = org_module_4(arena, NULL, NULL), org_pair_make(arena, org_string_from_c(arena, "math"), org_var_math));
    (org_var_res = org_call(arena, org_table_get(arena, org_var_math, org_string_from_c(arena, "add")), org_list_make(arena, 2, org_int_from_str(arena, "10"), org_int_from_str(arena, "5"))), org_pair_make(arena, org_string_from_c(arena, "res"), org_var_res));
    org_call(arena, org_var_print, org_var_res);
    org_call(arena, org_var_print, org_string_from_c(arena, "\n"));

    // Program end
    
    arena_free(arena);
    return 0;
}
