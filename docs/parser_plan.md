# Parser Planning Document

## Overview

The OrgLang parser is a **Pratt parser** (Top-Down Operator Precedence parser). It consumes the flat token stream from the lexer and produces an **AST**.

Every token has two potential parsing roles:

- **NUD** (Null Denotation): How the token behaves at the **start** of an expression (prefix position).
- **LED** (Left Denotation): How the token behaves **after** a left-hand operand (infix position).

And a **Left Binding Power** (LBP) that determines how tightly it binds to the left.

## Key Design Decisions

### No Juxtaposition

OrgLang does **not** have implicit juxtaposition for function application. There is no implicit infix "apply" triggered by two adjacent operands. All operand consumption happens through explicit operator binding:

1. **Prefix operators** (NUD): `- x`, `! x`, `@ stdout`, `this (right - 1)`.
2. **Infix operators** (LED): `a + b`, `a -> b`, `a : b`.

### Identifiers as Operators

Most "operators" (`+`, `-`, `->`, `&&`, `??`, etc.) are lexed as `IDENTIFIER` tokens. The parser resolves their binding power by looking them up in a table:

1. **Hardcoded** for built-in operators (`+`, `->`, `&&`, `?`, `o`, `$`, etc.).
2. **User-defined** from `N{...}N` syntax, registered during parsing.
3. **Default**: Unknown identifiers → BP 100 (same as function call in C/Java).

### All Identifiers are Potential Prefix Operators

Since user-defined functions (`square : { right * right }`) can be used as prefix operators (`square 4`), and the parser cannot know at parse time which identifiers are functions, the parser treats **every identifier in NUD position** as a potential prefix operator with default BP 100.

This means `square 4` parses as `PrefixExpr(square, 4)`. If `square` turns out not to be callable at runtime, an error is raised then.

Additionally, the keywords `this`, `left`, and `right` are prefix operators. `this(right - 1)` → `PrefixExpr(this, GroupExpr(right - 1))`.

### Table Parsing: The Space-Separator Rule

Inside `[...]`, **space acts as the element separator**. The critical rule:

> **Structural operators used without spaces bind their operands. Non-structural identifiers used as operators do not bind inside tables — they become standalone atoms.**

Examples:

```rust
[a:1]           # One element: BindingExpr(a, 1) — no space around ':'
[a: 3]          # One element: BindingExpr(a, 3) — ':' grabs next atom
[a : 1]         # Three elements: Name(a), Name(:), Integer(1) — spaces break
[1 + 2 3]       # Four atoms: 1, +, 2, 3 — '+' is a plain identifier
[1 + 2 3]       # NOT [3 3] — no expression evaluation between atoms
[matrix.1]      # One element: DotExpr(matrix, 1) — '.' is structural
[a: (1 + 2)]    # One element: BindingExpr(a, GroupExpr(1+2))
```

### Source File as Implicit Table

Every source file is an implicit table with `;` as the separator (not space). Newlines are **not** significant.

```rust
# Source file:
a : 1;
b : 2;

# Is equivalent to:
[
a : 1
b : 2
]
```

## Binding Power Table

For **left-associative** operators: RBP == LBP.
For **right-associative** operators: RBP < LBP.

| BP  | Operators                                       | Description             | Assoc.    | Handler |
| :-- | :---------------------------------------------- | :---------------------- | :-------- | :------ |
| 900 | `@`                                             | Resource inst./infix    | Prefix/L  | NUD+LED |
| 900 | `~`, `!`, `-`, `++`, `--`                       | Unary prefix            | Prefix    | NUD     |
| 800 | `.`                                             | Dot access              | Left      | LED     |
| 750 | `?`, `??`, `?:`                                 | Selection/Error/Elvis   | Left      | LED     |
| 500 | `**`                                            | Exponentiation          | **Right** | LED     |
| 400 | `o`, `\|>`                                      | Composition/Injection   | Left      | LED     |
| 300 | `*`, `/`, `%`, `&`                              | Product/Bitwise AND     | Left      | LED     |
| 200 | `+`, `-`, `\|`, `^`, `<<`, `>>`                | Sum/Bitwise OR/XOR      | Left      | LED     |
| 150 | `=`, `<>`, `~=`, `<`, `>`, `<=`, `>=`           | Comparisons             | Left      | LED     |
| 140 | `&&`                                            | Logical AND             | Left      | LED     |
| 130 | `\|\|`                                          | Logical OR              | Left      | LED     |
| 100 | *(user-defined / default)*                      | Custom, `$`             | Left      | LED     |
| 80  | `:`, `@:`                                       | Binding/Resource def    | **Right** | LED     |
| 60  | `,`                                             | Comma (table building)  | Left      | LED     |
| 50  | `->`, `-<`, `-<>`                               | Flow/Dispatch/Join      | Left      | LED     |
| 0   | `;`                                             | Statement terminator    | N/A       | N/A     |

### Right-Associative RBP Values

| Operator | LBP | RBP (recursive parse) |
| :------- | :-- | :-------------------- |
| `**`     | 500 | 499                   |
| `:`      | 80  | 79                    |
| `@:`     | 80  | 79                    |

