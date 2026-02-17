# UTF-8 Support

## Decisions

1. **Backslash escapes adopted**: `\` is the escape character in string literals.
2. **Single quotes for raw strings**: `'...'` strings have no escape processing.
3. **Unicode identifiers allowed**: Characters with Unicode properties `Letter`, `Symbol`, and `Number` are allowed in identifiers, dependent on editor support — no special syntax to represent them. `Punctuation` is explicitly excluded.
4. **Strings are tables of Unicode codepoints**: Each element of a string table is a single Unicode codepoint (U+0000 to U+10FFFF), not a byte or a grapheme cluster.

## String Escape Sequences

Inside `"..."` and `"""..."""` strings, the following escape sequences are recognized:

| Escape       | Meaning                        | Example               |
| :----------- | :----------------------------- | :-------------------- |
| `\n`         | Newline (LF, U+000A)           | `"line1\nline2"`      |
| `\t`         | Tab (U+0009)                   | `"col1\tcol2"`        |
| `\r`         | Carriage return (U+000D)       | `"text\r\n"`          |
| `\\`         | Literal backslash              | `"path\\file"`        |
| `\"`         | Literal double quote           | `"say \"hi\""`        |
| `\0`         | Null (U+0000)                  | `"null\0byte"`        |
| `\uXXXX`     | Unicode BMP codepoint (4 hex)  | `"\u00E9"` → `é`      |
| `\u{XXXXXX}` | Unicode codepoint (1-6 hex)    | `"\u{1F389}"` → `🎉`  |

Any other `\X` sequence is an error.

## Raw Strings

Raw strings use **single quotes** (`'...'`) and have **no escape processing**. Every character between the quotes is literal:

```rust
path : 'C:\Users\file';           # Literal backslash, no escaping
regex : '(\d+)\s+(\w+)';          # No need to double-escape
```

For multiline raw strings, use triple single quotes (`'''...'''`):

```rust
raw_block : '''
    This is raw.
    \n is literal — not a newline.
''';
```

### Raw String Constraints

- Cannot contain a literal single quote `'` (no escape mechanism inside raw strings).
- Triple-quoted raw strings (`'''...'''`) cannot contain `'''`.
- The same indentation stripping rules as `"""..."""` docstrings apply to `'''...'''`.

## Strings as Tables of Codepoints

Strings in OrgLang are semantically **tables of Unicode codepoints**. Each element is one codepoint (not a byte, not a grapheme cluster).

### Why Codepoints (not bytes or grapheme clusters)

