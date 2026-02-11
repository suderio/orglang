---
author: Paulo Suderio
date: 28 janeiro 2026
title: "Especificação de Design: OrgLang"
---

# Language definition

Welcome to the OrgLang reference manual! This guide provides a
comprehensive and user-friendly overview of the language, covering
syntax, operators, data types, and constructs. OrgLang is designed to be
expressive, flexible, and easy to use, supporting arithmetic operations,
logic, table manipulation, and more.

## Initial Design Goals

-   Scripting language, used to fast development
-   Very few types:
    -   Error
    -   Boolean
    -   Number (Integer, Rational and Decimal subtypes)
    -   Table (String subtype)
    -   Operator
    -   Resource
-   Variables can hold any type (dynamic type, only values have type,
    not variables)
-   Error is the basic value ; in (almost) any operation with error,
    error is returned
-   Resources are, e.g., files, stdin, stdout, etc.
-   Every Operator (i.e., Function) must be able to return a value for
    any type, or Error if the value is not supported.
-   Extreme orthogonality of the basic operators. For instance,
    `[0] + [0 0] = 3` (uses the size of the Table). `"Hello" - 1 = 4`.
-   Spaces are used to separate tokens. Some Symbols can be used
    without spaces, such as `@`, `:` (and other assignments), `.`,
    `(` , `)` , `[` , `]` , `{` , `}` and `,`. Therefore, they cannot be used
    in identifiers. Identifiers can contain letters, numbers, and underscores
    and any other symbol.

## Basic Structure

### File

A OrgLang program is a sequence of expressions, separated by semicolons
(`;`). Example:


```orglang
x : 42;
y : x + 8;
result : y * 2;
```

### Modules and Imports

Every source file returns a Table (more about it below) with the result
of every expression and all assigned variables.

There is a built-in resource `@org` that imports source files from the
same directory as the current file (or some directory under it). It is
used as a module system. The table returned by the source file is
assigned to the current scope.

```orglang
myModule : "file.org" @ org;
```

### Assignments

A variable is assigned using `:`. The left-hand side is the identifier,
and the right-hand side is an expression. An assignment is also an
expression that returns its right-hand side value.

``` orglang
name : "John Doe";
age : 30;
```

Obs.: Since the source code creates a Table, assignments are actually table insertions
whith key: value pairs. The key is the identifier and the value is the expression.

We could have a source file with

```orglang
[
    x : 42
    y : 84
]
```

And that would be the same as the syntatic sugar

```orglang
x : 42;
y : 84;
```

It is perfectly valid to assign a value to a key that already exists, in which case the value
is updated.

### Scope and Lifecycle

Since the source code creates a Table, the lifecycle of variables is tied to the lifecycle of the Table. Also, file-scope is the same as block-scope.

Variables are shadowed when a variable with the same name is assigned in a nested scope. Every new operator is created in a new scope, and every assignment is an insertion in the current scope.

### Resources

Resources are created using the `@` operator. They are closed when the Table is closed.

This is a debatable design choice. It is not clear if it is better to have a deterministic destructor mechanism or not.

## Supported Literals

1.  Integers:

``` orglang
x : 42;
```

1.  Decimals:

``` orglang
pi : 3.14;
```

1.  Rationals:

``` orglang
two_thirds : 2/3;
```

1.  Strings:

2.  Simple: Enclosed in double quotes (`"`).

3.  Multiline (DocStrings): Enclosed in `"""`.

``` orglang
simpleString : "Hello, World!";
docString : """
    Line 1
    Line 2
""";
```

The String literal creates a Table indexed by integers starting at 0, with every character as a value.

1.  Booleans:

``` orglang
flag : true;
```

1.  Error:

A value that has the distinct property of not being assignable. Everytime an error is returned, it is expected that some message is printed to stderr.

As Error is a value, it can be returned by operators. For instance, if an operator is called with an argument of the wrong type, it will return Error.

