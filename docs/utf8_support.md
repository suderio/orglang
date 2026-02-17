# UTF-8 Support

## Decisions

1. **Backslash escapes adopted**: `\` is the escape character in string literals.
2. **Single quotes for raw strings**: `'...'` strings have no escape processing.
3. **Unicode identifiers allowed**: Characters with Unicode properties `Letter`, `Symbol`, and `Number` are allowed in identifiers, dependent on editor support — no special syntax to represent them. `Punctuation` is explicitly excluded.

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

1. **String escape processing**: When scanning `"..."`, detect `\` and process the escape table above. Emit an error for unknown `\X` sequences.
2. **Raw string scanning**: When scanning `'...'`, consume characters literally until the closing `'` (or `'''`).
3. **Unicode identifier support**: Character classification must use Unicode properties — a character is an identifier character if it matches `\p{Letter}` | `\p{Symbol}` | `\p{Number}` | `_`, and is NOT a structural character or `\p{Punctuation}`.

## Parser Impact

Minimal. The parser receives tokens as before — the escape processing is fully handled by the lexer. The only new consideration:

- `STRING` tokens may now come from either `"..."` or `'...'` — the token should carry a flag or type indicating whether it was raw.
