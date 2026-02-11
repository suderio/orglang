package lexer

import (
	"orglang/pkg/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `x : 42;
y : x + 8;
name : "John";
pi : 3.14;
[1 2 3];
{ right + 1 };
@stdout;
"path" @ file;
-> -< -<> ?? ?: |> o ~
-1 +2.5 -3.14 +42
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.COLON, ":"},
		{token.INT, "42"},
		{token.SEMICOLON, ";"},
		// y : x + 8;
		{token.IDENT, "y"},
		{token.COLON, ":"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.INT, "8"},
		{token.SEMICOLON, ";"},
		// name : "John";
		{token.IDENT, "name"},
		{token.COLON, ":"},
		{token.STRING, "John"},
		{token.SEMICOLON, ";"},
		// pi : 3.14;
		{token.IDENT, "pi"},
		{token.COLON, ":"},
		{token.FLOAT, "3.14"},
		{token.SEMICOLON, ";"},
		// [1 2 3];
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.INT, "2"},
		{token.INT, "3"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		// { right + 1 };
		{token.LBRACE, "{"},
		{token.IDENT, "right"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		// @stdout;
		{token.AT, "@"},
		{token.IDENT, "stdout"},
		{token.SEMICOLON, ";"},
		// "path" @ file;
		{token.STRING, "path"},
		{token.AT, "@"},
		{token.IDENT, "file"},
		{token.SEMICOLON, ";"},
		// -> -< -<> ?? ?: |> o ~
		{token.ARROW, "->"},
		{token.DARROW, "-<"},
		{token.JOIN_ARROW, "-<>"},
		{token.ERROR_CHECK, "??"},
		{token.ELVIS, "?:"},
		{token.PIPE, "|>"},
		{token.COMPOSE, "o"},
		{token.NOT, "~"},

		// Signed Numbers
		{token.INT, "-1"},
		{token.FLOAT, "+2.5"},
		{token.FLOAT, "-3.14"},
		{token.INT, "+42"},

		{token.EOF, ""},
	}

	l := NewCustom(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