Most operators will return Error if one of the arguments is Error. The exceptions are the operators that are designed to handle errors, such as `??` and `?:`. They must be used with caution.

``` orglang
value : Error; # Actually, this is not valid code, since Error cannot be assigned.
```

1.  Special Identifiers:

2.  `this`: Refers to the whole expression of an operator. Used in
    recursion.

3.  `right`: The right argument, or the argument of prefix operators.

4.  `left`: The left argument, or the argument of postfix operators.

## Operators

There are few primitive operators. Most operators are just a \'default
library\'.

### Primitive

  Operator        Description           Example                        
  --------------- -------------------------- ---------------------------
  `.`             Table filter              `table.key`                    
  `@`             Unit                      `"Hello World" -> @stdout`      
  `?`             Test (truthy?)            `(1 = 0) ?`                    
  `??`            Error Test                `x ?? 42`                      
  `?:`            Elvis (falsey?)           `x ?: 42`                      
  `->`            Broadcast Map             `[1 2 3] -> sum`               
  `-<`            Balanced Map              `[1 2 3] -< [sum1 sum2]`       
  `-<>`           Join Map                  `[1 2 3] -<> sum`              
  `o`             Compose                   `g o f`                        
  `|>`            Partial application       `add1: 1 |> +`
  `N{, }N`        Left/Right Binding Power  `501{right - left}502`

### Precedence Table

The precedence table is defined in the Pratt Parser.

In a Pratt Parser, precedence is determined by the **Binding Power** of an operator. A higher number "pulls" the expression more strongly. If an operator is **Left-Associative**, its Right Binding Power (RBP) is slightly higher than its Left Binding Power (LBP), forcing the parser to consume the next token as part of the current expression.

Here is the definitive binding power table for **OrgLang**.

### OrgLang Precedence Table

| Operator | Description | LBP (Left) | RBP (Right) | Associativity |
| --- | --- | --- | --- | --- |
| `;` | Statement Terminator | 0 | 0 | N/A |
| `:` | Binding (Assignment) | 10 | 9 | **Right** |
| `,` | List/Pair Separator | 20 | 21 | Left |
| `\|>` | Left Injector (Anchor) | 40 | 41 |
| `->`, `-<`, `-<>` | Flow / Dispatch / Join | 50 | 51 | Left |
| **(Default)** | **User Defined Operators** | **100** | **101** | **Left** |
| `==`, `~=`, `<`, `>` | Comparisons | 150 | 151 | Left |
| `+`, `-` | Addition / Subtraction | 200 | 201 | Left |
| `*`, `/` | Mult / Division | 300 | 301 | Left |
| `o` | Composition | 400 | 401 | Left |
| `^` | Exponentiation | 501 | 500 | **Right** |
| `.` | Evaluative Lookup | 800 | 801 | Left |
| `~`, `@`, `-` | Unary (Prefix) | 0 | 900 | N/A (High) |

### Key Design Decisions:

1. **The Default Zone (100):**
By setting the default for user-defined operators to **100**, we ensure that custom operators are stronger than the structural flow (`->`, `|>`) but weaker than standard arithmetic (`+`, `*`). This prevents a custom operator from "stealing" an operand that should mathematically belong to a sum or product.
2. **Binding (`:`) is Right-Associative:**
This allows for chained bindings like `a : b : 10;`, which is parsed as `a : (b : 10);`. This is useful for aliasing or multiple names for the same thunk.
3. **Flow (`->`) vs. Injector (`|>`):**
`|>` has a slightly lower precedence than `->`. This ensures that in an expression like `data |> filter -> output`, the `data` is first injected into `filter` before the resulting function is used to start the flow.
4. **Lookup (`.`) is the Strongest Infix:**
The evaluative lookup must bind tighter than anything else. `user.name + "!"` must be parsed as `(user.name) + "!"`, not `user.(name + "!")`.
5. **Unary Operators:**
Prefix operators have an LBP of `0` because they do not bind to anything on their left. Their RBP is extremely high (`900`) to ensure they capture the immediate expression following them.

