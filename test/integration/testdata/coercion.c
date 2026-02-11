#include "orglang.h"
#include <stdio.h>

int main() {
    Arena *arena = arena_create(1024 * 1024);
    
    // Program start
    org_print(arena, org_op_infix(arena, "+", org_list_make(arena, 3, org_int_from_str(arena, "1"), org_int_from_str(arena, "2"), org_int_from_str(arena, "4")), org_int_from_str(arena, "1")));
    org_print(arena, org_op_infix(arena, ">", org_string_from_c(arena, "test"), org_bool(arena, 1)));

    // Program end
    
    arena_free(arena);
    return 0;
}
