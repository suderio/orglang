# Operator Orthogonality Study

This document identifies OrgLang operators that have specific type restrictions on their operands, deviating from "Extreme Orthogonality" (where an operator would accept Number, Boolean, and Table/String equally).

## Type Context

As per the README, the standard type hierarchy is:

- **Number** (Integer, Rational, Decimal)
- **Boolean**
- **Table** (includes **String**)
- **Resource** (Accepted by all operators as source or sink)

*Note: The metadata types `Name` and `Operator` are ignored for this study.*

## Non-Orthogonal Operators

The following operators have restrictions on the types they accept for their operands.

| Operator | Left Operand | Right Operand | Description / Restrictions |
| :--- | :--- | :--- | :--- |
| `$` | String (Table) | Table | String interpolation. Rejects Number/Boolean on both sides. |
| `.` | Table | Number, String or Name | Access. Rejects Number/Boolean on Left. |
| `?` | *Any* | Table | Selection. Rejects Number/Boolean on Right. |
| `->` | *Any* | Resource / Operator / Table of Operators | Push. Rejects Number/Boolean/Table (unless only operators) on Right. |
| `-<` | *Any* | Resource / Operator / Table of Operators | Dispatch. Rejects Number/Boolean on Right. Rejects Number/Boolean/Table (unless only operators) on Right. |
| `-<>` | *Any* | Resource / Operator / Table of Operators | Join. Rejects Number/Boolean on Left. Rejects Number/Boolean/Table (unless only operators) on Right. |
| `@` | Table or Nothing (Unary) | Resource | Instantiation. Rejects Table/Boolean (unless as resource spec). |
| `@:` | Name | Table | Resource Definition. Rejects Number/Boolean on Right. |
| `o` | Operator | Operator | Operator Composition. Rejects anything but operators on both sides. |
| `\|>` | *Any* | Operator | Partial Application. Rejects anything but operators on the Right. |

## Orthogonal Operators (Omitted)

The following operators are considered fully orthogonal as they accept Number, Boolean, and Table/String (using coercion or size/truthiness rules):

- **Arithmetic**: `+`, `-`, `*`, `/`, `%`, `**`
- **Bitwise**: `&`, `|`, `^`, `~` (unary), `<<`, `>>`
- **Comparison**: `=`, `<>`, `~=`, `<`, `<=`, `>`, `>=`
- **Logical**: `&&`, `||`, `!` (unary)
- **Coalescing**: `??`, `?:`
- **Construction**: `,`