## The `@` Operator

`@` is a single operator with two arities and BP 800:

**Prefix** (NUD): Resource instantiation — `@stdout` → `ResourceInst(stdout)`.

**Infix** (LED): Syscall/import — `"path" @ org` → `InfixExpr("path", @, org)`.

## NUD Handlers (Prefix Position)

| Token Type   | AST Node          | Notes                                          |
| :----------- | :---------------- | :--------------------------------------------- |
| `INTEGER`    | `IntegerLiteral`  | Or `FunctionLiteral` if `{` follows adjacently |
| `DECIMAL`    | `DecimalLiteral`  |                                                |
| `RATIONAL`   | `RationalLiteral` |                                                |
| `STRING`     | `StringLiteral`   |                                                |
| `DOCSTRING`  | `StringLiteral`   | Indent-stripped                                |
| `BOOLEAN`    | `BooleanLiteral`  |                                                |
| `IDENTIFIER` | `PrefixExpr`      | Default BP 100; consumes right operand         |
| `KEYWORD`    | `PrefixExpr`      | `this`, `left`, `right` — prefix operators     |
| `AT` (`@`)   | `ResourceInst`    | Consumes next operand                          |
| `LPAREN`     | `GroupExpr`       | Parse inner expression, consume `)`            |
| `LBRACKET`   | `TableLiteral`    | Table-mode parsing until `]`                   |
| `LBRACE`     | `FunctionLiteral` | Parse body until `}`                           |

### Prefix Identifier Resolution

When `IDENTIFIER` appears in NUD:

1. Is it a **known prefix operator** (`-`, `!`, `~`, `++`, `--`)? → use its prefix BP.
2. Otherwise → use default prefix BP 100, consume right operand, produce `PrefixExpr`.

### Function Literal with Binding Powers

When `INTEGER` in NUD is immediately adjacent to `LBRACE`:

1. Record integer as Left Binding Power.
2. Parse `{ body }`.
3. If `RBRACE` is adjacent to `INTEGER`, record as Right Binding Power.
4. Produce `FunctionLiteral(lbp, body, rbp)`.

## LED Handlers (Infix Position)

| Token/Trigger          | AST Node     | Notes                               |
| :--------------------- | :----------- | :---------------------------------- |
| `IDENTIFIER` (infix)   | `InfixExpr`  | Binary op: `left OP right`          |
| `DOT` (`.`)            | `DotExpr`    | RHS can be any expression           |
| `COLON` (`:`)          | `BindingExpr`| Right-associative                   |
| `AT_COLON` (`@:`)      | `ResourceDef`| Right-associative                   |
| `COMMA` (`,`)          | `CommaExpr`  | Creates/extends tables              |
| `ELVIS` (`?:`)         | `ElvisExpr`  | Falsy coalescing                    |
| `AT` (`@`)             | `InfixExpr`  | Infix at BP 800                     |

### Comma Semantics

The `,` operator (BP 60, left-associative) creates or extends tables:

- **Atom , Atom** → new Table with both: `1, 2` → `[1 2]`
- **Table , Atom** → append to Table: `[1 2], 3` → `[1 2 3]`
- **Atom , Table** → new Table with nesting: `1, [2 3]` → `[1 [2 3]]`
- **Table , Table** → append as single element: `[1 2], [3 4]` → `[1 2 [3 4]]`

Left-associativity makes `1, 2, 3` → `((1, 2), 3)` → `[1 2 3]`.

## Parsing Algorithm

```python
def parseExpression(minBP):
    token = advance()
    left = nud(token)

    while peek().lbp > minBP:
        token = advance()
        left = led(token, left)

    return left
```

### Top-Level (Source File)

```python
def parseProgram():
    statements = []
    while peek() != EOF:
        statements.append(parseExpression(0))
        if peek() == SEMICOLON:
            advance()
    return Program(statements)
```

### Table (`[...]`) — Atom-Mode Parsing

```python
def parseTable():
    advance()  # consume '['
    elements = []
    while peek() != RBRACKET:
        elements.append(parseAtom())
    advance()  # consume ']'
    return TableLiteral(elements)

def parseAtom():
    # Parse a single "atom group" — tokens bound together without spaces.
    # Structural operators (. : @ @:) bind across atoms.
    # The exact logic depends on the lexer providing whitespace-awareness
    # or the parser checking token adjacency via position info.
    ...
```

### Function (`{...}`)

```python
def parseFunction(leadingBP):
    advance()  # consume '{'
    body = []
    while peek() != RBRACE:
        body.append(parseExpression(0))
        if peek() == SEMICOLON:
            advance()
    advance()  # consume '}'
    trailingBP = None
    if peek() == INTEGER and adjacent():
        trailingBP = advance().value
    return FunctionLiteral(leadingBP, body, trailingBP)
```

## AST Node Types

