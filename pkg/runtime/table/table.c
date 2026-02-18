#include "table.h"
#include <string.h>

/* ---- Hashing ---- */

/* FNV-1a hash for string data */
static uint32_t fnv1a(const char *data, size_t len) {
  uint32_t h = 2166136261u;
  for (size_t i = 0; i < len; i++) {
    h ^= (uint8_t)data[i];
    h *= 16777619u;
  }
  return h;
}

uint32_t org_hash_value(OrgValue key) {
  if (ORG_IS_SMALL(key)) {
    /* Integer key: mix the bits */
    uint64_t k = (uint64_t)key;
    k = (k ^ (k >> 16)) * 0x45d9f3b;
    k = (k ^ (k >> 16)) * 0x45d9f3b;
    k ^= k >> 16;
    return (uint32_t)k;
  }
  if (ORG_IS_PTR(key) && org_get_type(key) == ORG_TYPE_STRING) {
    OrgString *s = (OrgString *)ORG_GET_PTR(key);
    return fnv1a(s->data, s->byte_len);
  }
  return 0; /* Non-hashable: shouldn't be used as key */
}

int org_key_equal(OrgValue a, OrgValue b) {
  /* Same tagged value → always equal */
  if (a == b)
    return 1;

  /* Both small ints: already compared above by value */
  if (ORG_IS_SMALL(a) && ORG_IS_SMALL(b))
    return 0;

  /* Both strings: compare contents */
  if (ORG_IS_PTR(a) && ORG_IS_PTR(b) && org_get_type(a) == ORG_TYPE_STRING &&
      org_get_type(b) == ORG_TYPE_STRING) {
    OrgString *sa = (OrgString *)ORG_GET_PTR(a);
    OrgString *sb = (OrgString *)ORG_GET_PTR(b);
    if (sa->byte_len != sb->byte_len)
      return 0;
    return memcmp(sa->data, sb->data, sa->byte_len) == 0;
  }

  return 0;
}

/* ---- Internal helpers ---- */

#define TABLE_INITIAL_CAP 8
#define TABLE_LOAD_PERCENT 75

static OrgTableEntry *alloc_entries(Arena *arena, uint32_t capacity) {
  size_t size = sizeof(OrgTableEntry) * capacity;
  OrgTableEntry *entries = (OrgTableEntry *)arena_alloc(arena, size, 8);
  if (!entries)
    return NULL;
  for (uint32_t i = 0; i < capacity; i++) {
    entries[i].key = ORG_UNUSED;
    entries[i].value = ORG_UNUSED;
    entries[i].hash = 0;
    entries[i]._pad = 0;
  }
  return entries;
}

static OrgTable *get_table(OrgValue v) { return (OrgTable *)ORG_GET_PTR(v); }

/* Find slot for a key. Returns the slot index.
 * If found: entries[slot].key != ORG_UNUSED and matches.
 * If not found: entries[slot].key == ORG_UNUSED (first empty slot). */
static uint32_t find_slot(OrgTableEntry *entries, uint32_t capacity,
                          OrgValue key, uint32_t hash) {
  uint32_t mask = capacity - 1;
  uint32_t idx = hash & mask;
  for (;;) {
    OrgValue k = entries[idx].key;
    if (ORG_IS_UNUSED(k))
      return idx;
    if (entries[idx].hash == hash && org_key_equal(k, key))
      return idx;
    idx = (idx + 1) & mask;
  }
}

/* Grow the hash table to double capacity. Rehash all entries. */
static int table_grow(Arena *arena, OrgTable *t) {
  uint32_t new_cap = t->capacity * 2;
  OrgTableEntry *new_entries = alloc_entries(arena, new_cap);
  if (!new_entries)
    return 0;

  /* Rehash existing entries */
  for (uint32_t i = 0; i < t->capacity; i++) {
    OrgTableEntry *e = &t->entries[i];
    if (ORG_IS_UNUSED(e->key))
      continue;
    uint32_t slot = find_slot(new_entries, new_cap, e->key, e->hash);
    new_entries[slot] = *e;
  }

  t->entries = new_entries;
  t->capacity = new_cap;
  /* Old entries memory is abandoned — Arena reclaims in bulk */
  return 1;
}

/* Check if a key is valid (String or SmallInt) */
static int is_valid_key(OrgValue key) {
  if (ORG_IS_SMALL(key))
    return 1;
  if (ORG_IS_PTR(key) && org_get_type(key) == ORG_TYPE_STRING)
    return 1;
  return 0;
}

/* ---- Public API ---- */

