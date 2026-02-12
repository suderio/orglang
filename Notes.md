# fun
## Definitions

### Structure

Every script starts with `{` and ends with `}`. Inside we may have:

### Comments

Anything from a `#` to the end of the line. It is completely ignored.

### Literals

Literals are convenient ways of representing concepts

#### Boolean

Just two, the usual stuff: `true`, `false`

#### Numbers

Can be Integers or Rational. A `.` means a Rational number, otherwise it is Integer.

* 100
* -1
* 1.0
* -0.01

#### UTF-8 Strings

Every String is assumed to be UTF-8. Encoding must be done in the input / output.

A String is made of Unicode codepoints and Expression placeholders.

* "A String"
* "" # Empty String
* "Hello ${name}" # A placeholder `name`

#### Labels

An identifier - what some languages would call variable name, only a bit more generic.

* a
* Xpto
* _

#### Arrays

Really dictionaries with Natural keys starting with `0`

* {} # Empty array
* {1, 2, 3}

#### Entry

A key:value pair.

* a: 1
* "a": 1
* 1: "a"
* true: a

#### Dictionaries

Really arrays with a key:value element. The key can be a label, a number, a boolean or a string.

* {a: 0, b: 1}
* {true: "Ok", false: "KO"}

A special element, `.` means "any element", to be used in filter expressions.

* {"this": 0, "that": 1, .: -1}

A reserved word, `this`, means the dictionary itself, allowing recursive definitions.

* {this} # a potential recursive bomb

A reserved word, `null`, means no value at all, and is equal to no value, including itself.

An expression in error will return `null` and, hopefully, an error message.

> Obs: Every array or dictionary has the same structure, an array of Entries.
> In an arrray the keys are the same as the index, in an dictionary they can be
> different. Therefore they are also sets of Entries (there cannot be two equal keys)

> Obs.: A dictionary may be made without some keys (ex.: `{a: 0, 1, 2}`). This
> will be the same as `{a: 0, 1: 1, 2: 2}`. This is why a valid file can be a
> list of valid expressions.

### Expressions

Arithmetic / boolean expressions, with the usual operators.

* 1 + 2 * 3 / 7 = 1 = true
* (1 - 1) * 2 > 0 = false
* 1 / 0 = null = false # There is no way of testing if value is null

Two arrays / dictionaries with no operator between them filters the left one with the keys of the right one.

* {a: 1, b: 2} {a} = {a: 1}
* {0, 1, 2} {\*} = {0, 1, 2}

A reserved word, `it` means the entire right array / dictionary that is being operated on the left.

* {a: it} {a, 1} = {a: {0, 1}}
* {0, it + 1} {0} = {0, {1}}
* {0, 1} {it} = {0 = it ? {it, null}, 1 = it ? {it, null}}

An expression with unbounded labels inside `[]` is a convenient way of building
a set that is used as a function.

* [a + 1] {1} = 2 # This is the same as {0: it + 1, a: it + 1}


## Examples

### Expressions

```
1 + 1
> 2
```

```
2 + (3 * 4)
> 14
```

### Arrays

```
{1, 2, 3}
```

### Dictionaries

```
{a: 1, 'b': 2, 3: 3, "four": 4, {five}: 5}
```

### Strings

```
"abc"

"""
abc
xpto
"""

"Hello $name" {name: "world!"};
```

### Labels
```
a: 1
```
```
b: {1, 2}
```
```
c: {a: 1, b: 2}
```
```
b {0}
> 1
```
```
c {a}
> 1
```

### Functions

#### Factorial using the any (`.`) element

f g -> returns every element of g from f

problem: arguments can be numbers / strings or just dicts?
```
fact: {
  0: 1,
  1: 1,
  .: this {0' - 1}
}
fact {3}
> 6
```

#### Sum of two numbers: `it n` is the idiom for the nth element of the argument array
```
sum: {
  .: it {0} + it {1}
}
sum {1, 2}
> 3
```

#### Expressions with unbounded labels define functions (surrounded by `[]`)
```
aSum: [a + b]
aSum {1, 2}
> 3

aSum {a: 1, b: 2}
> 3
```

#### Maps

A function (dict) with a `.` element will apply to every element of the filtered function
```
{2, 3} fact
> {2, 6}
{1, 2} sum
> null
{{1, 2}, {1, 3}} sum
> {2, 5}
```

#### Filters
```
{a: 1, b: 2, c: 3} {a, b}
> {a: 1, b: 2}

{a, b} {a: 1, b: 2, c:3}
> {}

{a: 1, b: 3} {a: 1, b: 2, c:3}
> {a: 1}
```

### Stdin, Stdout, Stderr
```
<< "Hello World!"
> "Hello World!"

<< "Hello ${name}" {name} >>
> World!
> Hello World!

<!< "Error!"
> "Error!"
```

### Everything together

```
a: {1, 2, 3}

a {}
> {1, 2, 3}

a {0}
> 1

a {0, 1}
> {1, 2}

a {-1}
> 3

a {1, -1}
> {2, 3}

a {-1, 1}
> {3, 2}

lenght a
> 3

lenght: {true: 1, false: 1 + this it {1, .}} {1' = null};
lenght: 1' = null ? { 1, 1 + this it {1, .}};

fib: {
    0: 0,
    1: 1,
    2: 1,
    .: this (0' - 1) + this (0' - 2)
}

lenght fib
> 4

x: {a: 1, b: 2}

x {}
> {a: 1, b: 2}

x {a}
> 1

x {a, c}
> {1, 3}

s: "abc"

s {}
> "abc"

s {0}
> "a"

s {0, 1}
> "ab"
```

### Syntatic Sugar

```
it {a}  -> a'

it {0} -> 0'

{true: "T", false: "F"} {a = 1} -> a = 1 ? {"T", "F"}
```

### Keywords
- this
- it
- null
- true
- false

### Operators

- arithm: + - * / % ( )
- logic: < > <= >= <> = & && | || ~ (not) ^ (xor)
- string: $ (string substitution)
- language: { } [ ] ' # (comment) : , ; << >> <\!< @ ?
- \ (latex strings)

### Imports
```
- someImport: fun >> @/someFile.fun
- anotherImport: >> fun @http://github.com/someRepo/someFile.fun
- iso-8891: (fun >> @strings.fun) iso-8891
```

### Files
```
- someString: iso-8891 >> @/someFile.txt
```

### Namespace

File based namespace: At each file you define (rename) the imports, and each valid
fun file is a Dictionary. By filtering the keys you can import only one element of
the file.


