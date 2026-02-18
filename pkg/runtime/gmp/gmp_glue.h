#ifndef ORG_GMP_GLUE_H
#define ORG_GMP_GLUE_H

#include "../core/arena.h"

/*
 * GMP Memory Glue — Redirects all GMP allocations through a
 * thread-local Arena.
 *
 * The scheduler MUST call org_gmp_set_arena() on each fiber resume
 * to set the current fiber's arena as the GMP allocation target.
 */

/* Initialize GMP memory functions to use Arena allocation.
 * Must be called once at program startup, before any GMP operations. */
void org_gmp_init(void);

/* Set the thread-local arena used by GMP for the current fiber. */
void org_gmp_set_arena(Arena *arena);

/* Get the current thread-local arena (mainly for internal use). */
Arena *org_gmp_get_arena(void);

#endif /* ORG_GMP_GLUE_H */
