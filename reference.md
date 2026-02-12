# OrgLang Reference Manual

## Introduction

## Notation

## Lexical analysis

Lexical analysis is the first stage of the OrgLang compiler. An OrgLang program is read as a sequence of characters, which are then grouped into meaningful units called **tokens**.

The lexer (or tokenizer) performs this transformation by scanning the source text from beginning to end. It identifies various types of tokens:
- **Identifiers and Keywords**: Symbolic names for variables and functions.
- **Literals**: Constant values such as numbers and strings.
- **Operators**: Symbols representing computations or flows.
- **Delimiters**: Structural symbols like parentheses and semicolons.

### Character encoding

OrgLang source files are expected to be encoded in **UTF-8**. While the current implementation primarily focuses on the ASCII subset for structural elements and identifiers, UTF-8 support ensures that strings and comments can contain any Unicode character.

### Line structure

#### Comments

OrgLang supports two types of comments: single-line and multiline (block) comments.

**Single-line comments** start with the hash character (`#`) and extend to the end of the line. They are ignored by the compiler.
```orglang
# This is a single-line comment
x : 42; # Comment after an expression
```

**Multiline comments** (also known as block comments) are enclosed in three consecutive hash characters (`###`). 
> [!IMPORTANT]
> The multiline comment marker `###` must start at the first column of the line.

Everything between the opening and closing `###` markers is treated as a comment and ignored.
```orglang
###
This is a multiline comment.
It can span multiple lines.
###
```

#### Blank lines

A line that contains only whitespace (spaces, tabs, and form feeds) is considered a blank line and is ignored by the compiler. Blank lines are recommended to separate logical blocks of code and improve readability.

#### Indentation

Unlike some other languages (like Python), indentation in OrgLang is generally not semantically significant. It is used primarily for readability and to reflect the structural hierarchy of nested blocks (e.g., inside `{}` or `[]`).

However, there are specific lexical rules where column position matters:
- The multiline comment marker `###` **must** start in the first column of the line.

#### Whitespace between tokens

Whitespace (spaces, tabs, and newlines) is used to separate tokens that would otherwise be joined. For example, `x : 42` requires whitespace around `:` if it were part of a larger word, but since `:` is a special symbol, it can often be used without spaces (e.g., `x:42`).

Certain symbols can be used without surrounding whitespace as they are recognized as distinct delimiters:
- `@`, `:`, `.`, `,`
- `(`, `)`, `[`, `]`, `{`, `}`

While not strictly required for these symbols, using whitespace is encouraged for visual clarity.

### Identifiers and keywords

#### Identifiers

Identifiers (also referred to as names) are used to name variables, functions, and resources. In OrgLang, identifiers have a very flexible structure, allowing many symbols that are typically reserved for operators in other languages.

An identifier must start with a letter (case-sensitive `a-z` or `A-Z`), an underscore (`_`), or any of the following symbols:
- `!`, `$`, `%`, `&`, `*`, `-`, `+`, `=`, `^`, `~`, `?`, `/`, `<`, `>`, `|`

After the first character, an identifier can contain any combination of letters, underscores, digits (`0-9`), and the symbols listed above.

However, identifiers **cannot** contain the following structural delimiters:
- `@`, `:`, `.`, `,`, `;`, `(`, `)`, `[`, `]`, `{`, `}`

**Examples of valid identifiers:**
- `variable_name`
- `isValid?`
- `++count`
- `>>`
- `my-module`
- `$price`

**Restricted Names:**
Identifiers that match any of the language's [Keywords](#keywords) are reserved and cannot be used as variable names.

> [!NOTE]
> Since digits are allowed in identifiers but an identifier cannot *start* with a digit, the lexer can easily distinguish between numeric literals and names. For example, `42x` is not a valid identifier, but `x42` is.

#### Keywords

The following identifiers are reserved as keywords and have special meaning in the OrgLang language. They cannot be used as ordinary identifiers:

- `true`: Boolean truth value.
- `false`: Boolean falsehood value.
- `resource`: Used in resource definitions.
- `this`: Refers to the current function or block (used for recursion).
- `left`: Predefined name for the left operand in a binary operator.
- `right`: Predefined name for the right operand (or the operand of a prefix operator).

#### Reserved classes of identifiers

### Literals

#### String literals

In OrgLang, strings are sequences of characters used for text representation. They can be defined as single-line or multiline literals.

##### Single-line Strings
Single-line strings are enclosed in double quotes (`"`). They must begin and end on the same line.

```orglang
message : "Hello, OrgLang!";
```

##### Multiline Strings (DocStrings)
Multiline strings are enclosed in triple double quotes (`"""`). They can span multiple lines and are designed for large blocks of text or documentation.

