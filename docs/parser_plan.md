# Parser Planning Document

## Overview

The OrgLang parser is a **Pratt parser** (Top-Down Operator Precedence parser). It consumes the flat token stream from the lexer and produces an **AST**.

Every token has two potential parsing roles:

- **NUD** (Null Denotation): How the token behaves at the **start** of an expression (prefix position).
- **LED** (Left Denotation): How the token behaves **after** a left-hand operand (infix position).

And a **Left Binding Power** (LBP) that determines how tightly it binds to the left.

## Key Design Decisions

### Dynamic Operator Registration

The parser maintains a **binding power table** that is updated during parsing. When the parser encounters a binding like `name : { body }`, it:

1. Inspects the function body for usage of `left` and `right` keywords.
2. Determines the arity:
   - Only `right` used → **prefix** (unary) operator.
   - Both `left` and `right` used → **binary** (infix) operator.
   - Neither used → **nullary** (value/thunk).
3. Registers `name` in the BP table with either the explicit BP from `N{...}N` syntax or the default BP 100.

This means the parser is **order-dependent**: definitions must appear before use. If an identifier is encountered that is not yet bound, it produces an **Error** node (see [Error Handling](#error-handling-for-undefined-identifiers)).

### Identifier Resolution in NUD Position

When an `IDENTIFIER` token appears in NUD position, the parser looks it up in the BP table:

1. **Known prefix operator** (hardcoded like `!`, `~`, `++`, `--` or dynamically registered like `square`): consume right operand at its prefix BP, produce `PrefixExpr`.
2. **Known infix-only operator** (like `+`, `*`, `->`, `&&`): return as `Name` — it will be used as a value/reference (e.g., right side of `|>`).
3. **Known value** (nullary binding, e.g., `x : 42`): return as `Name`.
4. **Unknown**: produce `ErrorExpr` with message (unless in assignment position — left side of `:`).

This cleanly resolves the core tension:

- `square 4`: `square` is registered as prefix at BP 100. NUD consumes `4`. → `PrefixExpr(square, 4)` ✓
- `left + right`: `left` is a keyword (Name), `+` fires as infix LED at BP 200. → `InfixExpr` ✓
- `5 |> +`: `|>` consumes `+` as a name reference (see below) → works ✓

### The `left` and `right` Keywords

`left` and `right` are implicit function parameters, not operators. They are always parsed as **Name** nodes. Their special semantics are resolved at evaluation time.

### The `this` Keyword

`this` refers to the current operator and **mirrors its arity**:

- Inside a **unary** operator (only `right` used): `this` is a prefix operator.
- Inside a **binary** operator (both `left` and `right` used): `this` is an infix operator.
- Inside a **nullary** operator: `this` is a value (Name node).

The parser determines `this`'s behavior from the enclosing function context:

```rust
# Unary — this is prefix, consumes right operand
factorial : { (right <= 1) ? [true: 1 false: (right * this(right - 1))] };

# Binary — this is infix, takes left and right
custom_op : { left this right };  # this acts as infix here
```

### Error Handling for Undefined Identifiers

If the parser encounters an identifier that is:

- **Not bound** to any value in the current scope, AND
- **Not in assignment position** (left side of `:`), AND
- **Not in selection position** (right side of `.` or `?`)

Then the parser produces an `ErrorExpr` node with a descriptive message (e.g., "undefined identifier: foo").

This is the first error type in OrgLang and aligns with the language's philosophy of errors as values.

### The `|>` and `o` Operators — Name-Reference Right Operand

The `|>` (partial application) and `o` (composition) operators are special: their **right operand is always an operator name** (a reference to a previously defined operator), not an expression to be evaluated.

```rust
add_ten : 10 |> +;              # Right side: Name(+)
inc_and_double : double o inc;   # Right side: Name(inc)
```

**Why special handling is needed**: If `+` in `10 |> +` were parsed as a normal expression, `+` (a hardcoded prefix at BP 900) would try to consume a right operand. Since `;` follows, this would be a parse error.

**Parser rule**: The LED handlers for `|>` and `o` consume the next token as a **single Name** (advancing one token and wrapping it as `Name`), rather than calling `parseExpression`.

### Table Parsing: The Space-Separator Rule

Inside `[...]`, **space acts as the element separator**:

> Structural operators used without spaces bind their operands. Non-structural identifiers (including operators like `+`) become standalone atoms inside tables.

```rust
[a:1]           # One element: BindingExpr(a, 1)
[a: 3]          # One element: BindingExpr(a, 3) — ':' grabs next atom
[a : 1]         # Three elements: Name(a), Name(:), Integer(1)
[1 + 2 3]       # Four atoms: 1, +, 2, 3
[matrix.1]      # One element: DotExpr(matrix, 1) — '.' is structural
[a: (1 + 2)]    # One element: BindingExpr(a, GroupExpr(1+2))
```

### Source File as Implicit Table

Every source file is an implicit table with `;` as separator. Newlines are **not** significant.

## Binding Power Table

For **left-associative**: RBP == LBP. For **right-associative**: RBP < LBP.

| BP  | Operators                                     | Description           | Assoc.    | Handler |
| :-- | :-------------------------------------------- | :-------------------- | :-------- | :------ |
| 900 | `@`                                           | Resource inst./infix  | Prefix/L  | NUD+LED |
| 900 | `~`, `!`, `++`, `--`                          | Unary prefix          | Prefix    | NUD     |
| 900 | `-` (prefix only)                             | Unary negation        | Prefix    | NUD     |
| 800 | `.`                                           | Dot access            | Left      | LED     |
| 750 | `?`, `??`, `?:`                               | Selection/Error/Elvis | Left      | LED     |
| 500 | `**`                                          | Exponentiation        | **Right** | LED     |
| 400 | `o`                                           | Composition           | Left      | LED*    |
| 400 | `\|>`                                         | Partial application   | Left      | LED*    |
| 300 | `*`, `/`, `%`, `&`                            | Product/Bitwise AND   | Left      | LED     |
| 200 | `+`, `-`, `\|`, `^`, `<<`, `>>`               | Sum/Bitwise OR/XOR    | Left      | LED     |
| 150 | `=`, `<>`, `~=`, `<`, `>`, `<=`, `>=`         | Comparisons           | Left      | LED     |
| 140 | `&&`                                          | Logical AND           | Left      | LED     |
| 130 | `\|\|`                                        | Logical OR            | Left      | LED     |
| 100 | *(user-defined / default)*                    | Custom, `$`           | Left      | LED     |
| 80  | `:`, `@:`                                     | Binding/Resource def  | **Right** | LED     |
| 60  | `,`                                           | Comma (table build)   | Left      | LED     |
| 50  | `->`, `-<`, `-<>`                             | Flow/Dispatch/Join    | Left      | LED     |
| 0   | `;`                                           | Statement terminator  | N/A       | N/A     |

`*` `o` and `|>` consume their right operand as a single Name token.

### Right-Associative RBP

| Operator | LBP | RBP |
| :------- | :-- | :-- |
| `**`     | 500 | 499 |
| `:`      | 80  | 79  |
| `@:`     | 80  | 79  |

## The `@` Operator

`@` has BP 800 in both roles:

- **Prefix** (NUD): `@stdout` → `ResourceInst(stdout)`.
- **Infix** (LED): `"path" @ org` → `InfixExpr("path", @, org)`.

## NUD Handlers (Prefix Position)

| Token Type   | AST Node                             | Notes                                          |
| :----------- | :----------------------------------- | :--------------------------------------------- |
| `INTEGER`    | `IntegerLiteral`                     | Or `FunctionLiteral` if `{` follows adjacently |
| `DECIMAL`    | `DecimalLiteral`                     |                                                |
| `RATIONAL`   | `RationalLiteral`                    |                                                |
| `STRING`     | `StringLiteral`                      |                                                |
| `DOCSTRING`  | `StringLiteral`                      | Indent-stripped                                |
| `RAWSTRING`  | `StringLiteral`                      | `'...'` — no escape processing                |
| `RAWDOC`     | `StringLiteral`                      | `'''...'''` — raw, indent-stripped             |
| `BOOLEAN`    | `BooleanLiteral`                     |                                                |
| `IDENTIFIER` | `PrefixExpr`, `Name`, or `ErrorExpr` | Based on BP table lookup                       |
| `this`       | `PrefixExpr`, `InfixExpr`, or `Name` | Mirrors enclosing operator arity               |
| `left`       | `Name`                               | Function parameter                             |
| `right`      | `Name`                               | Function parameter                             |
| `AT` (`@`)   | `ResourceInst`                       | Consumes next operand                          |
| `LPAREN`     | `GroupExpr`                          | Parse inner expression, consume `)`            |
| `LBRACKET`   | `TableLiteral`                       | Table-mode parsing until `]`                   |
| `LBRACE`     | `FunctionLiteral`                    | Parse body until `}`                           |

### Known Prefix-Only Operators (Hardcoded)

| Identifier | Prefix BP | Notes            |
| :--------- | :-------- | :--------------- |
| `@`        | 900       | Structural token |
| `-`        | 900       | Unary negation   |
| `!`        | 900       | Logical NOT      |
| `~`        | 900       | Bitwise NOT      |
| `++`       | 900       | Increment        |
| `--`       | 900       | Decrement        |

> `this` is **not** in this table because its arity is dynamic — it mirrors the enclosing operator. The parser tracks context to determine `this`'s behavior.

### Dynamic NUD Resolution

```python
def nud_identifier(token):
    name = token.value
    entry = bp_table.lookup(name)

    if entry is None:
        # Not in BP table — check if we're in assignment context
        # (handled by the caller when ':' follows)
        return ErrorExpr(f"undefined identifier: {name}")

    if entry.is_prefix:
        right = parseExpression(entry.prefix_bp)
        return PrefixExpr(name, right)

    # Known but not prefix (e.g., infix-only like '+')
    return Name(name)
```

### Binding Registration

When the parser encounters `name : { body }`:

```python
def register_binding(name, func_literal):
    uses_left = body_contains_keyword(func_literal.body, "left")
    uses_right = body_contains_keyword(func_literal.body, "right")

    if func_literal.lbp is not None:
        bp = func_literal.lbp
    else:
        bp = 100  # default

    if uses_left and uses_right:
        bp_table.register_infix(name, bp)
    elif uses_right:
        bp_table.register_prefix(name, bp)
    else:
        bp_table.register_value(name)  # nullary
```

### Function Literal with Binding Powers

When `INTEGER` in NUD is immediately adjacent to `LBRACE`:

1. Record integer as Left Binding Power.
2. Parse `{ body }`.
3. If `RBRACE` is adjacent to `INTEGER`, record as Right Binding Power.
4. Produce `FunctionLiteral(lbp, body, rbp)`.

## LED Handlers (Infix Position)

| Token/Trigger        | AST Node      | Notes                               |
| :------------------- | :------------ | :---------------------------------- |
| `IDENTIFIER` (infix) | `InfixExpr`   | Binary op: `left OP right`          |
| `DOT` (`.`)          | `DotExpr`     | RHS can be any expression           |
| `COLON` (`:`)        | `BindingExpr` | Right-associative; triggers reg.    |
| `AT_COLON` (`@:`)    | `ResourceDef` | Right-associative                   |
| `COMMA` (`,`)        | `CommaExpr`   | Creates/extends tables              |
| `ELVIS` (`?:`)       | `ElvisExpr`   | Falsy coalescing                    |
| `AT` (`@`)           | `InfixExpr`   | Infix at BP 800                     |

### Comma Semantics

Left-associative at BP 60:

- **Atom , Atom** → `[left right]`
- **Table , Atom** → append
- **Atom , Table** → `[left table]`
- **Table , Table** → append as element

### The `:` Binding Handler

The `:` LED handler has extra responsibility: after parsing the right side, if the right side is a `FunctionLiteral`, it calls `register_binding` to update the BP table.

```python
def led_colon(left):
    right = parseExpression(79)  # right-assoc: RBP = 80 - 1
    if isinstance(right, FunctionLiteral) and isinstance(left, Name):
        register_binding(left.value, right)
    return BindingExpr(left, right)
```

### The `|>` and `o` LED Handlers

These consume a single Name token as the right operand:

```python
def led_pipe_inject(left):
    name_token = advance()  # consume next token as-is
    return InfixExpr(left, "|>", Name(name_token.value))

def led_compose(left):
    name_token = advance()  # consume next token as-is
    return InfixExpr(left, "o", Name(name_token.value))
```

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

### Table (`[...]`) — Atom-Mode

```python
def parseTable():
    advance()  # consume '['
    elements = []
    while peek() != RBRACKET:
        elements.append(parseAtom())
    advance()  # consume ']'
    return TableLiteral(elements)
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
type ErrorExpr       struct { Message string }

// Program
type Program         struct { Statements []Node }
```

## Resolved Gaps

| #  | Gap                         | Resolution                                          |
| :- | :-------------------------- | :-------------------------------------------------- |
| 1  | Infix `@` BP                | BP 800, same as prefix                              |
| 2  | `?` vs identifier           | Predefined with special BP                          |
| 3  | Function calls              | Dynamic BP registration; definitions before use     |
| 4  | Newlines                    | Not significant; `;` separates                      |
| 5  | `@:` vs `@ :`               | Always AT_COLON                                     |
| 6  | Comma                       | Creates/appends tables; left-assoc                  |
| 7  | Dot RHS                     | Arbitrary expression                                |
| 8  | `$` BP                      | 100                                                 |
| 9  | `++`/`--`                   | Prefix-only                                         |
| 10 | Atoms in tables             | Space separates; structural ops bind when adjacent  |
| 11 | `->` structural             | Doc bug → `TODO.md`                                 |
| 12 | `?` non-table RHS           | Doc fixed                                           |
| 13 | `<-` typo                   | Fixed to `<=`                                       |
| 14 | `this` arity                | Mirrors enclosing operator (prefix/infix/nullary)   |
| 15 | Negation vs `**`            | Keep BP 900; doc update → `TODO.md`                 |
| 16 | `left`/`right`              | Name nodes (function params), never prefix          |
| 17 | `\|>` right side            | Consumed as single Name token                       |
| 18 | Undefined identifiers       | Produce `ErrorExpr` if not in binding/selection pos |

## Remaining Implementation Details

### Whitespace Awareness for Table Parsing

The lexer must provide position info (line, column) so the parser can detect token adjacency inside `[...]`.

### Forward References

Since the parser registers operators during binding (`:`), forward references produce `ErrorExpr`. This enforces a **definitions-before-use** model, consistent with the top-to-bottom table construction.

### Nested Scopes

Inside `[...]` and `{...}`, new scopes may shadow outer bindings. The BP table should support scope stacking (push/pop).

## Package Layout

```shell
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
3. **Dynamic registration**: After `square : { right ** 2 }`, `square 4` → `PrefixExpr`.
4. **Binary op from binding**: After `add : { left + right }`, `4 add 5` → `InfixExpr`.
5. **Partial application**: `10 |> +` → `InfixExpr(10, |>, Name(+))`.
6. **Composition**: `double o inc` → `InfixExpr(Name(double), o, Name(inc))`.
7. **Keywords**: `left + right` → `InfixExpr(Name(left), +, Name(right))`.
8. **`this` prefix**: `this(right - 1)` → `PrefixExpr`.
9. **Table atoms**: `[1 2 3]` → three atoms, `[a:1]` → one binding.
10. **Table space rule**: `[a : 1]` → three atoms vs `[a:1]` → one binding.
11. **Function literals**: `{ left + right }`, `600{ left ** right }601`.
12. **Prefix `@`**: `@stdout` → `ResourceInst`.
13. **Infix `@`**: `"path" @ org` → `InfixExpr`.
14. **Resource def**: `Logger @: [next: {...}]`.
15. **Comma building**: `1, 2, 3` → left-assoc `CommaExpr`.
16. **Flow**: `source -> transform -> sink`.
17. **Error on undefined**: `foo 5` before `foo` is defined → `ErrorExpr`.
18. **All example files**.