### Customizing Power

As we discussed, a user can override these defaults using the slot notation within a function definition:

```orglang
// Defining a custom operator with high precedence
# : 700{ left ** right }700;

```

This allows the user to place their custom "Molecular" operators anywhere in the hierarchy.

### Arithmetic

Arithmetic operators always return a number (integer or decimal).

  ---------- ---------------- ----------
  Operator   Description      Example
  --------   --------------   --------
  `+`        Addition         `3 + 2`
  `-`        Subtraction      `5 - 1`
  `*`        Multiplication   `4 * 2`
  `/`        Division         `8 / 4`
  `%`        Modulo           `5 % 2`
  `**`        Exponentiation  `2 ** 3`
  ---------- ---------------- ----------

### Extended Assignment

  Operator    Description                     Example
  ----------- ------------------------------- -----------
  `:+`        Addition and Assignment         `x :+ 2`
  `:-`        Subtraction and Assignment      `x :- 1`
  `:*`        Multiplication and Assignment   `x :* 3`
  `:/`        Division and Assignment         `x :/ 4`
  `:%`        Modulo and Assignment           `x :% 5`
  `--`        Increment and Assignment        `--x`
  `++`        Decrement and Assignment        `++x`
  `:>>`       Right Shift and Assignment      `x :>> 5`
  `:<<`       Left Shift and Assignment       `x :<< 5`
  `:&`        AND and Assignment              `x :& y`
  `:^`        XOR and Assignment              `x :^ y`
  `:|~`       OR and Assignment               `x :\ | y`

### Logical

Logical operators always return a Boolean.

  Operator   Description         Example
  ---------- ------------------- -----------------
  `&&`       Short-circuit AND   `true && false`
  `||`       Short-circuit OR    `true || false`
  `&`        Bitwise AND         `x & y`
  `|`        Bitwise OR          `x | y`
  `^`        Bitwise XOR         `x ^ y`
  `~`        NOT                 `~x`

### Comparison

Comparison operators always return a Boolean.

  Operator     Description                Example
  ------------ -------------------------- ----------
  `=`          Equal to                   `x = y`
  `<>`, `~=`   Not equal to               `x <> y`
  `<`          Less than                  `x < y`
  `<=`         Less than or equal to      `x <= y`
  `>`          Greater than               `x > y`
  `>=`         Greater than or equal to   `x >= y`

Care must be taken when doing comparison chaining. Since every comparison operator returns a Boolean, the result of a comparison chain is the result of the last comparison.

### Miscellaneous

  Operator   Description           Example
  ---------- --------------------- ---------------------
  `$`        String substitution   `"Hello $0" $ [42]`
  `..`       Numeric range         `1..5`

### Tables

1.  Construction

    Tables are lists of elements or key-value pairs, constructed with
    square brackets (`[]`):

    ``` orglang
    table1 : [1 2 3];
    table2 : ["key": "value" "anotherKey": 42];
    ```

2.  Concatenation

    Commas (`,`) can be used to concatenate elements into tables:

    ``` orglang
    result : [1 2], [3 4];
    # [1 2 [3 4]]
    ```

    If the left hand is a list, ',' adds the right hand argument. If
    it is not a list, it creates a new list with the two elements.

3.  Accessing Elements

    Elements can be accessed using `.`:

    ``` orglang
    table : [1 2 3];
    result : table.0;
    # result: 1
    ```

Trying to access a key that does not exist returns `Error`.

``` orglang
table : [1 2 3];
result : table.4;
# result: Error
```

4. Laziness of values

Values on tables are not evaluated until they are accessed. This is
different from functions, which are evaluated when they are called.

A table is just a collection of thunks.

