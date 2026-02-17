# UTF-8 Support Planning Document

## Problem

OrgLang is defined as UTF-8, but there is currently no syntax for representing characters that cannot be typed directly from a keyboard — such as accented letters (`é`, `ñ`), CJK characters (`漢`), emoji (`🎉`), mathematical symbols (`∑`, `π`), or control characters (`\n`, `\t`, null).

This affects two contexts:

1. **String literals**: How to embed arbitrary Unicode codepoints in `"..."` and `"""..."""`.
2. **Source identifiers**: Whether non-ASCII Unicode characters can be used in names (e.g., `café`, `π`).

Additionally, OrgLang strings currently have **no escape mechanism at all** — no way to represent a newline, tab, null byte, or even the `"` character inside a string.

## The `\` (Backslash) Character

The backslash `\` is currently **completely unused** in OrgLang:

- Not in the identifier alphabet (`!$%&*-+=^~?/<>|` + letters/digits/`_`)
- Not a structural character (`@:.,;()[]{}`)
- Not mentioned anywhere in the spec

This makes it an ideal candidate for escape syntax, since it has no existing meaning to conflict with.

### Should `\` be forbidden in identifiers?

**Recommendation: Yes.** `\` should be excluded from identifiers because:

1. It creates no conflict — it's not currently allowed anyway (not in the identifier character list).
2. It prevents ambiguity: `foo\nbar` should never be parsed as an identifier.
3. It's consistent with virtually every other language.

Since `\` is not in the `!$%&*-+=^~?/<>|` set, it is **already implicitly forbidden** in identifiers. No spec change is needed.

## Alternatives for UTF-8 Character Representation

### Alternative 1: Traditional Backslash Escapes

The universally recognized approach. `\` followed by a code introduces a special character.

#### Escape Sequences

| Escape       | Meaning                          | Example               |
| :----------- | :------------------------------- | :-------------------- |
| `\n`         | Newline (LF, U+000A)            | `"line1\nline2"`      |
| `\t`         | Tab (U+0009)                    | `"col1\tcol2"`        |
| `\r`         | Carriage return (U+000D)        | `"text\r\n"`          |
| `\\`         | Literal backslash               | `"path\\file"`        |
| `\"`         | Literal double quote            | `"say \"hi\""`        |
| `\0`         | Null (U+0000)                   | `"null\0byte"`        |
| `\uXXXX`     | Unicode BMP codepoint (4 hex)   | `"\u00E9"` → `é`     |
| `\u{XXXXXX}` | Unicode codepoint (1-6 hex)     | `"\u{1F389}"` → `🎉` |

**Pros:**

- Universally familiar (C, Java, JavaScript, Rust, Python, Go, etc.)
- Covers all Unicode codepoints via `\u{...}`
- Clear and unambiguous

**Cons:**

- Verbose for common operations (e.g., paths on Windows, regex patterns)
- Requires doubling `\\` for literal backslashes

### Alternative 2: Raw Strings + Backslash Escapes

Combine backslash escapes in normal strings with a "raw string" variant that has **no escaping**.

- `"Hello\nWorld"` — escape sequences are processed.
- `r"C:\Users\file"` — raw string, no escaping, `\` is literal.

**Pros:**

- Best of both worlds: escapes when needed, raw when not.
- Familiar from Python, Rust, C#.

**Cons:**

- Adds a new string prefix syntax (`r"..."` or equivalent).
- Need to define what the raw-string delimiter looks like in OrgLang.

### Alternative 3: Named Character References

Like HTML entities. Use a distinct syntax to reference characters by name.

```rust
msg : "caf\{e-acute} is delicious";     # \{name} → character
msg : "sum: \{greek-small-letter-pi}";  # π
```

**Pros:**

- Highly readable for rare characters.
- Self-documenting.

**Cons:**

- Requires a character name database.
- Verbose for common use cases.
- Not widely used in programming languages (only Perl 6/Raku and some XML).

### Alternative 4: String Interpolation with Unicode Lookup

Since OrgLang has `$` for string substitution, Unicode characters could be provided via a built-in resource or table:

```rust
pi_char : unicode.0x03C0;              # π from a unicode table
msg : "The value of $0" $ [pi_char];   # The value of π
```

**Pros:**

- Leverages existing language mechanisms.
- No new syntax needed.

**Cons:**

- Very verbose for simple use cases like newlines.
- Doesn't solve control characters in string literals (`\n`, `\t`).
- Requires the `unicode` resource to exist.

### Alternative 5: Multiline Strings Cover Most Escaping Needs

OrgLang already has `"""..."""` for multiline strings, which naturally embed newlines. For other special characters, rely on direct UTF-8 encoding in the source file:

```rust
msg : """
    This string has newlines.
    And this is café — typed directly.
    And 🎉 — also typed directly.
""";
```

**Pros:**

- Minimal syntax addition.
- Modern editors can input any Unicode character.

**Cons:**

- Cannot represent **control characters** (tab, null, carriage return) explicitly.
- Cannot embed a `"` or `"""` inside a string.
- Source files must be edited with Unicode-aware editors.

## Recommendation

### For OrgLang: Alternative 1 + 2 (Backslash Escapes with Raw Strings)

| Feature | Syntax | Example |
| :------ | :----- | :------ |
| Standard escapes | `\n`, `\t`, `\r`, `\\`, `\"`, `\0` | `"line1\nline2"` |
| Unicode (BMP) | `\uXXXX` | `"\u00E9"` → `é` |
| Unicode (full) | `\u{X...}` | `"\u{1F389}"` → `🎉` |
| Raw strings | TBD prefix | No escape processing |

This is the most practical and user-friendly approach because:

1. **Familiarity**: Every developer knows `\n`, `\t`, `\"`.
2. **Full coverage**: `\u{...}` handles all 1,114,112 Unicode codepoints.
3. **Practicality**: Raw strings avoid escaping pain for paths, regex, etc.

### Raw String Syntax Options for OrgLang

Since OrgLang strings already use `"..."` and `"""..."""`, possible raw-string syntaxes:

| Option | Regular | Raw |
| :----- | :------ | :-- |
| Prefix `r` | `"hello\n"` | `r"hello\n"` → literal `hello\n` |
| Prefix `'` (single quotes) | `"hello\n"` | `'hello\n'` → literal `hello\n` |
| Backtick | `"hello\n"` | `` `hello\n` `` → literal `hello\n` |

**Prefix `r`** is the most familiar. But since `r` is a valid identifier start, `r"hello"` might be ambiguous: is it the identifier `r` followed by the string `"hello"`, or a raw string?

**Single quotes** avoid the ambiguity entirely (since `'` is not used for anything in OrgLang) but introduce a new delimiter type.

**Backticks** are also unused but visually less clear.

### Identifiers and Unicode

For **source code identifiers**, two options:

1. **ASCII-only identifiers**: Keep the current spec. Non-ASCII chars are only available in strings via `\u{...}`. This is simpler.

2. **Unicode identifiers**: Allow Unicode letters in identifiers (e.g., `café`, `π`, `Σ`). The rule would be: any Unicode letter (`\p{Letter}`) or `_` can start an identifier, and Unicode letters, digits, and the current symbol set can continue it. This follows Java, Rust, Python 3, and Swift.

## Summary of Decisions Needed

1. **Adopt `\` for string escapes?** (Recommended: yes)
2. **Which escape sequences?** (Recommended: `\n \t \r \\ \" \0 \uXXXX \u{...}`)
3. **Add raw strings?** (Recommended: yes, but syntax needs choosing)
4. **Allow Unicode identifiers?** (Question for language design taste)
