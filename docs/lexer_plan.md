# Lexer Planning Document

## Overview

The lexer is the first stage of the OrgLang compiler. It reads a UTF-8 source file and produces a flat stream of **tokens**, each annotated with its type, literal value, and source position. The parser then consumes this stream.

## Design Principles

### Identifiers are Operators

A critical design insight of OrgLang: most "operator" symbols (`+`, `-`, `*`, `<`, `>`, `!`, `~`, `&`, `|`, `^`, `=`, `%`, `?`, `/`, `$`) are **valid identifier characters**. This means that tokens like `->`, `++`, `<=`, `&&`, `??`, `**` are just ordinary identifiers from the lexer's perspective. The parser (and its binding power table) gives them operator semantics.

### Structural Characters

Only a small set of characters are **structural** — they break identifiers and do not require surrounding whitespace:

| Character(s)                | Role                  |
| :-------------------------- | :-------------------- |
| `@`  `:` `.`  `,`           | Structural operators  |
| `(` `)` `[` `]` `{` `}`     | Delimiters            |
| `;`                         | Statement terminator  |
| `#`                         | Comment introducer    |
| `"` (double quote)          | String delimiter      |

Any multi-character operator that contains a structural character is **language-defined** and must be handled by the lexer (e.g., `@:`, `?:`). Any operator composed entirely of non-structural characters (e.g., `->`, `|>`, `++`) is just an identifier.

### The `/` Exception

The `/` character is a valid identifier character, but it also forms **rational literals** (e.g., `1/2`). The lexer must detect the pattern `INTEGER/INTEGER` (no whitespace) and emit a single `RATIONAL` token instead of three separate tokens.

## Token Types

### Literals

| Token Type   | Pattern                                       | Examples                |
| :----------- | :-------------------------------------------- | :---------------------- |
| `INTEGER`    | Optional sign glued to `[0-9]+`               | `42`, `-7`, `+3`        |
| `DECIMAL`    | Optional sign glued to `[0-9]+.[0-9]+`        | `3.14`, `-0.001`        |
| `RATIONAL`   | `INTEGER/INTEGER` (no spaces)                 | `1/2`, `-3/4`           |
| `STRING`     | `"..."`                                       | `"hello"`               |
| `DOCSTRING`  | `"""..."""` (multiline, strips common indent) | `"""\n  a\n  b\n"""`    |
| `BOOLEAN`    | `true` or `false`                             | `true`, `false`         |

### Identifiers & Keywords

| Token Type   | Description                                                             |
| :----------- | :---------------------------------------------------------------------- |
| `IDENTIFIER` | Starts with `[a-zA-Z_!$%&*+\-=^~?/<>\|]`, continues with those + digits |
| `KEYWORD`    | An identifier matching a reserved word                                  |

Reserved keywords: `true`, `false`, `this`, `left`, `right`.

> `resource` is **no longer a keyword**. Resource definition uses the `@:` operator.

### Structural Delimiters

| Token       | Symbol |
| :---------- | :----- |
| `LPAREN`    | `(`    |
| `RPAREN`    | `)`    |
| `LBRACKET`  | `[`    |
| `RBRACKET`  | `]`    |
| `LBRACE`    | `{`    |
| `RBRACE`    | `}`    |
| `SEMICOLON` | `;`    |

### Structural Operators

These are the **only** operator tokens the lexer emits as distinct types. They are composed of characters that cannot appear in identifiers:

| Token       | Symbol | Notes                                         |
| :---------- | :----- | :-------------------------------------------- |
| `AT`        | `@`    | Resource instantiation (prefix)               |
| `AT_COLON`  | `@:`   | Resource definition                           |
| `COLON`     | `:`    | Binding                                       |
| `DOT`       | `.`    | Dot access                                    |
| `COMMA`     | `,`    | Table construction                            |

#### Compound Structural Operators

These are multi-character sequences that **contain** a structural character mixed with an identifier character:

| Token       | Symbol | Notes                                        |
| :---------- | :----- | :------------------------------------------- |
| `ELVIS`     | `?:`   | `?` (identifier char) + `:` (structural)     |

> The lexer detects `?:` when scanning: if a standalone `?` identifier is immediately followed by `:` → emit `ELVIS` instead of `IDENTIFIER(?)` + `COLON`.

#### Extended Assignment Operators

All start with `:` (structural), so they are language-defined. Reserved but **not yet implemented**:

`:+` `:-` `:*` `:/` `:%` `:>>` `:<<` `:&` `:^` `:|` `:~`