``` orglang
x: [1 2 (1 + 3)] # 1 + 3 is not evaluated
x.2 # 4
# The `.` operator selects the element and evaluates it.

### Mutability

Tables are mutable. Elements can be modified using the `.` operator:

``` orglang
result : table.0 : 5;
# result: 5
```

### Functions and Operators

1.  Defining Unary Operators

    Custom operators can be defined using '{}':

    ``` orglang
    increment : {right + 1};
    ```

    Invocation

    ``` orglang
    result : increment 5;
    # result: 6
    ```

    Obs.: the `left` argument is `Error`.


2.  Defining Binary Operators

    ``` orglang
    reverse_minus : {right - left};
    ```

    Invocation

    ``` orglang
    result : 2 reverse_minus 6;

3.  Recursion

    Every operator defines a `this` operator that references itself.

    ``` orglang
    factorial : {[1 1].right ?? right * (this(right - 1))};
    ```

    Every user defined operator has the same precedence and associativity (right-associative).

    To change it, use the Right/Left Binding Power operators.

### Strings and Substitutions

1.  Positional Substitution

    Use `$N` to substitute values from a list:

    ``` orglang
    message : "Value: $0, Double: $1" $ [42, 84];
    ```

2.  Variable Substitution

    Use `$var` to substitute the value of a declared variable:

    ``` orglang
    x : 10;
    message : "The value is $x";
    ```

### Comments

1.  Single-Line

    Start with `#`:

    ``` orglang
    # This is a single-line comment
    ```

2.  Block

    Enclosed in `###`:

    ``` orglang
    ###
    This is a
    block comment.
    ###
    ```

## Resource and Flux

### Resource Primitives

  ----------- ----------------- ------------------------------------------------------------
  Primitive   Abstraction       Role in the Runtime
  ----------- ----------------- ------------------------------------------------------------
  \@handle    I/O Channel       Unique interface for Files and Sockets.
  \@mem       Space Channel     Addressable memory treated as a seekable stream.
  \@signal    Event Channel     Temporal triggers (`@clock`) or system (`@metadata`).
  \@sys       Event Channel     Direct syscall invocation (ex: read, write, open).
  ----------- ----------------- ------------------------------------------------------------

### `@` (The Resource Atom)

The `@` operator is the Lifter. It takes a symbol or a string
and "prompts" the runtime to associate it with a system effect or a
resource handle.

1.  `@stdout`: A sink for the standard output.

2.  `@file`: A constructor that turns a path into a streamable handle.

3.  Visual logic: Think of `@` as the boundary. Anything with a
    `@` touches the "outside world" (the OS).

### `->` (The Pulse / Flow)

The `->` operator is the Stream Binder. It moves data from a
source to a destination.

1.  `source -> sink`: Every time `source` produces a "pulse" of data,
    it is pushed into `sink`.

2.  `"Hello World" -> @stdout` treats the string as a single-pulse
    stream and sends it to the console.

When the left side is a table, every one of its elements is sent to the sink. To send a table as a single pulse, wrap it in a list:

``` orglang
[1 2 3] -> @stdout; # sends 1, 2, 3
[[1 2 3]] -> @stdout; # sends [1 2 3]
```

When the right side is a table, each one of the left side elements is sent to every sink in the right side table.

``` orglang
[1 2 3] -> [[@stdout] ["output.txt" @ file ]]; # sends 1, 2, 3 to both
```

### `|>` (The Anchor / Left Injector)

The `|>` operator is the Partial Applier. It takes a value and
"anchors" it into the `left` slot of the following expression,
returning a new function.

1.  `("Hello " |> concat)`: This creates a new unary function that
    already has `"Hello "` as its `left` value. It is now waiting for a
    `right` value to complete the concatenation.

### `o` (The Composer)

The `o` operator is the Kleisli Composer. it merges two
functions or resources into a single pipeline before any data flows
through them.

1.  `h : g o f`: Creates a new transformation where the output of `f`
    becomes the input of `g`.

### `-<>` (The Barrier / Join)

The `-<>` operator is the Synchronizer. It is used when a flow
has been split (forked). It tells the runtime: "Do not proceed to the
next node until all previous parallel branches have completed their
current frame."