| Level | Unit | `"café"` size | `"café".3` | Complexity |
| :--- | :--- | :--- | :--- | :--- |
| Bytes | UTF-8 code unit | 5 (é = 2 bytes) | `0xC3` 💥 broken | Trivial |
| **Codepoints** | Unicode scalar value | **4** | **`é`** ✅ | Moderate |
| Grapheme clusters | Visual "character" | 4 | `é` ✅ | High (needs UAX #29) |

**Decision**: Codepoints are the pragmatic middle ground — intuitive for most text, consistent with Go (`rune`), Python 3, and Rust (`char`), without requiring complex Unicode segmentation.

### Consequences

- **Indexing**: `s.0` returns the first codepoint as a single-character string.
- **Length/Size**: `"café" + 0` = 4 (codepoint count, used for arithmetic coercion).
- **Iteration**: `"café" -> f` sends 4 codepoints to `f`, one at a time.
- **Equality**: `"é"` (U+00E9, precomposed) and `"e\u0301"` (decomposed) are **different** strings with different lengths (1 vs 2). No implicit normalization.

### Edge Cases

| Expression | Result | Reason |
| :--- | :--- | :--- |
| `"🇧🇷" + 0` | 2 | Flag = 2 regional indicator codepoints |
| `"🇧🇷".0` | `"🇷"` (half a flag) | Each codepoint is independent |
| `"e\u0301" + 0` | 2 | Decomposed é = `e` + combining accent |
| `"👨‍👩‍👧‍👦" + 0` | 7 | Family emoji = 7 codepoints (4 people + 3 ZWJ) |

> [!NOTE]
> Grapheme-aware operations (e.g., iterating by visual characters) can be provided by the standard library in the future, but the core string type operates on codepoints.

## Identifiers and Unicode

Identifiers may contain characters from three Unicode general categories, plus `_` and the ASCII operator symbol set:

### Allowed Unicode Categories

| Category | Subcategories | Examples | Rationale |
| :--- | :--- | :--- | :--- |
| `\p{Letter}` (L) | Lu, Ll, Lt, Lm, Lo | `é`, `π`, `Σ`, `漢`, `名` | Natural extension of ASCII letters |
| `\p{Symbol}` (S) | Sm, Sc, Sk, So | `∑`, `√`, `∞`, `≤`, `€`, `£`, `→` | Natural extension of ASCII operator symbols (`!$%&*-+=^~?/<>\|`) |
| `\p{Number}` (N) | Nd, Nl, No | `٠`–`٩`, `Ⅱ`, `½`, `²` | Natural extension of ASCII digits |

### Excluded: `\p{Punctuation}` (P)

Punctuation is **explicitly excluded** because it contains OrgLang's structural characters:

- `Ps`/`Pe` (Open/Close): `(`, `)`, `[`, `]`, `{`, `}`, `«`, `⟨` — confusable with delimiters
- `Pi`/`Pf` (Quotes): `"`, `"`, `'`, `'` — confusable with string delimiters `"` and `'`
- `Po` (Other): `.`, `,`, `:`, `;`, `@`, `#` — includes ALL current structural characters
- `Pd` (Dash): `–` (en dash), `—` (em dash) — visually similar to `-` but different codepoints, source of invisible bugs

The sole exception is `_` (U+005F), which is `Pc` (Connector Punctuation) and is explicitly whitelisted.

### Identifier Rules

- **Start**: `\p{Letter}` | `\p{Symbol}` | `\p{Number}` | `_`
- **Continue**: All of the above, plus ASCII digits `0-9` (redundant with `\p{Number}` but explicit)
- **Excluded from identifiers**: `\p{Punctuation}` (except `_`), structural characters, whitespace

> [!NOTE]
> Since OrgLang integer literals only use ASCII `0-9`, non-ASCII digits and number-like characters (e.g., `Ⅱ`, `½`) cause no ambiguity at identifier start position.

### Examples

```rust
café : "coffee";          # \p{Letter}
π : 3.14159;              # \p{Letter} (Greek)
∑ : { left + right };     # \p{Symbol} (Sm, Math)
€price : 42;              # \p{Symbol} (Sc, Currency)
名前 : "name";             # \p{Letter} (CJK)
x² : x * x;              # \p{Number} (No, superscript)
```

This is purely editor-dependent — OrgLang does not provide a special syntax for entering Unicode characters in identifiers. If your editor can type `∑`, you can use it.

### Characters Forbidden in Identifiers

The following characters are **structural** and cannot appear inside identifiers:

| Characters            | Role                     |
| :-------------------- | :----------------------- |
| `@`, `:`, `.`, `,`    | Structural operators     |
| `;`                   | Statement terminator     |
| `(`, `)`, `[`, `]`    | Delimiters               |
| `{`, `}`              | Function delimiters      |
| `\`                   | Escape character         |
| `'`, `"`              | String delimiters        |
| `#`                   | Comment start            |
| `\p{Punctuation}`     | Entire category excluded |
| Whitespace            | Token separator          |

## Lexer Impact

The lexer needs:

1. **String escape processing**: When scanning `"..."`, detect `\` and process the escape table above. Emit an error for unknown `\X` sequences. Store the decoded codepoints, not the raw UTF-8 bytes.
2. **Raw string scanning**: When scanning `'...'`, consume characters literally until the closing `'` (or `'''`). Store as codepoints.
3. **Unicode identifier support**: Character classification must use Unicode properties — a character is an identifier character if it matches `\p{Letter}` | `\p{Symbol}` | `\p{Number}` | `_`, and is NOT a structural character or `\p{Punctuation}`.

## Parser Impact

Minimal. The parser receives tokens as before — the escape processing is fully handled by the lexer. The only new consideration:

- `STRING` tokens may now come from either `"..."` or `'...'` — the token should carry a flag or type indicating whether it was raw.

## Runtime Impact

The runtime string representation must store **codepoints** (not bytes):

- Internal storage can use UTF-8 encoding for memory efficiency, but indexing and length operations count codepoints.
- `s.N` must skip N codepoints (O(n) in UTF-8, or O(1) with a codepoint index cache).
- Iteration (`s -> f`) decodes one codepoint at a time from the UTF-8 byte stream.
