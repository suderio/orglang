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