> The lexer handles these by peeking after `:`. If the next char is an assignment operator symbol, emit the compound token. Otherwise emit `COLON`.

### Non-Structural "Operators" (Identifiers)

The following are **not** special lexer tokens. They are just `IDENTIFIER`s that the parser gives meaning to via binding powers:

```rust
->  -<  -<>  |>  ++  --  **  
+  -  *  %  =  <  >  <=  >=  <>  ~=
<<  >>  &&  ||  !  ~  &  |  ^
??  o
```

### Special Tokens

| Token     | Description           |
| :-------- | :-------------------- |
| `EOF`     | End of file           |
| `ILLEGAL` | Unrecognized sequence |

## Lexer Rules & Edge Cases

### 1. Comments

- **Single-line**: `#` to end-of-line. Discard entirely.
- **Block**: `###` at **column 1** opens, next `###` at column 1 closes. Discard.

### 2. Whitespace

Spaces, tabs, newlines: consumed as separators, never emitted. Significant only for sign gluing and the `###` column rule.

### 3. Sign Gluing (Numbers)

When `+` or `-` is encountered:

1. If the **previous token** was an operator, delimiter, or start-of-file (prefix position), **and** the next character (no whitespace) is a digit → consume as part of a numeric literal.
2. Otherwise → it is part of an identifier scan.

### 4. Rational Literals

After producing an `INTEGER`, if the **very next character** (no whitespace) is `/` followed immediately by a digit → consume and produce a single `RATIONAL` token.

### 5. Decimal Disambiguation

- `1.0` → `DECIMAL`. Digits on both sides of the dot.
- `1.` → `INTEGER(1)` + `DOT`. No digit after the dot.
- `.5` → `DOT` + `INTEGER(5)`. Dot is structural, breaks the scan.

### 6. Binding Power Adjacency (`50{...}60`)

When an `INTEGER` is immediately followed by `{` (no whitespace), the lexer emits the `INTEGER` and `LBRACE` as adjacent tokens. The parser uses position information to detect this. Similarly for `}` immediately followed by a digit.

### 7. String Escaping

In `"..."` strings: `\"`, `\\`, `\n`, `\t`. Other escapes TBD.

In `"""..."""` strings: raw content, no escapes. Strip common leading whitespace and surrounding blank lines.

### 8. Structural Characters Break Identifiers

When scanning an identifier, any structural character (`@`, `:`, `.`, `,`, `;`, `(`, `)`, `[`, `]`, `{`, `}`, `"`, `#`) immediately terminates the identifier.

Example: `x:42` → `IDENTIFIER(x)` `COLON` `INTEGER(42)`.

### 9. The `@:` and `?:` Compound Rules

- After seeing `@`, peek: if `:` follows → emit `AT_COLON`. Else → emit `AT`.
- When scanning an identifier, if it consists of exactly `?` and the next char is `:` → emit `ELVIS`. (If the identifier is longer, e.g., `isValid?`, then `:` just terminates the identifier normally and emits `COLON` separately.)

## Token Structure

```go
type Token struct {
    Type    TokenType
    Literal string    // Raw text from source
    Line    int       // 1-indexed
    Column  int       // 1-indexed
}
```

## Package Layout

```dot
pkg/
├── token/
│   └── token.go       # TokenType enum, Token struct, keyword map
└── lexer/
    ├── lexer.go        # Lexer struct, NextToken(), helpers
    └── lexer_test.go   # Table-driven tests
```

## Testing Strategy

1. **Basic tokens**: Each delimiter and structural operator in isolation.
2. **Identifiers as operators**: `->`, `++`, `<=`, `&&`, `??` lex as single `IDENTIFIER` tokens.
3. **Sign gluing**: `-42` (one token) vs `x - 42` (three tokens).
4. **Rational vs division**: `1/2` → `RATIONAL` vs `1 / 2` → `IDENTIFIER`, `IDENTIFIER(/)`, `INTEGER`.
5. **Structural breaking**: `x:42`, `f(x)`, `list.0`, `@stdout`.
6. **Compound structural**: `@:` → `AT_COLON`, `?:` → `ELVIS`, `:+` → extended assignment.
7. **Comments**: Single-line, block, nested content.
8. **Strings**: Escapes, multiline, empty strings, unterminated.
9. **Edge cases**: Empty input, only comments, binding power adjacency.
10. **Integration**: Lex all `examples/*.org` files, verify no `ILLEGAL` tokens.
