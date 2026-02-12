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
	ops[token.LT] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.GT] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.LT_EQ] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.GT_EQ] = OperatorPower{Infix: bp(EQUALS, EQUALS)}
	ops[token.PLUS] = OperatorPower{Infix: bp(SUM, SUM)}
	ops[token.MINUS] = OperatorPower{
		Prefix: bp(0, PREFIX_LVL),
		Infix:  bp(SUM, SUM),
	}
	ops[token.SLASH] = OperatorPower{Infix: bp(PRODUCT, PRODUCT)}
	ops[token.ASTERISK] = OperatorPower{Infix: bp(PRODUCT, PRODUCT)}
	ops[token.POWER] = OperatorPower{Infix: bp(POWER_LVL+1, POWER_LVL)} // Right-associative
	ops[token.LPAREN] = OperatorPower{Infix: bp(CALL, CALL), Prefix: bp(0, LOWEST)}
	ops[token.DOT] = OperatorPower{Infix: bp(CALL, CALL+1)} // Strong left
	ops[token.QUESTION] = OperatorPower{Infix: bp(CALL, CALL+1)}
	ops[token.ELVIS] = OperatorPower{Infix: bp(CALL, CALL+1)}
	ops[token.ERROR_CHECK] = OperatorPower{Infix: bp(CALL, CALL+1)}
	ops[token.LBRACE] = OperatorPower{Infix: bp(CALL, CALL), Prefix: bp(0, LOWEST)}
	ops[token.ARROW] = OperatorPower{Infix: bp(FLOW, FLOW+1)}           // Left-associative
	ops[token.COMMA] = OperatorPower{Infix: bp(COMMA_LVL, COMMA_LVL+1)} // Left-associative
	ops[token.AT] = OperatorPower{
		Prefix: bp(0, PREFIX_LVL),
		Infix:  bp(CALL, CALL), // args @ sys
	}
	ops[token.COLON] = OperatorPower{Infix: bp(BINDING+1, BINDING)} // Right-associative (81, 80)
	ops[token.AND] = OperatorPower{Infix: bp(LOGICAL_AND, LOGICAL_AND)}
	ops[token.OR] = OperatorPower{Infix: bp(LOGICAL_OR, LOGICAL_OR)}
	ops[token.BIT_AND] = OperatorPower{Infix: bp(PRODUCT, PRODUCT)}
	ops[token.BIT_OR] = OperatorPower{Infix: bp(SUM, SUM)}
	ops[token.BIT_XOR] = OperatorPower{Infix: bp(SUM, SUM)}

	// Prefix Only
	ops[token.NOT] = OperatorPower{Prefix: bp(0, PREFIX_LVL)}

	return ops
}
