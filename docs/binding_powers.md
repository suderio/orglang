# Operator Binding Powers

This document lists the binding powers for all basic operators in OrgLang, as implemented in the parser. The parser uses a Pratt Parser approach where each operator has a Left and/or Right binding power.

## Precedence Levels

| Level | Value | Description |
| :--- | :--- | :--- |
| `LOWEST` | 1 | Base precedence |
| `EQUALS` | 2 | `==`, `<>` |
| `LESSGREATER` | 3 | `<`, `>`, `<=`, `>=` |
| `SUM` | 4 | `+`, `-` |
| `PRODUCT` | 5 | `*`, `/`, `**` |
| `PREFIX` | 6 | `-X`, `!X`, `~X`, `@X` |
| `CALL` | 7 | `func(x)`, `table.key`, `block { ... }` |
| `INDEX` | 8 | `array[i]` |

## Operator Map

### Infix Operators (Left Associative)

| Operator | Left BP | Right BP | Precedence |
| :--- | :--- | :--- | :--- |
| `==`, `<>` | `EQUALS` (2) | `EQUALS` (2) | Equality |
| `<`, `>`, `<=`, `>=` | `LESSGREATER` (3) | `LESSGREATER` (3) | Comparison |
| `+`, `-` | `SUM` (4) | `SUM` (4) | Arithmetic |
| `*`, `/` | `PRODUCT` (5) | `PRODUCT` (5) | Arithmetic |
| `(` (Call) | `CALL` (7) | `CALL` (7) | Function Call |
| `.` (Access) | `CALL` (7) | `CALL` (7) | Member Access |
| `{` (Block) | `CALL` (7) | `CALL` (7) | Block Definition |

### Infix Operators (Right Associative)

| Operator | Left BP | Right BP | Precedence |
| :--- | :--- | :--- | :--- |
| `**` | `PRODUCT` + 1 (6) | `PRODUCT` (5) | Power |

### Prefix Operators (Right Associative)

| Operator | Right BP | Precedence |
| :--- | :--- | :--- |
| `-` (Negation) | `PREFIX` (6) | Unary Minus |
| `~` (Not) | `PREFIX` (6) | Logical Not |
| `@` (Sys) | `PREFIX` (6) | System Resource |
| `{` (Block Start) | `LOWEST` (1) | Block Literal |
| `(` (Group Start) | `LOWEST` (1) | Grouped Expr |

### Mixed Operators (Prefix & Infix)

*   **`-`**:
    *   **Prefix**: Unary Negation (`-1`) -> BP 6
    *   **Infix**: Subtraction (`1 - 2`) -> BP 4
*   **`(`**:
    *   **Prefix**: Grouping `(1 + 2)` -> BP 1
    *   **Infix**: Call `func(1, 2)` -> BP 7
*   **`{`**:
    *   **Prefix**: Block `{ ... }` -> BP 1
    *   **Infix**: Binding Power `700{ ... }` -> BP 7
