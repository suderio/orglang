package parser

import "orglang/pkg/token"

// BindingPower defines the left and right binding power for an operator
type BindingPower struct {
	Left  int
	Right int
}

// OperatorPower holds the binding powers for prefix and infix usage
type OperatorPower struct {
	Prefix *BindingPower // nil if not prefix
	Infix  *BindingPower // nil if not infix
}

// OperatorBindingPowers returns the map of all basic operators and their binding powers
func OperatorBindingPowers() map[token.TokenType]OperatorPower {
	// Helper to make ptr
	bp := func(l, r int) *BindingPower { return &BindingPower{l, r} }

	ops := make(map[token.TokenType]OperatorPower)

	// Infix Operators
	ops[token.EQ] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.NOT_EQ] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.LT] = OperatorPower{Infix: bp(LESSGREATER, LESSGREATER)}
	ops[token.GT] = OperatorPower{Infix: bp(LESSGREATER, LESSGREATER)}
	ops[token.LT_EQ] = OperatorPower{Infix: bp(LESSGREATER, LESSGREATER)}
	ops[token.GT_EQ] = OperatorPower{Infix: bp(LESSGREATER, LESSGREATER)}
	ops[token.PLUS] = OperatorPower{Infix: bp(SUM, SUM)}
	ops[token.SLASH] = OperatorPower{Infix: bp(PRODUCT, PRODUCT)}
	ops[token.ASTERISK] = OperatorPower{Infix: bp(PRODUCT, PRODUCT)}
	ops[token.POWER] = OperatorPower{Infix: bp(PRODUCT+1, PRODUCT)}
	ops[token.LPAREN] = OperatorPower{Infix: bp(CALL, CALL), Prefix: bp(0, LOWEST)} // Grouped vs Call
	ops[token.DOT] = OperatorPower{Infix: bp(CALL, CALL)}
	ops[token.LBRACE] = OperatorPower{Infix: bp(CALL, CALL), Prefix: bp(0, LOWEST)} // Block
	ops[token.ARROW] = OperatorPower{Infix: bp(LESSGREATER, LESSGREATER)}           // -> Flow

	// Mixed Operators (Prefix and Infix)
	ops[token.MINUS] = OperatorPower{
		Prefix: bp(0, PREFIX), // -X
		Infix:  bp(SUM, SUM),  // X - Y
	}

	// Prefix Only
	ops[token.NOT] = OperatorPower{Prefix: bp(0, PREFIX)}
	ops[token.AT] = OperatorPower{Prefix: bp(0, PREFIX)}

	return ops
}