A flows into B and C in parallel. 
-<> ensures D only starts after B and C finish. A -> [ B C ] -<> D;

## Laziness

The transition from a static "if-table" to a **Lazy Mapping Operation** (`->`) is where OrgLang moves from simple logic to a high-performance streaming engine.

When you apply a mapping operation to a stream of data using lazy evaluation, you shift the execution model from **Batch Processing** (calculate everything now) to **Demand-Driven Processing** (calculate only when the next node pulls).

Here are the specific consequences for the performance and execution model:

---

### 1. Execution Model: The "Pull" Pipeline

In an eager model, a map operation would transform the entire list before moving to the next node. In the OrgLang lazy model, the mapping operation creates a **Chain of Thunks**.

* **How it works:** When `@stdout` asks for data, it sends a "pull" signal up the chain. The Map node receives this, pulls one element from the source, applies the function `{ }`, and passes the result down.
* **The Benefit:** You achieve **Constant Memory Usage**. Whether you are processing 10 items or 10 billion, the memory footprint remains the same because only one "frame" of the map is active at a time.

---

### 2. Performance: Thunk Fusing (Zero-Copy)

One of the biggest performance wins in lazy mapping is **Operator Fusing**. If you have multiple maps in a row:
`data -> { right + 1 } -> { right * 2 } -> @stdout`

* **Eager Model:** Would create an intermediate list after the first map, then another after the second.
* **OrgLang Lazy Model:** The Go compiler "fuses" these into a single C function. Instead of passing data between functions, it generates a single loop that performs `(x + 1) * 2`.
* **Result:** You eliminate the overhead of function calls and intermediate memory writes, reaching **raw C loop speeds**.

---

### 3. Consequence: Latency vs. Throughput

Laziness changes the "rhythm" of your program:

* **Lower Initial Latency:** The first item reaches the `@stdout` almost instantly. You don't have to wait for the entire map to finish.
* **Overhead per Item:** There is a tiny "bookkeeping" cost to manage the thunk state for each item. In high-throughput scenarios (millions of small integers), this overhead might be visible compared to a hand-optimized C array loop.

---

### 4. Memory Management: Arena Lifetimes

In a lazy map, the **Arena** becomes a moving window.

* **The Risk:** If a mapping function allocates memory (e.g., creating a string) and that string is passed down the pipeline, the memory must stay alive until the final sink (`@stdout`) is done with it.
* **The Teardown:** OrgLang handles this by linking the "Frame Life" to the Arena. Once the `@stdout` finishes processing a frame, the "Frame Arena" is reset. This prevents memory from ballooning during long-running streams.

---

### 5. Interaction with Parallelism (`-> [ B, C ]`)

This is where laziness gets complex. If you fork a stream to two parallel nodes:

1. The "Pull" signal comes from two different places.
2. The Lazy Map must now act as a **Buffer**. It pulls once from the source and "holds" the value until both B and C have consumed it.
3. **Performance Cost:** This introduces a synchronization primitive (a mutex or a semaphore) in the C runtime to ensure the lazy value is not evaluated twice or freed too early.

---

### Summary Table: Eager vs. Lazy Mapping

| Metric | Eager Mapping | OrgLang Lazy Mapping |
| --- | --- | --- |
| **Memory Usage** |  (proportional to data size) | ** (constant per frame)** |
| **Start-up Time** | Slow (waits for full map) | **Instant (pull-based)** |
| **CPU Cache** | Good (linear array access) | **Excellent (data stays in L1 cache between steps)** |
| **Composition** | Expensive (creates temp lists) | **Free (thunk fusing)** |

### Final Conclusion

For a systems language like **OrgLang**, laziness in mapping is a net win. It allows you to model infinite streams (like `@stdin` or a network socket) that would be impossible to handle eagerly.

The "cost" of managing thunks is heavily offset by **Cache Locality**: because the data is processed through the entire pipeline one piece at a time, it often stays in the CPU's L1/L2 cache, avoiding the massive performance penalty of hitting Main Memory (RAM).

