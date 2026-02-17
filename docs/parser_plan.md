# Parser Planning Document

## Overview

The OrgLang parser is a **Pratt parser** (Top-Down Operator Precedence parser). It consumes the flat token stream from the lexer and produces an **AST**.

Every token has two potential parsing roles:

- **NUD** (Null Denotation): How the token behaves at the **start** of an expression (prefix position).
- **LED** (Left Denotation): How the token behaves **after** a left-hand operand (infix position).

And a **Left Binding Power** (LBP) that determines how tightly it binds to the left.

## Key Design Decisions

### Identifiers as Operators

Most "operators" (`+`, `-`, `->`, `&&`, `??`, etc.) are lexed as `IDENTIFIER` tokens. The parser resolves their binding power by looking them up in a table:

1. **Hardcoded** for built-in operators (`+`, `->`, `&&`, `?`, `o`, `$`, etc.).
2. **User-defined** from `N{...}N` syntax, registered during parsing.
3. **Default**: Unknown identifiers → BP 100 (same as function call in C/Java).

### The `left` and `right` Keywords

`left` and `right` are **not** operators — they are implicit function parameters available inside operator bodies. They resolve to the operands passed to the enclosing function:

- In a **binary** expression `a op b`: `left` = `a`, `right` = `b`.
- In a **unary** expression `op x`: `right` = `x`, `left` = Error.

At the parser level, `left` and `right` are parsed as simple `Name` nodes. Their special semantics are handled at evaluation/codegen time.

### The `this` Keyword

`this` is a **prefix operator** keyword. It refers to the current function and can consume a right operand (for recursion):

```rust
this(right - 1)  # PrefixExpr(this, GroupExpr(right - 1))
```

### The `|>` (Partial Application) Operator

`|>` has BP 400, left-associative. Its semantics:

- **Left operand**: Any value to inject as `left`.
- **Right operand**: Always a previously defined **operator name** (unary or binary).
- **Result**: A nullary (if the right was unary) or unary (if the right was binary) operator.

```rust
add5 : 5 |> +;       # Fix left=5 into '+', result is unary '+ 5'
add5 10;              # 15
```

