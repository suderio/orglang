#ifndef ORG_OPS_H
#define ORG_OPS_H

#include "../core/values.h"

/*
 * Arithmetic Dispatch — Implements the numeric promotion matrix:
 *
 *   Left\Right  | Integer  | Rational | Decimal
 *   ------------|----------|----------|--------
 *   Integer     | Integer  | Rational | Decimal
 *   Rational    | Rational | Rational | Decimal
 *   Decimal     | Decimal  | Decimal  | Decimal
 *
 * All operations use the fast path for SmallInt+SmallInt when possible,
 * with automatic overflow promotion to BigInt.
 */

/* Arithmetic */
OrgValue org_add(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_sub(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_mul(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_div(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_mod(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_neg(Arena *arena, OrgValue a);
OrgValue org_pow(Arena *arena, OrgValue base, OrgValue exp);

/* Comparison — returns ORG_TRUE or ORG_FALSE */
OrgValue org_eq(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_lt(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_le(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_gt(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_ge(Arena *arena, OrgValue a, OrgValue b);
OrgValue org_ne(Arena *arena, OrgValue a, OrgValue b);

/*
 * Normalize an integer result: if a BigInt fits in a SmallInt,
 * convert it back. Used after arithmetic to keep values compact.
 */
OrgValue org_normalize_int(OrgValue v);

#endif /* ORG_OPS_H */