OrgValue org_table_new(Arena *arena) {
  return org_table_new_sized(arena, TABLE_INITIAL_CAP);
}

OrgValue org_table_new_sized(Arena *arena, uint32_t expected) {
  /* Round up to next power of 2 */
  uint32_t cap = TABLE_INITIAL_CAP;
  while (cap < expected)
    cap *= 2;

  OrgTable *t = (OrgTable *)arena_alloc(arena, sizeof(OrgTable), 8);
  if (!t)
    return ORG_ERROR;

  t->header.type = ORG_TYPE_TABLE;
  t->header.flags = 0;
  t->header._pad = 0;
  t->header.size = (uint32_t)sizeof(OrgTable);
  t->count = 0;
  t->capacity = cap;
  t->next_index = 0;
  t->_pad = 0;

  t->entries = alloc_entries(arena, cap);
  if (!t->entries)
    return ORG_ERROR;

  return ORG_TAG_PTR_VAL(t);
}

OrgValue org_table_set(Arena *arena, OrgValue table, OrgValue key,
                       OrgValue value) {
  if (!ORG_IS_PTR(table) || org_get_type(table) != ORG_TYPE_TABLE)
    return ORG_ERROR;
  if (!is_valid_key(key))
    return ORG_ERROR;

  OrgTable *t = get_table(table);

  /* Check load factor before insertion */
  if ((t->count + 1) * 100 > t->capacity * TABLE_LOAD_PERCENT) {
    if (!table_grow(arena, t))
      return ORG_ERROR;
  }

  uint32_t hash = org_hash_value(key);
  uint32_t slot = find_slot(t->entries, t->capacity, key, hash);

  if (ORG_IS_UNUSED(t->entries[slot].key)) {
    t->count++;
  }

  t->entries[slot].key = key;
  t->entries[slot].value = value;
  t->entries[slot].hash = hash;

  return table;
}

OrgValue org_table_push(Arena *arena, OrgValue table, OrgValue value) {
  if (!ORG_IS_PTR(table) || org_get_type(table) != ORG_TYPE_TABLE)
    return ORG_ERROR;

  OrgTable *t = get_table(table);
  OrgValue key = ORG_TAG_SMALL_INT(t->next_index);
  t->next_index++;

  return org_table_set(arena, table, key, value);
}

OrgValue org_table_get(OrgValue table, OrgValue key) {
  if (!ORG_IS_PTR(table) || org_get_type(table) != ORG_TYPE_TABLE)
    return ORG_ERROR;
  if (!is_valid_key(key))
    return ORG_ERROR;

  OrgTable *t = get_table(table);
  uint32_t hash = org_hash_value(key);
  uint32_t mask = t->capacity - 1;
  uint32_t idx = hash & mask;

  for (;;) {
    OrgTableEntry *e = &t->entries[idx];
    if (ORG_IS_UNUSED(e->key))
      return ORG_ERROR; /* Not found */
    if (e->hash == hash && org_key_equal(e->key, key))
      return e->value;
    idx = (idx + 1) & mask;
  }
}

OrgValue org_table_get_cstr(OrgValue table, const char *name) {
  if (!ORG_IS_PTR(table) || org_get_type(table) != ORG_TYPE_TABLE)
    return ORG_ERROR;

  /* Build a temporary key for lookup without arena allocation.
   * We use the hash of the raw string bytes and compare against
   * string entries directly. */
  OrgTable *t = get_table(table);
  size_t name_len = strlen(name);
  uint32_t hash = fnv1a(name, name_len);
  uint32_t mask = t->capacity - 1;
  uint32_t idx = hash & mask;

  for (;;) {
    OrgTableEntry *e = &t->entries[idx];
    if (ORG_IS_UNUSED(e->key))
      return ORG_ERROR;
    if (e->hash == hash && ORG_IS_PTR(e->key) &&
        org_get_type(e->key) == ORG_TYPE_STRING) {
      OrgString *s = (OrgString *)ORG_GET_PTR(e->key);
      if (s->byte_len == (uint32_t)name_len &&
          memcmp(s->data, name, name_len) == 0) {
        return e->value;
      }
    }
    idx = (idx + 1) & mask;
  }
}

OrgValue org_table_has(OrgValue table, OrgValue key) {
  OrgValue v = org_table_get(table, key);
  return ORG_IS_ERROR(v) ? ORG_FALSE : ORG_TRUE;
}

uint32_t org_table_count(OrgValue table) {
  if (!ORG_IS_PTR(table) || org_get_type(table) != ORG_TYPE_TABLE)
    return 0;
  return get_table(table)->count;
}
