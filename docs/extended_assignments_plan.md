# Extended Assignment Operators Implementation Plan

This plan outlines the steps to implement extended assignment operators (e.g., `:+`, `:-`) in OrgLang. These operators allow modifying existing bindings in a concise way (e.g., `x :+ 1` instead of `x : x + 1`).

## 1. Lexer Support

The lexer currently handles extended assignment operators via `readColon` (lexer.go:542) by emitting them as `IDENTIFIER` tokens with the full literal (e.g., `:+`).

### Current Status

- Supported in `readColon`: `:+`, `:-`, `: *`, `:/`, `:%`, `: &`, `:^`, `:|`, `:~`, `:>>`, `:<<`, `:>`, `:<`.
- **Action**: Add `:**` (Power assignment) to `readColon`.
- **Verification**: `readColon` currently does NOT check for `*` after `*`.

## 2. Token definitions

No new token types are strictly needed since they are lexed as `IDENTIFIER`.
However, we might want to consider if `BINDING_OP` is a better category than `IDENTIFIER` for parser logic, but `IDENTIFIER` works if registered in binding table.

## 3. Parser Support

### binding_powers.go

Register the extended assignment operators in the Binding Table.

- **Precedence**: They should share the same precedence as standard binding (`:`), which is 80.
- **Associativity**: Right-associative (Standard for assignment).
- **Operators to Register**:
  - `:+`, `:-`, `:*`, `:/`, `:%`
  - `: &`, `:|`, `:^`, `:~`
  - `:<<`, `:>>`
  - (Note: `++` and `--` treated as prefix operators, potentially deprecated).

### AST Changes (parser.go / ast.go)

Standard `InfixExpr` is unsuitable because it evaluates the Left side to a value. Extended assignment requires the **Left side to be an L-Value** (Name or specialized access).

**Plan A: Extend `BindingExpr`**
Modify `ast.BindingExpr` to include the operator.

```go
type BindingExpr struct {
    Name     *Name
    Operator string // ":" by default, or ":+", ":-", etc.
    Value    Expression
    IsResource bool
}
```

**Plan B: `AssignmentExpr`**
Separate node if logic diverges significantly. But `BindingExpr` is basically assignment in OrgLang.

**Preferred**: Extend `BindingExpr`.

### Parsing Logic

Update `led` handling in `parser.go`.

- Currently `case token.COLON:` calls `ledBinding`.
- We need to handle `token.IDENTIFIER` when the literal is an assignment operator.
- In `led`:

  ```go
  case token.IDENTIFIER:
      if isAssignment(t.Literal) {
          return p.ledBinding(left, t.Literal) // Pass op
      }
      // ... existing infix handling ...
  ```

- Update `ledBinding` to accept the operator string and validate logical L-Value constraints (though `BindingExpr` requires `*ast.Name` currently).

## 4. Runtime Support (Future Phase)

This plan focuses on Lexer/Parser, but note for Runtime:

- `Evaluator` for `BindingExpr` needs to handle operators.
- `scope.Set` vs `scope.Update`. `:+` implies logical "Get, Op, Set".
- `Scope` needs to support this.

## 5. Test Cases

Add tests in `parser_tests.go` or `parser_examples_test.go`.

```rust
// Basic
count :+ 1;
// Compound
x :* 2;
// Mask
flags :| 0x01;
```

Expected AST:

```lisp
(count :+ 1)
(x :* 2)
(flags :| 0x01)
```

## 6. Documentation

Update `TODO.md` or `reference.md` to formally list these operators.
