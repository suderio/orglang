# OrgLang EBNF Grammar (Draft)

This grammar describes the syntactic structure of OrgLang.
It relies on a Pratt Parser for expression precedence, so the EBNF here describes the *valid sequences of tokens*, not the evaluation tree.

## Lexical Tokens

```ebnf
/* Whitespace is significant as a separator but ignored for parsing structure */
WHITESPACE  ::= [ \t\n\r]+

/* Comments */
COMMENT     ::= "#" [^\n]* 
              | "###" (.*?) "###"

/* Literals */
INTEGER     ::= [0-9]+
DECIMAL     ::= [0-9]+ "." [0-9]+
STRING      ::= '"' ([^"] | '\"')* '"' 
              | '"""' (.*?) '"""'
BOOLEAN     ::= "true" | "false"
ERROR       ::= "Error"

/* Identifiers */
/* Can contain a-z, A-Z, 0-9, _, and allowed symbols */
/* Excludes structural: @ ; : . [ ] ( ) { } \ " ' , */
/* Allowed Symbols: ! $ % & * - + = ^ ~ ? / < > | ` */
/* Note: Identifiers cannot start with a digit unless it is a number literal */
/* Unary operators like - and ~ must be separated by space if used as operators */
IDENTIFIER  ::= [a-zA-Z_!$%&*+\-=\^~?/<|>][a-zA-Z0-9_!$%&*+\-=\^~?/<|>]*

/* Structural Tokens */
/* These tokens break identifiers and do not require whitespace */
LPAREN      ::= "("
RPAREN      ::= ")"
LBRACKET    ::= "["
RBRACKET    ::= "]"
LBRACE      ::= "{"
RBRACE      ::= "}"
SEMICOLON   ::= ";"
COLON       ::= ":"
DOT         ::= "."
COMMA       ::= ","
AT          ::= "@"
BINDING_TAG ::= "!" [0-9]+

/* Literals */
INTEGER     ::= [0-9]+
DECIMAL     ::= [0-9]+ "." [0-9]+
STRING      ::= '"' ([^"] | '\"')* '"' 
              | '"""' (.*?) '"""'
BOOLEAN     ::= "true" | "false"
ERROR       ::= "Error"

```

## Syntactic Grammar

```ebnf
Program ::= Statement*

Statement ::= Expression ";"

/* 
   In a Pratt parser, "Expression" consumes tokens based on binding power.
   Lexically, an Expression is a sequence of Terms and Operators.
*/

Expression ::= Term (Operator Term)*  /* Conceptual Simplification */

/* A Term is the atomic unit that starts an expression (NUD position) */
Term ::= Literal
       | Identifier
       | Group
       | List
       | Block
       | PrefixExpression

Group ::= "(" Expression ")"

/* Lists (Tables) contain a sequence of expressions separated by commas */
/* In OrgLang, the comma is an operator, so [1, 2] is parsed as list op list */
/* Space separated values like [1 2] are parsed as Function Application (Juxtaposition) */
/* Wait, [1 2] means list with 1 and 2. Juxtaposition inside [] is usually list construction. */
/* Correction: Inside [], juxtaExpression ::= Term (Operator Term)*
Term ::= Literal | Identifier | Group | List | Block | PrefixExpression
List ::= "[" (Expression)* "]"
/* Binding power syntax: 700{...}701. Spaces not allowed between integer and brace. */
Block ::= (INTEGER)? "{" Expression "}" (INTEGER)?

/* Operations */

/* Prefix Operation (NUD): op right */
/* @identifier is parsed here */
PrefixExpression ::= Operator Term

/* Infix Operation (LED): left op right */
/* "path" @ file is parsed here */
/* List construction with comma is parsed here */
/* Term Operator Term */

Operator ::= IDENTIFIER | AT | DOT | COLON | COMMA | ELVIS

/* Special Token Definitions */
ELVIS ::= "?:"


```

## Key Parsing Rules (Pratt)

1.  **Juxtaposition**: Two `Term`s appearing consecutively (e.g., `increment 5` or `list.0`).
    *   **Logic**: Does the grammar treat `Space` as an implicit infix operator?
    No. Space is just a separator. The parser checks if the *next* token has a Left Binding Power (Infix).
        *   If `increment` (NUD) is followed by `5` (NUD), and `5` has no LBP, this is a **Juxtaposition** (Function Call).
        *   If `1` (NUD) is followed by `+` (LBP > 0), this is an **Infix Operation**.

2.  **The `@` Operator**:
    *   Defined as Unary Prefix (`@identifier`).
    *   Usage: `@stdout` (Prefix).
    *   Usage: `"path" @ file` (Infix context).
    *   The `@` operator is special. It is a prefix operator when used as `@identifier` (or `@ identifier`), but it is an infix operator when used as `"path" @ file`.

3.  **Binding Power**:
    *   The `!N{...}N!` syntax is handled as a `Block` with optional metadata.
    Yes, it is a block. The binding tags are just metadata.

## Resolved Ambiguities

1.  **Comma Separator**: `,` is a table-construction operator. It merges elements or tables.
2.  **Comparison Chaining**: `a < b < c` parses as `(a < b) < c`. Since `true=1`, `1 < c` is valid.
3.  **Assignments**: `x : 1` is a table insertion expression that returns the assigned value.