Since `|>` has BP 400 and parses its right side at BP 400, the right operand `+` cannot consume anything further (there's nothing with high enough BP after it). It is naturally returned as a `Name(+)`.

### Table Parsing: The Space-Separator Rule

Inside `[...]`, **space acts as the element separator**. The critical rule:

> **Structural operators used without spaces bind their operands. Non-structural identifiers (including operators like `+`) become standalone atoms inside tables.**

Examples:

```rust
[a:1]           # One element: BindingExpr(a, 1) — no space around ':'
[a: 3]          # One element: BindingExpr(a, 3) — ':' grabs next atom
[a : 1]         # Three elements: Name(a), Name(:), Integer(1) — spaces break
[1 + 2 3]       # Four atoms: 1, +, 2, 3 — '+' is a plain identifier
[matrix.1]      # One element: DotExpr(matrix, 1) — '.' is structural
[a: (1 + 2)]    # One element: BindingExpr(a, GroupExpr(1+2))
```

### Source File as Implicit Table

Every source file is an implicit table with `;` as the separator. Newlines are **not** significant.

```rust
a : 1;     # equivalent to the content inside [a:1 b:2]
b : 2;
```

## Open Question: Function Call Mechanism

The README shows function calls without parentheses:

```rust
square : { right * right };
result : square 4;         # 16

add : { left + right };
result : 4 add 5;          # 9
```

This creates a tension in the Pratt parser:

### The Problem

If `square` is parsed as a **prefix operator** (consuming `4` to its right), then `left + right` inside function bodies would break — `left` would try to consume `+` as its prefix operand instead of `+` being an infix operator.

Traced execution of `left + right` with "all identifiers are prefix at BP 100":

```
1. left NUD → prefix BP 100, calls parseExpression(100)
2.   + NUD → prefix BP 100, calls parseExpression(100)
3.     right NUD → prefix BP 100, nothing to consume → Name(right)
4.   + consumed right → PrefixExpr(+, right)
5. left consumed PrefixExpr(+, right) → PrefixExpr(left, PrefixExpr(+, right))
```

**Result**: `PrefixExpr(left, PrefixExpr(+, right))` — completely wrong!
**Expected**: `InfixExpr(Name(left), +, Name(right))`

### The Root Cause

The infix LED handler for `+` (BP 200) never fires because `left`'s prefix NUD at BP 100 greedily consumes `+` before the Pratt loop can offer `+` as an infix operator.

### Proposed Solutions

**Solution A: Identifiers are always Names in NUD, never prefix.**
All identifiers (including `square`, `factorial`) are just `Name` nodes in NUD position. Function calls require parentheses: `square(4)`, `factorial(5)`.

This requires `LPAREN` to have a LED handler (like a function-call operator) at high BP:

- `square(4)` → Name(square), `(` LED at BP 800 → CallExpr(square, 4).
- `left + right` → Name(left), `+` LED at BP 200 → InfixExpr ✓
- `this(right - 1)` → handled specially: `this` has hardcoded prefix NUD ✓

> [!IMPORTANT]
> This would require examples like `square 4` to be written as `square(4)`. The README examples would need updating.

**Solution B: Identifiers are prefix ONLY when they have known prefix BP.**
The parser tracks which identifiers have been defined as prefix operators (via `N{...}` syntax at parse time). Only those plus hardcoded keywords (`this`, `!`, `~`, `-`, `++`, `--`, `@`) consume right operands.

- `square 4` → `square` has no prefix BP → Name, `4` is separate expression (broken!).
- `! true` → `!` has hardcoded prefix BP → PrefixExpr ✓
- `left + right` → Name, LED, Name ✓

Same problem as A: `square 4` doesn't work.

**Solution C: Differentiate prefix vs. infix at NUD time.**
An identifier in NUD checks: is the **next** token a literal, `(`, `[`, or `{` (operands that cannot be infix operators)? If so, consume as prefix. If the next token is an identifier that could be infix, don't consume.

- `square 4` → `square` sees `4` (literal, can't be infix) → prefix → PrefixExpr ✓
- `left + right` → `left` sees `+` (identifier with known infix BP 200) → don't prefix → Name ✓
- `5 |> +` → inside parseExpression(400), `+` sees `;` → don't prefix → Name ✓
- `f g` → `f` sees `g` (identifier, unknown infix BP) → ??? Ambiguous.

> [!WARNING]
> This works for known operators but is ambiguous for unknown identifier-identifiers. `f g` could be prefix or two names.

**Question for review**: Which solution should we adopt? Solution A (always require parens) is the cleanest. Solution C gives the best compatibility with existing examples but has edge cases.

## Binding Power Table

For **left-associative** operators: RBP == LBP.
For **right-associative** operators: RBP < LBP.

| BP  | Operators                                       | Description             | Assoc.    | Handler |
| :-- | :---------------------------------------------- | :---------------------- | :-------- | :------ |
| 900 | `@`                                             | Resource inst./infix    | Prefix/L  | NUD+LED |
| 900 | `~`, `!`, `++`, `--`                            | Unary prefix            | Prefix    | NUD     |
| 900 | `-` (prefix only)                               | Unary negation          | Prefix    | NUD     |
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
| `IDENTIFIER` | `Name` or `PrefixExpr` | Depends on chosen solution (A, B, or C)   |
| `this`       | `PrefixExpr`      | Hardcoded prefix; consumes right operand       |
| `left`       | `Name`            | Function parameter reference; never prefix     |
| `right`      | `Name`            | Function parameter reference; never prefix     |
| `AT` (`@`)   | `ResourceInst`    | Consumes next operand                          |
| `LPAREN`     | `GroupExpr`       | Parse inner expression, consume `)`            |
| `LBRACKET`   | `TableLiteral`    | Table-mode parsing until `]`                   |
| `LBRACE`     | `FunctionLiteral` | Parse body until `}`                           |

### Known Prefix-Only Operators

These identifiers **always** consume a right operand in NUD position:

| Identifier | Prefix BP | Notes                |
| :--------- | :-------- | :------------------- |
| `this`     | hardcoded | Keyword              |
| `@`        | 900       | Structural token     |
| `-`        | 900       | Only when prefix     |
| `!`        | 900       | Logical NOT          |
| `~`        | 900       | Bitwise NOT          |
| `++`       | 900       | Increment (prefix)   |
| `--`       | 900       | Decrement (prefix)   |

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
| `LPAREN` (`(`)         | `CallExpr`   | *Only if Solution A is chosen*      |

### Comma Semantics

The `,` operator (BP 60, left-assoc) creates or extends tables:

- **Atom , Atom** → new Table: `1, 2` → `[1 2]`
- **Table , Atom** → append: `[1 2], 3` → `[1 2 3]`
- **Atom , Table** → nest: `1, [2 3]` → `[1 [2 3]]`
- **Table , Table** → append nested: `[1 2], [3 4]` → `[1 2 [3 4]]`

Left-associativity: `1, 2, 3` → `((1, 2), 3)` → `[1 2 3]`.

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
```

Inside `[...]`, each space-separated token group is an atom. Structural operators (`:`, `.`, `@`, `@:`) bind when used without spaces.

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
type CallExpr        struct { Func Node; Arg Node }    // Only if Solution A
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

| #  | Gap                            | Resolution                                                   |
| :- | :----------------------------- | :----------------------------------------------------------- |
| 1  | Infix `@` BP                  | BP 800, same operator as prefix                              |
| 2  | `?` vs identifier              | Predefined with special BP; not reserved                     |
| 3  | Juxtaposition                  | Not supported; see Open Question                             |
| 4  | Newline significance           | Not significant; `;` separates                               |
| 5  | `@:` vs `@ :`                  | `@:` always AT_COLON                                         |
| 6  | Comma semantics                | Creates/appends; left-assoc                                  |
| 7  | Dot RHS                        | Arbitrary expression                                         |
| 8  | `$` precedence                 | BP 100                                                       |
| 9  | `++`/`--`                      | Prefix-only                                                  |
| 10 | Atoms in tables                | Space separates; structural ops bind when no space           |
| 11 | `->` as structural             | Doc bug → `TODO.md`                                          |
| 12 | `?` non-table RHS              | Doc fixed                                                    |
| 13 | `<-` typo                      | Fixed to `<=`                                                |
| 14 | `this(x)` call                 | `this` is prefix keyword                                     |
| 15 | Negation vs `**`               | Keep BP 900; doc update → `TODO.md`                          |
| 16 | `left`/`right`                 | Function params (Name nodes), not prefix operators           |
| 17 | `\|>` right side               | Always an operator name; works via BP mechanics              |

## Remaining Implementation Detail

### Atom-Mode Whitespace Awareness

The table parser needs token position info to detect adjacency. The lexer must provide line/column on each token.

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
3. **Keywords**: `left + right` → `InfixExpr(Name(left), +, Name(right))`.
4. **`this` prefix**: `this(right - 1)` → `PrefixExpr(this, GroupExpr)`.
5. **Table literals**: `[1 2 3]` → three atoms.
6. **Table bindings**: `[a:1 b:2]` → two bindings vs `[a : 1]` → three atoms.
7. **Function literals**: `{ left + right }`, `600{ left ** right }601`.
8. **Prefix `@`**: `@stdout` → `ResourceInst`.
9. **Infix `@`**: `"path" @ org` → `InfixExpr`.
10. **Resource def**: `Logger @: [next: {...}]`.
11. **Partial application**: `5 |> +` → `InfixExpr(5, |>, Name(+))`.
12. **Comma building**: `1, 2, 3` → left-assoc chain.
13. **Flow**: `source -> transform -> sink`.
14. **Elvis**: `x ?: default`.
15. **Dot chaining**: `matrix.1.1`.
16. **Dot expression**: `table.(1 + 1)`.
17. **All example files**.