```go
type Node interface{ node() }

// Literals
type IntegerLiteral  struct { Value string }
type DecimalLiteral  struct { Value string }
type RationalLiteral struct { Numerator, Denominator string }
type StringLiteral   struct { Value string; IsDoc bool }
type BooleanLiteral  struct { Value bool }

// Expressions
type Name            struct { Value string }
type PrefixExpr      struct { Op string; Right Node }
type InfixExpr       struct { Left Node; Op string; Right Node }
type DotExpr         struct { Left Node; Key Node }
type BindingExpr     struct { Name Node; Value Node }
type ResourceDef     struct { Name Node; Table Node }
type ResourceInst    struct { Name Node }
type ElvisExpr       struct { Left Node; Right Node }
type CommaExpr       struct { Left Node; Right Node }
type GroupExpr       struct { Inner Node }
type TableLiteral    struct { Elements []Node }
type FunctionLiteral struct { LBP *int; Body []Node; RBP *int }

// Program
type Program         struct { Statements []Node }
```

## Resolved Gaps

| #  | Gap                                | Resolution                                                       |
| :- | :--------------------------------- | :--------------------------------------------------------------- |
| 1  | Infix `@` binding power           | BP 800, same operator as prefix with different arity             |
| 2  | `?` vs user identifier             | Predefined identifier with special BP; not reserved              |
| 3  | Juxtaposition                      | Not needed; all identifiers are potential prefix operators        |
| 4  | Newline significance               | Not significant; `;` separates statements                        |
| 5  | `@:` vs `@ :`                      | `@:` is always `AT_COLON` per lexer longest-match                |
| 6  | Comma semantics                    | Creates/appends tables; left-assoc builds flat lists             |
| 7  | Dot RHS                            | Can be arbitrary expression                                      |
| 8  | `$` precedence                     | BP 100 (user-defined default)                                    |
| 9  | `++`/`--` arity                    | Prefix-only                                                      |
| 10 | Atoms in tables                    | Space separates; structural ops bind when no space               |
| 11 | `->` listed as structural          | Doc bug → added to `TODO.md`                                     |
| 12 | `?` with non-table RHS             | Doc fixed by user                                                |
| 13 | `<-` in example 05                 | Typo; fixed to `<=`                                              |
| 14 | `this(x)` call syntax              | `this` is a prefix keyword; `(x)` is its right operand          |
| 15 | Negation vs exponentiation         | Keep `-` at BP 900; doc update → added to `TODO.md`             |

## Remaining Open Questions

### Atom-Mode Parser Implementation

The table atom-mode parser needs **whitespace awareness**. The lexer must provide position information (line, column) so the parser can detect whether two tokens are adjacent (no space) or separated.

**Key question**: Should the lexer emit a `WHITESPACE` token between elements inside `[...]`? Or should the parser infer separation from column gaps? The latter is simpler but requires careful position tracking.

### Identifier as Prefix vs. Infix

When an identifier appears after a left-hand expression, is it infix (LED) or does parsing stop?

- If the identifier has known infix BP → LED applies: `4 add 5` → `InfixExpr`.
- If unknown → default infix BP 100 → LED applies: `4 foo 5` → `InfixExpr`.

This means **bare identifiers are always infix** in LED position. The parser never stops mid-expression because of an unknown identifier. This is consistent with `4 add 5` working and with the default BP 100.

## Package Layout

```dot
pkg/
├── token/
│   └── token.go
├── lexer/
│   ├── lexer.go
│   └── lexer_test.go
└── parser/
    ├── parser.go           # Pratt parser core
    ├── ast.go              # AST node definitions
    ├── binding_powers.go   # BP table and lookup
    └── parser_test.go      # Table-driven tests
```

## Testing Strategy

1. **Precedence**: `1 + 2 * 3` → `1 + (2 * 3)`.
2. **Right-associativity**: `a : b : c` → `a : (b : c)`.
3. **Prefix operators**: `square 4` → `PrefixExpr(square, 4)`.
4. **Binary operators**: `4 add 5` → `InfixExpr(4, add, 5)`.
5. **Table literals (atom mode)**: `[1 2 3]` → three atoms.
6. **Table bindings**: `[a:1 b:2]` → two bindings (no space around `:`).
7. **Table with spaces around `:`**: `[a : 1]` → three atoms.
8. **Function literals**: `{ left + right }`, `600{ left ** right }601`.
9. **Prefix `@`**: `@stdout` → `ResourceInst(stdout)`.
10. **Infix `@`**: `"path" @ org` → `InfixExpr`.
11. **Resource def**: `Logger @: [next: {...}]` → `ResourceDef`.
12. **Comma building**: `1, 2, 3` → left-assoc `CommaExpr`.
13. **Flow**: `source -> transform -> sink` → left-assoc chain.
14. **Elvis**: `x ?: default` → `ElvisExpr`.
15. **Dot chaining**: `matrix.1.1` → nested `DotExpr`.
16. **Dot with expression**: `table.(1 + 1)`.
17. **Keywords as prefix**: `this (right - 1)` → `PrefixExpr(this, GroupExpr)`.
18. **All example files**: Parse `examples/*.org` and verify AST.