To keep the source code clean, multiline strings automatically strip **common leading whitespace** (indentation) from all non-empty lines. The amount of whitespace removed is determined by the line with the least indentation.

```orglang
# The resulting string will have no leading spaces on "Line 1" and "Line 2"
doc : """
    Line 1
    Line 2
""";
```

> [!NOTE]
> Leading and trailing blank lines (usually surrounding the delimiters) are also stripped.

##### Strings as Tables
A fundamental design choice in OrgLang is that **Strings are semantically Tables** (ordered lists of characters). 

- **Indexing**: A string can be indexed by integers starting at `0`. Accessing `s.0` returns the first character.
- **Table Properties**: Because they are tables, operators like `->` for iteration treat strings as a stream of characters.

##### String Length and Orthogonality
Strings behave like tables when used in arithmetic contexts:
- **Numeric Value**: When used with arithmetic operators (like `+` or `-`), a string evaluates to its **length**.
- **No Concatenation**: Unlike many languages, the `+` operator does **not** concatenate strings. Instead, it adds their lengths (or a string's length to a number).

```orglang
s1 : "ABC";
s2 : "DE";
res : s1 + s2; # res is 5 (3 + 2)
```
To join strings, use interpolation or specialized table operations (to be defined in the standard library).

#### Numeric literals

In OrgLang, all first-class numeric literals (integers and decimals) are designed to be implemented with **arbitrary precision**. This means that, by default, numbers are not limited by the standard 32-bit or 64-bit constraints of the underlying hardware, allowing for exact computations with very large or very precise values.

> [!NOTE]
> While the language semantics favor arbitrary precision, future versions of the compiler will introduce support for specific machine types (like `int`, `long`, `float`, `double`) as internal optimizations. These will be used when the compiler can prove that the range and precision requirements are satisfied by the more efficient machine representations.

#### Integer literals

Integer literals are represented as a sequence of one or more digits (`0-9`).

##### Signed Integer Literals
An integer literal can be preceded by an optional sign character (`+` or `-`). 

> [!IMPORTANT]
> To be treated as a single numeric literal, there **must not** be any whitespace between the sign and the digits. 

- `-42`: A single integer literal token with value negative 42.
- `- 42`: Two tokens: the unary negation operator `-` followed by the integer literal `42`.

While the external behavior might often be similar, the distinction is important for the binding power of operators and lexer-level identification of values.

#### Decimal literals

Decimal literals represent non-integer numbers using a fixed decimal point notation. In OrgLang, these are distinct from the "floating point" types found in many other languages because they are designed for arbitrary precision and avoid the precision loss typical of binary floating point representations.

##### Syntax
A decimal literal consists of an integer part and a fractional part separated by a dot (`.`). 

- **Digits on both sides**: For an unsigned number starting with a digit, there must be at least one digit on both sides of the dot (e.g., `3.14`).
- **Lexical Distinction**: A number followed by a dot without a subsequent digit (e.g., `1.`) is lexically interpreted as an [Integer literal](#integer-literals) followed by the [Dot operator](#operators).
- **No Scientific Notation**: OrgLang does not currently support scientific notation (e.g., `1e10`) in its literal syntax.

##### Signed Decimal Literals
Like integers, decimal literals can be preceded by an optional sign (`+` or `-`) with no intervening whitespace.

```orglang
pi : 3.14159;
negative_small : -0.0001;
positive : +1.0;
```

#### Rational literals

Rational literals represent exact fractional numbers using a syntax that clearly separates the numerator and denominator. This ensures precision in calculations involving fractions.

##### Syntax
A rational literal is formed by two integer literals separated by a forward slash (`/`).

- **Numerator and Denominator**: Both the numerator and the denominator are integer literals (which can be positive or negative, as defined in [Integer literals](#integer-literals)).
- **No Whitespace**: There must be no whitespace between the numerator, the slash, and the denominator.
- **Zero Denominator**: A zero in the denominator is syntactically valid but will result in a runtime error (division by zero) during evaluation.

##### Examples

```orglang
one_half : 1/2;
three_quarters : 3/4;
negative_fraction : -1/2;
large_fraction : 123456789/987654321;
```

#### Table literals

Tables (also referred to as Lists) are the primary data structure in OrgLang. There is a single, unified model for table construction, whether using blocks or operators.

##### Construction: Blocks and Commas
A table can be constructed in two ways that produce the same semantic object:

1.  **Blocks (`[]`)**: Square brackets group a sequence of statements or expressions into a Table. Elements are typically separated by whitespace.
2.  **The Comma Operator (`,`)**: The comma is a binary operator that creates or extends a Table. 

Because `[]` evaluates its contents and collects the results into a new Table, using commas inside brackets results in **nesting**.

```orglang
# A simple table
t1 : [1 2 3];

# The same table using commas (outside brackets)
t2 : 1, 2, 3;

# NESTED: [1, 2, 3] creates a Table containing a Table
t3 : [1, 2, 3]; # Result is [[1 2 3]]
```

##### Implicit Indexing vs. Bindings
A Table consists of a sequence of elements. These elements are categorized into two types:
- **Bindings**: Pairs created using the binding operator (`:`). These are accessed by their key and **do not** consume a numeric index slot.
- **Positional Elements**: All other expressions. These are assigned implicit numeric indexes (`0, 1, 2...`) based on their order of occurrence among other positional elements.

##### Mixed Content and Indexing
When a Table contains both bindings and positional elements, the numeric indexes skip the bindings.

```orglang
mixed : [10, "status" : "active", 20];

val0 : mixed.0;        # 10
status : mixed."status"; # "active"
val1 : mixed.1;        # 20 (NOT mixed.2)
```

##### Tables as Blocks
Since every Org file is itself a Table, the rules for table literals apply to the top-level structure of a program. A file containing `a:1; b:2;` is a Table where `a` and `b` can be accessed by name.

##### Laziness in Tables
Values within a table are **lazy by default**. They are represented as thunks and are only evaluated when accessed (e.g., using the `.` or `?` operators).

```orglang
computation : [1 + 1, 2 * 2]; # Expressions are not evaluated yet
result : computation.0;       # 2 (evaluation happens here)
```

### Operators

In OrgLang, almost every operation and structural construct is modeled as an **operator**. The language is designed to be highly orthogonal, with a minimal set of core rules that govern how these operators interact within expressions. Unlike many traditional languages that distinguish between operators, functions, and control structures, OrgLang treats nearly everything—from arithmetic to resource management and conditional evaluation—as an expression driven by operators.

#### Philosophy and Mechanics

OrgLang operators are strictly **unary** (prefix) or **binary** (infix). This strictness simplifies the language's grammar and execution model but introduces a different way of thinking about computation:

- **Binding Power**: The behavior of an expression is determined by the binding power (precedence) of its operators. Operators with higher binding power "pull" operands closer than those with lower power.
- **Everything is an Expression**: Operators don't just "perform actions"; they transform values and return new ones. This allows for deeply nested and highly expressive chains of computation.

#### Limitations and Patterns

The limitation to unary and binary forms (maximum of two operands) may seem restrictive compared to the variety of arities found in other languages. However, OrgLang overcomes this through several powerful patterns:

- **Tables as Parameters**: To pass multiple values to an operation that only accepts one operand (like a unary function call), those values are grouped into a [Table literal](#table-literals). The operation then extracts exactly what it needs from the table.
- **Currying**: Binary operators can be used to "partially apply" data. An expression like `a op b` can return a new thunk or function that is "ready" to take more data later.
- **Abstractions**: Simple operators can be composed and bound to names, creating high-level abstractions that behave like complex built-in features in other languages.

By embracing these patterns, OrgLang achieves a high degree of expressiveness while maintaining a structurally simple core.

#### Arithmetic operators

Arithmetic operators perform standard mathematical calculations. In OrgLang, these operators are designed to work with arbitrary-precision [Numeric literals](#numeric-literals).

| Operator | Name | Arity | Description |
| :--- | :--- | :--- | :--- |
| `+` | Addition | Binary | Returns the sum of two numbers. |
| `-` | Subtraction | Binary | Returns the difference between two numbers. |
| `-` | Negation | Unary | Returns the additive inverse of a number. |
| `*` | Multiplication | Binary | Returns the product of two numbers. |
| `/` | Division | Binary | Returns the quotient of two numbers. |
| `**` | Power | Binary | Returns the left operand raised to the power of the right operand (Right-associative). |

> [!NOTE]
> **Implicit Coercion**: Any arithmetic operator can be applied to [Table literals](#table-literals) and [Strings](#string-literals), in which case their **size** is used as the numeric value. Additionally, [Boolean literals](#boolean-literals) are coerced to numbers: `true` is treated as `1`, and `false` as `0`.

#### Bitwise operators

> [!NOTE]
> **TBD**: Pure bitwise operations (e.g., bit shifting `<<`, `>>`) are not yet implemented. The symbols `&`, `\|`, and `^` are currently available as **non-short-circuit [Boolean operators](#boolean-operators)**.

#### Comparison operators

Comparison operators compare two values and always return a [Boolean literal](#boolean-literals) (`true` or `false`). OrgLang supports standard comparison operations, as well as automatic [type coercion](#arithmetic-conversions) (e.g., comparing a string length to an integer).

| Operator | Description | Example |
| :--- | :--- | :--- |
| `=` | Equal to | `x = y` |
| `<>`, `~=` | Not equal to | `x <> y` |
| `<` | Less than | `x < y` |
| `<=` | Less than or equal to | `x <= y` |
| `>` | Greater than | `x > y` |
| `>=` | Greater than or equal to | `x >= y` |

> [!NOTE]
> **Implicit Coercion**: Comparison operators follow the same coercion rules as [Arithmetic operators](#arithmetic-operators): Tables and Strings use their size, and Booleans are treated as `1` (`true`) or `0` (`false`).

> [!IMPORTANT]
> **Comparison Chaining**: Since every comparison returns a Boolean, the result of a chain (e.g., `x < y < z`) is the result of the **last comparison** in the chain. This differs from languages where such a chain might be shorthand for `(x < y) && (y < z)`.

#### Boolean operators

Boolean operators are used to perform logical calculations.

| Operator | Name | Arity | Description |
| :--- | :--- | :--- | :--- |
| `~` | NOT | Unary | Returns the logical negation of a boolean value. |
| `&&` | AND | Binary | Short-circuit logical AND (returns `true` only if both are `true`). |
| `\|\|` | OR | Binary | Short-circuit logical OR (returns `true` if at least one is `true`). |
| `&` | Logical AND | Binary | **Non-short-circuit** logical AND. |
| `\|` | Logical OR | Binary | **Non-short-circuit** logical OR. |
| `^` | Logical XOR | Binary | Returns `true` if exactly one of the operands is `true`. |

> [!NOTE]
> **Truthiness**: Boolean operators can be applied to [Table literals](#table-literals) and [Strings](#string-literals). They follow a "size-based" truthiness rule: a size of `0` is treated as `false`, and every other value (size `> 0`) is treated as `true`.

#### Conditional operators

Conditional operators allow for selection and branching within expressions without traditional `if/else` statements.

| Operator | Name | Arity | Description |
| :--- | :--- | :--- | :--- |
| `.` | Dot Access | Binary | Static/Positional access to a Table's elements or keys. |
| `?` | Selection Access | Binary | Conditional or dynamic selection from a Table. Evaluation-driven. |
| `??` | Error Check | Binary | Returns the right operand if the left operand is an **Error**; otherwise, returns the left operand. |
| `?:` | Elvis Operator | Binary | Returns the right operand if the left operand is **"falsy"** (false, Error, or an empty Table/String); otherwise, returns the left operand. |

#### Resource operators

Resource operators manage the lifecycle and data flow of [Resources](#the-resources).

##### Resource Instantiation (`@`)
The prefix `@` operator is used to instantiate a resource. When applied to a resource name or literal, it executes the resource's `setup` block and returns a **Resource Instance**.

```orglang
# Instantiate stdout
@stdout
```

##### Data Flow (`->`)
The binary `->` operator drives data from a source (left operand) to a sink (right operand).

- **Source -> Sink**: Drives all data from the source into the sink until completion.
- **Iterator -> Function**: Creates a new projection (map) that will process elements lazily.

```orglang
# Send a string to stdout
"Hello" -> @stdout;

# Send input through a transform to output
@stdin -> { args * 2 } -> @stdout;
```

#### Assignment operators

In OrgLang, assignment is strictly an operation that binds a value to a name within a [Table](#tables-as-blocks).

| Operator | Name | Description |
| :--- | :--- | :--- |
| `:` | Binding | Binds the result of the right expression to the name specified on the left. |

#### Operator definitions

OrgLang allows for the definition of custom operators and the refinement of existing ones using the **Binding Power** syntax. This syntax defines the left and right binding powers, determining the operator's precedence and associativity.

```orglang
# Define a unary operator with prefix power 100
! : 100 { ... };

# Define a binary operator with left power 50 and right power 60
op : 50 { ... } 60;
```

When an operator is called, the expression within the braces is evaluated with `args` bound to the operand(s). For binary operators, `args` typically contains the right operand, while the left operand is made available via `this` or positional access depending on the context.

### Delimiters

## Data model

### Objects, values and types

### The standard type hierarchy

- Error
  - Expression
    - Name
    - Table
      - String
    - Number
      - Integer
      - Rational
      - Decimal
    - Boolean
    - Operator

### Special method names

## Execution model

### Naming and binding

### Errors

## Expressions

### Arithmetic conversions

### Atoms

#### Identifiers (Names)

#### Literals

#### Parenthesized forms

#### Displays for Tables, Lists, Sets and Dictionaries

#### Table displays

### Unary arithmetic and bitwise operations

### Binary arithmetic operations

### Shifting operations

### Binary bitwise operations

### Comparisons

### Boolean operations

### Conditional expressions

### Lambdas

### Expression lists

### Evaluation order

### Operator precedence

### Assignment 

### The resources

### The ? operator

### Operator definitions

### Module definitions

### File input

### Interactive input

### Expression input

### Full Grammar specification