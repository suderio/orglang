# UTF-8 Support

## Decisions

1. **Backslash escapes adopted**: `\` is the escape character in string literals.
2. **Single quotes for raw strings**: `'...'` strings have no escape processing.
3. **Unicode identifiers allowed**: Any Unicode letter can be used in identifiers, dependent on editor support — no special syntax to represent them.

## String Escape Sequences

Inside `"..."` and `"""..."""` strings, the following escape sequences are recognized:

| Escape       | Meaning                        | Example               |
| :----------- | :----------------------------- | :-------------------- |
| `\n`         | Newline (LF, U+000A)          | `"line1\nline2"`      |
| `\t`         | Tab (U+0009)                  | `"col1\tcol2"`        |
| `\r`         | Carriage return (U+000D)      | `"text\r\n"`          |
| `\\`         | Literal backslash             | `"path\\file"`        |
| `\"`         | Literal double quote          | `"say \"hi\""`        |
| `\0`         | Null (U+0000)                 | `"null\0byte"`        |
| `\uXXXX`     | Unicode BMP codepoint (4 hex) | `"\u00E9"` → `é`     |
| `\u{XXXXXX}` | Unicode codepoint (1-6 hex)   | `"\u{1F389}"` → `🎉` |

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

Identifiers may contain **Unicode letters** (any character with the Unicode `Letter` property). The rules:

- **Start**: Any Unicode letter or `_`.
- **Continue**: Unicode letters, digits (`0-9`), `_`, and the operator symbol set (`!$%&*-+=^~?/<>|`).

Examples of valid identifiers:

```rust
café : "coffee";
π : 3.14159;
Σ : { left + right };
名前 : "name";
```

This is purely editor-dependent — OrgLang does not provide a special syntax for entering Unicode characters in identifiers. If your editor can type `π`, you can use it.

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
| Whitespace            | Token separator          |

## Lexer Impact

The lexer needs:

1. **String escape processing**: When scanning `"..."`, detect `\` and process the escape table above. Emit an error for unknown `\X` sequences.
2. **Raw string scanning**: When scanning `'...'`, consume characters literally until the closing `'` (or `'''`).
3. **Unicode identifier support**: The character classification for "can this continue an identifier?" must use Unicode properties (`\p{Letter}`) instead of ASCII-only letter checks.

## Parser Impact

Minimal. The parser receives tokens as before — the escape processing is fully handled by the lexer. The only new consideration:

- `STRING` tokens may now come from either `"..."` or `'...'` — the token should carry a flag or type indicating whether it was raw.
