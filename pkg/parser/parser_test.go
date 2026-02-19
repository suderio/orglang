package parser

import (
	"testing"

	"orglang/pkg/lexer"
)

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Integer Literal",
			input:    "5;",
			expected: "5",
		},
		{
			name:     "Decimal Literal",
			input:    "5.5;",
			expected: "5.5",
		},
		{
			name:     "Rational Literal",
			input:    "5/2;",
			expected: "5/2",
		},
		{
			name:     "Prefix Expression",
			input:    "- 5;", // Space to ensure prefix operator, not negative number
			expected: "(- 5)",
		},
		{
			name:     "Negative Integer",
			input:    "-5;",
			expected: "-5", // Parsed as IntegerLiteral
		},
		{
			name:     "Infix Expression",
			input:    "5 + 5;",
			expected: "(5 + 5)",
		},
		{
			name:     "Operator Precedence",
			input:    "5 + 5 * 2;",
			expected: "(5 + (5 * 2))",
		},
		{
			name:     "Operator Precedence 2",
			input:    "(5 + 5) * 2;",
			expected: "(((5 + 5)) * 2)", // GroupExpr adds parens ((...))
		},
		{
			name:     "Right Associativity",
			input:    "a:1; b:2; c:3; a : b : c;",
			expected: "(a : 1)\n(b : 2)\n(c : 3)\n(a : (b : c))",
		},
		{
			name:     "Dot Expression",
			input:    "a:1; b:2; a.b;",
			expected: "(a : 1)\n(b : 2)\n(a.b)",
		},
		{
			name:     "Method Call chain",
			input:    "a:1; b:2; c:3; a.b.c;",
			expected: "(a : 1)\n(b : 2)\n(c : 3)\n((a.b).c)",
		},
		{
			name:     "Function Literal",
			input:    "{ 1 + 1 };",
			expected: "{ (1 + 1) }",
		},
		{
			name:     "Dynamic Binding Prefix",
			input:    "sq : { right * right }; sq 5;",
			expected: "(sq : { (right * right) })\n(sq 5)",
		},
		{
			name:     "Dynamic Binding Infix",
			input:    "add : { left + right }; 1 add 2;",
			expected: "(add : { (left + right) })\n(1 add 2)",
		},
		{
			name:     "Pipe Operator",
			input:    "inc:1; 10 |> inc;",
			expected: "(inc : 1)\n(10 |> inc)",
		},
		{
			name:     "Pipe Operator Block",
			input:    "10 |> { left + 1 };",
			expected: "(10 |> { (left + 1) })",
		},
		{
			name:     "Pipe Operator Group",
			input:    "ops:1; inc:1; 10 |> (ops.inc);",
			expected: "(ops : 1)\n(inc : 1)\n(10 |> ((ops.inc)))",
		},
		{
			name:     "Compose Operator",
			input:    "double:1; inc:1; double o inc;",
			expected: "(double : 1)\n(inc : 1)\n(double o inc)",
		},
		{
			name:     "Table Literal Space",
			input:    "[1 2 3];",
			expected: "[1 2 3]",
		},
		{
			name:     "Table Literal Comma",
			input:    "[1, 2];",
			expected: "[(1 , 2)]",
		},
		{
			name:     "Table Literal Mixed",
			input:    "[1 + 2 3];",
			expected: "[1 + 2 3]",
		},
		{
			name:     "Table Literal with Binding",
			input:    "a:1; [a: 1];",
			expected: "(a : 1)\n[(a : 1)]",
		},
		{
			name:     "Elvis Operator",
			input:    "a:1; b:2; a ?: b;",
			expected: "(a : 1)\n(b : 2)\n(a ?: b)",
		},
		{
			name:     "Resource Definition",
			input:    "Log @: {};",
			expected: "(Log @: {  })",
		},
		{
			name:     "Resource Instantiation",
			input:    "Log @: {}; @Log;",
			expected: "(Log @: {  })\n(@ Log)",
		},
		{
			name:     "Implicit Semicolon/lines",
			input:    "1\n2",
			expected: "1\n2",
		},
		{
			name:     "String Literals",
			input:    `"hello" """doc""" 'raw'`,
			expected: "\"hello\"\n\"\"\"doc\"\"\"\n'raw'",
		},
		{
			name:     "Boolean",
			input:    "true false",
			expected: "true\nfalse",
		},
		// New Tests High Complexity
		{
			name:     "Really Big Integer",
			input:    "1234567890123456789012345678901234567890;",
			expected: "1234567890123456789012345678901234567890",
		},
		{
			name:     "Really Big Decimal",
			input:    "12345678901234567890.12345678901234567890;",
			expected: "12345678901234567890.12345678901234567890",
		},
		{
			name:     "Really Big Rational",
			input:    "12345678901234567890/98765432109876543210;",
			expected: "12345678901234567890/98765432109876543210",
		},
		{
			name:     "Unicode Strings",
			input:    `"Hello 世界 🌍" "こんにちは" "💩"`,
			expected: "\"Hello 世界 🌍\"\n\"こんにちは\"\n\"💩\"",
		},
		// Pre-Runtime Fix Tests
		{
			name:     "Not Equal ~=",
			input:    "a:1; b:2; a ~= b;",
			expected: "(a : 1)\n(b : 2)\n(a ~= b)",
		},
		{
			name:     "Not Equal <> and ~= same precedence",
			input:    "a:1; b:2; a <> b;",
			expected: "(a : 1)\n(b : 2)\n(a <> b)",
		},
		{
			name:     "Infix @ Module Loading",
			input:    `lib:1; "path" @ lib;`,
			expected: "(lib : 1)\n(\"path\" @ lib)",
		},
		{
			name:     "Pipe with Docstring Atom",
			input:    "fn:1; fn |> \"\"\"doc\"\"\";",
			expected: "(fn : 1)\n(fn |> \"\"\"doc\"\"\")",
		},
		{
			name:     "Pipe with Raw String Atom",
			input:    "fn:1; fn |> 'raw';",
			expected: "(fn : 1)\n(fn |> 'raw')",
		},
		// Short circuit tests
		{
			name:     "Short Circuit AND",
			input:    "true && false",
			expected: "(true && false)",
		},
		{
			name:     "Short Circuit OR",
			input:    "true || false",
			expected: "(true || false)",
		},
		{
			name:     "Short Circuit Precedence (AND higher than OR)",
			input:    "true && false || 1", // (true && false) || 1
			expected: "((true && false) || 1)",
		},
		{
			name:     "Short Circuit Precedence (OR lower than AND)",
			input:    "true || false && 1", // true || (false && 1)
			expected: "(true || (false && 1))",
		},
		{
			name:     "Short Circuit vs Equality",
			input:    "true = false && 1", // (true = false) && 1 falseecause = is 150, && is 140
			expected: "((true = false) && 1)",
		},
		{
			name:     "Short Circuit vs Comparison",
			input:    "true < false || 1 > []", // (true < false) || (1 > []). < > are 150, || is 130
			expected: "((true < false) || (1 > []))",
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
		// Missing Operator Tests
		{
			name:     "Bitwise AND",
			input:    "a:10; b:2; res: a & b;",
			expected: "(a : 10)\n(b : 2)\n(res : (a & b))",
		},
		{
			name:     "Bitwise OR",
			input:    "a:10; b:2; res: a | b;",
			expected: "(a : 10)\n(b : 2)\n(res : (a | b))",
		},
		{
			name:     "Bitwise XOR",
			input:    "a:10; b:2; res: a ^ b;",
			expected: "(a : 10)\n(b : 2)\n(res : (a ^ b))",
		},
		{
			name:     "Bitwise Left Shift",
			input:    "a:1; c:2; res: a << c;",
			expected: "(a : 1)\n(c : 2)\n(res : (a << c))",
		},
		{
			name:     "Bitwise Right Shift",
			input:    "a:8; c:1; res: a >> c;",
			expected: "(a : 8)\n(c : 1)\n(res : (a >> c))",
		},
		{
			name:     "Comparison Less Equal",
			input:    "a:1; b:2; res: a <= b;",
			expected: "(a : 1)\n(b : 2)\n(res : (a <= b))",
		},
		{
			name:     "Comparison Greater Equal",
			input:    "a:1; b:2; res: a >= b;",
			expected: "(a : 1)\n(b : 2)\n(res : (a >= b))",
		},
		{
			name:     "Logical NOT",
			input:    "t:true; res: ! t;",
			expected: "(t : true)\n(res : (! t))",
		},
		{
			name:  "Flux Operator",
			input: "a:1; b:{}; res: a -> b;",
			// Flux (50) < Binding (80), so binding happens first!
			// a -> b is parsed as (res : a) -> b ?
			// Wait, binding is Right Assoc.
			// res : a -> b.
			// : has BP 80. -> has BP 50.
			// Parser calls parseExpression(0).
			// nud(res).
			// led(:). BP 80.
			// recurse parseExpression(79) (Right Assoc).
			// inside: nud(a).
			// look ahead -> (BP 50).
			// 50 < 79. Stop!
			// So val is 'a'.
			// Result: (res : a).
			// Loop continues in outer parseExpression.
			// next op is ->.
			// led(->, (res:a)).
			// recurse parseExpression(50).
			// nud(b).
			// Result: ((res : a) -> b).
			// This seems correct according to precedence table!
			expected: "(a : 1)\n(b : {  })\n((res : a) -> b)",
		},
		{
			name:     "Dispatch Operator",
			input:    "a:1; b:[]; res: a -< b;",
			expected: "(a : 1)\n(b : [])\n((res : a) -< b)",
		},
		{
			name:     "Join Operator",
			input:    "a:[]; b:{}; res: a -<> b;",
			expected: "(a : [])\n(b : {  })\n((res : a) -<> b)",
		},
		{
			name: "Increment Operator",
			// Space required?
			// "++i" lexed as identifier "++i" if no space?
			// Lexer: isIdentStart('+') -> yes.
			// isIdentContinue('i') -> yes.
			// So "++i" is ONE identifier.
			// And "++i" is undefined.
			// We must use space! "++ i"
			input:    "i:0; ++ i;",
			expected: "(i : 0)\n(++ i)",
		},
		{
			name:     "Decrement Operator",
			input:    "i:0; -- i;",
			expected: "(i : 0)\n(-- i)",
		},
		{
			name:     "String Interpolation",
			input:    `fmt: "Hello $0"; args: ["World"]; res: fmt $ args;`,
			expected: "(fmt : \"Hello $0\")\n(args : [\"World\"])\n(res : (fmt $ args))",
		},
		// Extended Assignments
		{
			name:     "Extended Assignment - Subtract",
			input:    "i:10; i :- 1;",
			expected: "(i : 10)\n(i :- 1)",
		},
		{
			name:     "Extended Assignment - Divide",
			input:    "i:10; i :/ 2;",
			expected: "(i : 10)\n(i :/ 2)",
		},
		{
			name:     "Extended Assignment - Modulo",
			input:    "i:10; i :% 3;",
			expected: "(i : 10)\n(i :% 3)",
		},
		{
			name:     "Extended Assignment - Bitwise AND",
			input:    "f:15; f :& 1;",
			expected: "(f : 15)\n(f :& 1)",
		},
		{
			name:     "Extended Assignment - Bitwise XOR",
			input:    "f:15; f :^ 1;",
			expected: "(f : 15)\n(f :^ 1)",
		},
		{
			name:     "Extended Assignment - Bitwise NOT",
			input:    "f:0; f :~ 0;",
			expected: "(f : 0)\n(f :~ 0)",
		},
		{
			name:     "Extended Assignment - Left Shift",
			input:    "v:1; v :<< 1;",
			expected: "(v : 1)\n(v :<< 1)",
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

func checkErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}
	t.FailNow()
}

func trim(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
