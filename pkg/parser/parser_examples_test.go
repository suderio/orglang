package parser

import (
	"testing"

	"orglang/pkg/lexer"
)

func TestParser_Examples(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 01_basics.org
		{
			name:     "Basics - Arithmetic",
			input:    "a:10; b:20; sum: a + b; diff: b - a; prod: a * b; quot: b / a; mod: b % 3; pow: 2 ** 3;",
			expected: "(a : 10)\n(b : 20)\n(sum : (a + b))\n(diff : (b - a))\n(prod : (a * b))\n(quot : (b / a))\n(mod : (b % 3))\n(pow : (2 ** 3))",
		},
		{
			name:     "Basics - Decimals and Rationals",
			input:    "pi: 3.14; r: 2.0; area: pi * r ** 2; half: 1/2; quarter: 1/4; three_quarters: half + quarter;",
			expected: "(pi : 3.14)\n(r : 2.0)\n(area : (pi * (r ** 2)))\n(half : 1/2)\n(quarter : 1/4)\n(three_quarters : (half + quarter))",
		},
		{
			name: "Basics - Strings and Booleans",
			// Added decls for a, b
			input:    "a:10; b:20; msg: \"Hello\"; is_eq: a = 10; is_gt: b > a; both: is_eq && is_gt;",
			expected: "(a : 10)\n(b : 20)\n(msg : \"Hello\")\n(is_eq : (a = 10))\n(is_gt : (b > a))\n(both : (is_eq && is_gt))",
		},

		// 02_tables.org
		{
			name:     "Tables - Basic List",
			input:    "list : [1 2 3 4 5];",
			expected: "(list : [1 2 3 4 5])",
		},
		{
			name:     "Tables - Access",
			input:    "list: [1]; first : list.0;",
			expected: "(list : [1])\n(first : (list.0))",
		},
		{
			name:  "Tables - Key Value",
			input: `person : ["name": "Alice" "age": 30];`,
			// Bindings inside table have parens
			expected: `(person : [("name" : "Alice") ("age" : 30)])`,
		},
		{
			name: "Tables - Nested",
			// Added space for decimal vs dot ambiguity (lexer prefers decimal 1.0)
			input:    "matrix : [[1 2] [3 4]]; val : matrix.1 . 0;",
			expected: "(matrix : [[1 2] [3 4]])\n(val : ((matrix.1).0))",
		},
		{
			name:  "Tables - Concatenation",
			input: "list1:[1]; list2:[2]; combined: list1, list2;",
			// Comma (60) < Colon (80) is FALSE. Colon (80) > Comma (60).
			// So BindingExpr consumes until comma check fails.
			// But 60 < 79 (Colon RBP).
			// So Colon SHOULD consume comma?
			// Wait. getBindingPower(,). 60.
			// minBP = 79.
			// 60 <= 79? YES.
			// So loop BREAKS.
			// So BindingExpr ends left of comma.
			// Top level consumes comma.
			// So ((combined : list1) , list2)
			expected: "(list1 : [1])\n(list2 : [2])\n((combined : list1) , list2)",
		},

		// 04_flow.org
		{
			name:  "Flow - Logic Chaining",
			input: "check : { right > 0 }; valid : check 10;",
			// Prefix expr now has space: (check 10)
			expected: "(check : { (right > 0) })\n(valid : (check 10))",
		},
		{
			name:  "Flow - Conditional ?",
			input: "res : (1 > 0) ? [true: \"Ok\" false: \"No\"];",
			// Input parens (1>0) -> inner parens.
			// Table bindings -> parens.
			expected: "(res : (((1 > 0)) ? [(true : \"Ok\") (false : \"No\")]))",
		},
		{
			name:     "Flow - Error Coalescing ??",
			input:    "invalid:0; safe : invalid ?? 0;",
			expected: "(invalid : 0)\n(safe : (invalid ?? 0))",
		},
		{
			name:     "Flow - Elvis ?:",
			input:    "zero:0; default : zero ?: 100;",
			expected: "(zero : 0)\n(default : (zero ?: 100))",
		},

		// 06_advanced.org
		{
			name: "Advanced - Pipelines",
			// Spaces in left + right.
			// add_10 is value, so add_10 5 is two statements.
			input:    "add:{left + right}; add_10: 10 |> add; res: add_10 5;",
			expected: "(add : { (left + right) })\n(add_10 : (10 |> add))\n(res : add_10)\n5",
		},
		{
			name:  "Advanced - Custom Precedence",
			input: "pow_op : 600{ left ** right }601; res : 2 pow_op 3 * 2;",
			// RBP parsing fixed.
			expected: "(pow_op : 600{ (left ** right) }601)\n(res : ((2 pow_op 3) * 2))",
		},
		{
			name:  "Advanced - Right Associativity Custom",
			input: "pow_op : 600{ left ** right }601; res : 2 pow_op 3 pow_op 2;",
			// RBP (601) > LBP (600) implies Left Associativity in this Pratt implementation.
			expected: "(pow_op : 600{ (left ** right) }601)\n(res : ((2 pow_op 3) pow_op 2))",
		},
		// Extended Assignments
		{
			name:     "Extended Assignment - Addition",
			input:    "count :+ 1;",
			expected: "(count :+ 1)",
		},
		{
			name:     "Extended Assignment - Multiplication",
			input:    "x :* 2;",
			expected: "(x :* 2)",
		},
		{
			name:     "Extended Assignment - Power",
			input:    "val :** 3;",
			expected: "(val :** 3)",
		},
		{
			name:     "Extended Assignment - Bitwise OR",
			input:    "flags :| 1;",
			expected: "(flags :| 1)",
		},
		{
			name:     "Extended Assignment - Right Shift",
			input:    "v :>> 2;",
			expected: "(v :>> 2)",
		},
		{
			name:     "Extended Assignment - Unsigned Right Shift",
			input:    "v :>>> 2;",
			expected: "(v :>>> 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]byte(tt.input))
			p := New(l)
			prog := p.ParseProgram()

			checkErrors(t, p)

			actual := prog.String()
			actual = trim(actual)
			expected := trim(tt.expected)

			if actual != expected {
				t.Errorf("expected:\n%q\ngot:\n%q", expected, actual)
			}
		})
	}
}
