#ifndef ORG_TABLE_H
#define ORG_TABLE_H

#include "../core/values.h"

/*
 * OrgTable — Hash+Array hybrid data structure.
 *
 * Tables are OrgLang's universal container: arrays, maps, and scopes.
 * They support both integer-indexed (auto-assigned 0,1,2,...) and
 * string-keyed access.
 *
 * Implementation: Open-addressing hash table with linear probing.
 * Load factor threshold: 75% triggers resize (2x).
 */

typedef struct OrgTableEntry {
  OrgValue key;   /* String or SmallInt key (ORG_UNUSED = empty slot) */
  OrgValue value; /* The stored value */
  uint32_t hash;  /* Cached hash of the key */
  uint32_t _pad;
} OrgTableEntry;

typedef struct OrgTable {
  OrgObject header;
  uint32_t count;      /* Number of live entries */
  uint32_t capacity;   /* Total slots (always power of 2) */
  uint32_t next_index; /* Next auto-index for positional elements */
  uint32_t _pad;
  OrgTableEntry *entries; /* Arena-allocated hash table array */
} OrgTable;

/* ---- Construction ---- */

/* Create a new empty table with default capacity (8). */
OrgValue org_table_new(Arena *arena);

/* Create a new empty table with a hint for expected size. */
OrgValue org_table_new_sized(Arena *arena, uint32_t expected);

/* ---- Insertion ---- */

/*
 * Set a key-value pair. If the key already exists, its value is updated.
 * Key must be a String or SmallInt. Returns ORG_ERROR on invalid key.
 */
OrgValue org_table_set(Arena *arena, OrgValue table, OrgValue key,
                       OrgValue value);

/*
 * Append a positional value (auto-assigns the next integer index).
 * Equivalent to org_table_set(arena, table, next_index++, value).
 */
OrgValue org_table_push(Arena *arena, OrgValue table, OrgValue value);

/* ---- Lookup ---- */

/*
 * Get a value by key. Returns ORG_ERROR if key not found.
 * Key must be a String or SmallInt.
 */
OrgValue org_table_get(OrgValue table, OrgValue key);

/*
 * Get a value by string name (convenience for scope lookups).
 * Equivalent to org_table_get(table, org_make_string(name)).
 */
OrgValue org_table_get_cstr(OrgValue table, const char *name);

/*
 * Check if a key exists in the table.
 * Returns ORG_TRUE or ORG_FALSE.
 */
OrgValue org_table_has(OrgValue table, OrgValue key);

/* ---- Size ---- */

/* Get the number of entries in the table. */
uint32_t org_table_count(OrgValue table);

/* ---- Internal ---- */

/* Compute hash for a key (String or SmallInt). */
uint32_t org_hash_value(OrgValue key);

/* Compare two keys for equality (String content or SmallInt value). */
int org_key_equal(OrgValue a, OrgValue b);

#endif /* ORG_TABLE_H */
