#include "orglang.h"
#include <stdio.h>

int main() {
    Arena *arena = arena_create(1024 * 1024);
    
    // Program start
    org_print(arena, org_string_from_c(arena, "Hello, OrgLang!"));

    // Program end
    
    arena_free(arena);
    return 0;
}
