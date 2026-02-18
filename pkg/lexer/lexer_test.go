package lexer

import (
	"os"
	"path/filepath"
	"testing"

	"orglang/pkg/token"
)

// helper to lex input and return all tokens (including EOF)
func lexAll(input string) []token.Token {
	l := New([]byte(input))
	return l.Tokenize()
}

// helper to assert a specific token at a given index
func assertToken(t *testing.T, tokens []token.Token, idx int, expectedType token.TokenType, expectedLiteral string) {
	t.Helper()
	if idx >= len(tokens) {
		t.Fatalf("expected token at index %d, but only got %d tokens", idx, len(tokens))
	}
	tok := tokens[idx]
	if tok.Type != expectedType {
		t.Errorf("token[%d]: expected type %s, got %s (literal=%q)", idx, expectedType, tok.Type, tok.Literal)
	}
	if tok.Literal != expectedLiteral {
		t.Errorf("token[%d]: expected literal %q, got %q (type=%s)", idx, expectedLiteral, tok.Literal, tok.Type)
	}
}

func assertTokenCount(t *testing.T, tokens []token.Token, expected int) {
	t.Helper()
	if len(tokens) != expected {
		t.Fatalf("expected %d tokens, got %d", expected, len(tokens))
	}
}

// --- Delimiters ---

func TestDelimiters(t *testing.T) {
	tokens := lexAll("( ) [ ] { } ;")
	assertTokenCount(t, tokens, 8) // 7 delimiters + EOF
	assertToken(t, tokens, 0, token.LPAREN, "(")
	assertToken(t, tokens, 1, token.RPAREN, ")")
	assertToken(t, tokens, 2, token.LBRACKET, "[")
	assertToken(t, tokens, 3, token.RBRACKET, "]")
	assertToken(t, tokens, 4, token.LBRACE, "{")
	assertToken(t, tokens, 5, token.RBRACE, "}")
	assertToken(t, tokens, 6, token.SEMICOLON, ";")
	assertToken(t, tokens, 7, token.EOF, "")
}

func TestDelimitersNoSpaces(t *testing.T) {
	tokens := lexAll("()[]{};")
	assertTokenCount(t, tokens, 8)
	assertToken(t, tokens, 0, token.LPAREN, "(")
	assertToken(t, tokens, 1, token.RPAREN, ")")
	assertToken(t, tokens, 2, token.LBRACKET, "[")
	assertToken(t, tokens, 3, token.RBRACKET, "]")
	assertToken(t, tokens, 4, token.LBRACE, "{")
	assertToken(t, tokens, 5, token.RBRACE, "}")
	assertToken(t, tokens, 6, token.SEMICOLON, ";")
}

// --- Structural Operators ---

func TestStructuralOperators(t *testing.T) {
	tokens := lexAll("@ @: : . ,")
	assertTokenCount(t, tokens, 6)
	assertToken(t, tokens, 0, token.AT, "@")
	assertToken(t, tokens, 1, token.AT_COLON, "@:")
	assertToken(t, tokens, 2, token.COLON, ":")
	assertToken(t, tokens, 3, token.DOT, ".")
	assertToken(t, tokens, 4, token.COMMA, ",")
}

func TestAtColonAdjacent(t *testing.T) {
	// @:value should lex as AT_COLON IDENTIFIER(value)
	tokens := lexAll("@:value")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.AT_COLON, "@:")
	assertToken(t, tokens, 1, token.IDENTIFIER, "value")
}

func TestAtAlone(t *testing.T) {
	tokens := lexAll("@stdout")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.AT, "@")
	assertToken(t, tokens, 1, token.IDENTIFIER, "stdout")
}

// --- Elvis ---

func TestElvis(t *testing.T) {
	tokens := lexAll("?:")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ELVIS, "?:")
}

func TestElvisInExpression(t *testing.T) {
	tokens := lexAll("x ?: y")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.ELVIS, "?:")
	assertToken(t, tokens, 2, token.IDENTIFIER, "y")
}

func TestQuestionMarkNotElvis(t *testing.T) {
	// isValid? followed by space shouldn't be ELVIS
	tokens := lexAll("isValid? x")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "isValid?")
	assertToken(t, tokens, 1, token.IDENTIFIER, "x")
}

func TestQuestionMarkAlone(t *testing.T) {
	tokens := lexAll("?")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.IDENTIFIER, "?")
}

// --- Extended Assignment Operators ---

func TestExtendedAssignment(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{":+", ":+"},
		{":-", ":-"},
		{":*", ":*"},
		{":/", ":/"},
		{":%", ":%"},
		{":&", ":&"},
		{":^", ":^"},
		{":|", ":|"},
		{":~", ":~"},
		{":>>", ":>>"},
		{":<<", ":<<"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.IDENTIFIER, tt.literal)
		})
	}
}

func TestColonGreaterNotShift(t *testing.T) {
	// :> (single >) should still lex as identifier
	tokens := lexAll(":>")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.IDENTIFIER, ":>")
}

func TestColonLessNotShift(t *testing.T) {
	// :< (single <) should still lex as identifier
	tokens := lexAll(":<")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.IDENTIFIER, ":<")
}

// --- Integers ---

func TestIntegers(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"42", "42"},
		{"0", "0"},
		{"123456789012345678901234567890", "123456789012345678901234567890"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.INTEGER, tt.literal)
		})
	}
}

// --- Sign Gluing ---

func TestSignGluingPrefix(t *testing.T) {
	// At start of file, -42 is one token
	tokens := lexAll("-42")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.INTEGER, "-42")
}

func TestSignGluingPrefixPlus(t *testing.T) {
	tokens := lexAll("+3")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.INTEGER, "+3")
}

func TestSignGluingInfix(t *testing.T) {
	// After an identifier, - is infix (identifier)
	tokens := lexAll("x - 42")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "-")
	assertToken(t, tokens, 2, token.INTEGER, "42")
}

func TestSignGluingAfterLParen(t *testing.T) {
	tokens := lexAll("(-3)")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.LPAREN, "(")
	assertToken(t, tokens, 1, token.INTEGER, "-3")
	assertToken(t, tokens, 2, token.RPAREN, ")")
}

func TestSignGluingAfterColon(t *testing.T) {
	tokens := lexAll("x : -5")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.COLON, ":")
	assertToken(t, tokens, 2, token.INTEGER, "-5")
}

func TestSignGluingAfterSemicolon(t *testing.T) {
	tokens := lexAll("; -5")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.SEMICOLON, ";")
	assertToken(t, tokens, 1, token.INTEGER, "-5")
}

func TestSignGluingAfterComma(t *testing.T) {
	tokens := lexAll("[1, -2]")
	assertTokenCount(t, tokens, 6)
	assertToken(t, tokens, 0, token.LBRACKET, "[")
	assertToken(t, tokens, 1, token.INTEGER, "1")
	assertToken(t, tokens, 2, token.COMMA, ",")
	assertToken(t, tokens, 3, token.INTEGER, "-2")
	assertToken(t, tokens, 4, token.RBRACKET, "]")
}

func TestSignNotGluedWithSpace(t *testing.T) {
	// - 42 (space between) at start of file: sign glue looks for adjacent digit
	tokens := lexAll("- 42")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "-")
	assertToken(t, tokens, 1, token.INTEGER, "42")
}

// --- Decimals ---

func TestDecimals(t *testing.T) {
	tests := []struct {
		input   string
		literal string
	}{
		{"3.14", "3.14"},
		{"0.001", "0.001"},
		{"1.0", "1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.DECIMAL, tt.literal)
		})
	}
}

func TestDecimalSignGlued(t *testing.T) {
	tokens := lexAll("-3.14")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.DECIMAL, "-3.14")
}

func TestDecimalDisambiguation_DotOnly(t *testing.T) {
	// 1. → INTEGER + DOT
	tokens := lexAll("1.")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.INTEGER, "1")
	assertToken(t, tokens, 1, token.DOT, ".")
}

func TestDecimalDisambiguation_DotDigit(t *testing.T) {
	// .5 → DOT + INTEGER
	tokens := lexAll(".5")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.DOT, ".")
	assertToken(t, tokens, 1, token.INTEGER, "5")
}

// --- Rationals ---

func TestRationals(t *testing.T) {
	tokens := lexAll("1/2")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.RATIONAL, "1/2")
}

func TestRationalSigned(t *testing.T) {
	tokens := lexAll("-3/4")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.RATIONAL, "-3/4")
}

func TestRationalNotDivision(t *testing.T) {
	// With spaces, it's not a rational
	tokens := lexAll("1 / 2")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.INTEGER, "1")
	assertToken(t, tokens, 1, token.IDENTIFIER, "/")
	assertToken(t, tokens, 2, token.INTEGER, "2")
}

func TestRationalInExpression(t *testing.T) {
	tokens := lexAll("x : 1/2;")
	assertTokenCount(t, tokens, 5)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.COLON, ":")
	assertToken(t, tokens, 2, token.RATIONAL, "1/2")
	assertToken(t, tokens, 3, token.SEMICOLON, ";")
}

// --- Strings ---

func TestStringSimple(t *testing.T) {
	tokens := lexAll(`"hello"`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.STRING, "hello")
}

func TestStringEmpty(t *testing.T) {
	tokens := lexAll(`""`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.STRING, "")
}

func TestStringEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", `"a\nb"`, "a\nb"},
		{"tab", `"a\tb"`, "a\tb"},
		{"carriage return", `"a\rb"`, "a\rb"},
		{"backslash", `"a\\b"`, "a\\b"},
		{"double quote", `"a\"b"`, "a\"b"},
		{"null", `"a\0b"`, "a\x00b"},
		{"unicode bmp", `"\u0041"`, "A"},
		{"unicode braced", `"\u{1F600}"`, "\U0001F600"},
		{"unicode braced short", `"\u{41}"`, "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.STRING, tt.expected)
		})
	}
}

func TestStringUnknownEscape(t *testing.T) {
	tokens := lexAll(`"a\xb"`)
	// ILLEGAL(\x), IDENTIFIER(b), ILLEGAL(unterminated string), EOF => 4 tokens
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.ILLEGAL, `unknown escape: \x`)
}

func TestStringUnterminated(t *testing.T) {
	tokens := lexAll(`"hello`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated string")
}

func TestStringUnterminatedEscape(t *testing.T) {
	tokens := lexAll(`"hello\`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated escape sequence")
}

func TestUnicodeEscapeEmpty(t *testing.T) {
	tokens := lexAll(`"\u{}"`)
	// ILLEGAL(empty), ILLEGAL(unterminated string), EOF => 3 tokens
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.ILLEGAL, `empty unicode escape \u{}`)
}

func TestUnicodeEscapeOutOfRange(t *testing.T) {
	tokens := lexAll(`"\u{FFFFFF}"`)
	// ILLEGAL(out of range), ILLEGAL(unterminated string), EOF => 3 tokens
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.ILLEGAL, "unicode codepoint out of range: U+FFFFFF")
}

func TestUnicodeEscapeTooLong(t *testing.T) {
	tokens := lexAll(`"\u{1234567}"`)
	// ILLEGAL(too long), RBRACE(}), ILLEGAL(unterminated string), EOF => 4 tokens
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.ILLEGAL, "unicode escape too long (max 6 hex digits)")
}

func TestUnicodeEscapeInvalidHexBraced(t *testing.T) {
	tokens := lexAll(`"\u{GG}"`)
	// ILLEGAL(invalid hex), IDENTIFIER(GG), RBRACE(}), ILLEGAL(unterminated string), EOF => 5 tokens
	assertTokenCount(t, tokens, 5)
	assertToken(t, tokens, 0, token.ILLEGAL, "invalid hex digit in unicode escape: G")
}

func TestUnicodeEscapeInvalidHex4(t *testing.T) {
	tokens := lexAll(`"\u00GG"`)
	// ILLEGAL(invalid hex), IDENTIFIER(G), ILLEGAL(unterminated string), EOF => 4 tokens
	assertTokenCount(t, tokens, 4)
	if tokens[0].Type != token.ILLEGAL {
		t.Errorf("expected ILLEGAL, got %s", tokens[0].Type)
	}
}

func TestUnicodeEscapeUnterminated4(t *testing.T) {
	tokens := lexAll(`"\u00`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated unicode escape \\uXXXX")
}

func TestUnicodeEscapeUnterminatedBraced(t *testing.T) {
	tokens := lexAll(`"\u{41`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, `unterminated unicode escape \u{...}`)
}

func TestUnicodeEscapeUnterminatedU(t *testing.T) {
	tokens := lexAll(`"\u`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated unicode escape")
}

// --- Docstrings ---

func TestDocstring(t *testing.T) {
	input := "\"\"\"" + "\n  hello\n  world\n" + "\"\"\""
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.DOCSTRING, "hello\nworld")
}

func TestDocstringUnterminated(t *testing.T) {
	tokens := lexAll("\"\"\"hello")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated docstring")
}

func TestDocstringWithEscapes(t *testing.T) {
	input := "\"\"\"" + "\nhello\\nworld\n" + "\"\"\""
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.DOCSTRING, "hello\nworld")
}

func TestDocstringEmpty(t *testing.T) {
	input := "\"\"\"\"\"\""
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.DOCSTRING, "")
}

func TestDocstringBadEscape(t *testing.T) {
	input := "\"\"\"\\x\"\"\""
	tokens := lexAll(input)
	// ILLEGAL(\unknown escape), DOCSTRING(""), EOF => 3 tokens
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.ILLEGAL, `unknown escape: \x`)
}

// --- Raw Strings ---

func TestRawString(t *testing.T) {
	tokens := lexAll(`'hello\nworld'`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.RAWSTRING, `hello\nworld`)
}

func TestRawStringUnterminated(t *testing.T) {
	tokens := lexAll(`'hello`)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated raw string")
}

// --- Raw Docstrings ---

func TestRawDocstring(t *testing.T) {
	input := "'''" + "\n  raw\n  text\n" + "'''"
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.RAWDOC, "raw\ntext")
}

func TestRawDocstringUnterminated(t *testing.T) {
	tokens := lexAll("'''hello")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "unterminated raw docstring")
}

func TestRawDocstringNoEscapes(t *testing.T) {
	input := "'''" + "\n\\n\\t\n" + "'''"
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.RAWDOC, `\n\t`)
}

// --- Booleans ---

func TestBooleans(t *testing.T) {
	tokens := lexAll("true false")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.BOOLEAN, "true")
	assertToken(t, tokens, 1, token.BOOLEAN, "false")
}

// --- Keywords ---

func TestKeywords(t *testing.T) {
	tokens := lexAll("this left right")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.KEYWORD, "this")
	assertToken(t, tokens, 1, token.KEYWORD, "left")
	assertToken(t, tokens, 2, token.KEYWORD, "right")
}

// --- Identifiers ---

func TestIdentifiersSimple(t *testing.T) {
	tokens := lexAll("x add foo_bar")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "add")
	assertToken(t, tokens, 2, token.IDENTIFIER, "foo_bar")
}

func TestIdentifierOperators(t *testing.T) {
	tests := []string{"->", "++", "<=", "&&", "??", "**", "|>", "-<", "-<>", "||", "!", "~", "&", "|", "^", "<<", ">>", "~=", "<>", "o"}
	for _, op := range tests {
		t.Run(op, func(t *testing.T) {
			tokens := lexAll(op)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.IDENTIFIER, op)
		})
	}
}

func TestIdentifierWithDigits(t *testing.T) {
	tokens := lexAll("x1 item2 count3")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x1")
	assertToken(t, tokens, 1, token.IDENTIFIER, "item2")
	assertToken(t, tokens, 2, token.IDENTIFIER, "count3")
}

// --- Unicode Identifiers ---

func TestUnicodeIdentifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"café", "café"},
		{"π", "π"},
		{"Σ", "Σ"},
		{"名前", "名前"},
		{"_foo", "_foo"},
		{"αβγ", "αβγ"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.IDENTIFIER, tt.expected)
		})
	}
}

func TestUnicodeSymbols(t *testing.T) {
	// Symbol category characters are valid identifiers
	tests := []string{"∑", "√", "∞", "€", "→"}
	for _, sym := range tests {
		t.Run(sym, func(t *testing.T) {
			tokens := lexAll(sym)
			assertTokenCount(t, tokens, 2)
			assertToken(t, tokens, 0, token.IDENTIFIER, sym)
		})
	}
}

// --- Comments ---

func TestLineComment(t *testing.T) {
	tokens := lexAll("x # this is a comment\ny")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "y")
}

func TestBlockComment(t *testing.T) {
	input := "x\n###\nthis is a block comment\nwith multiple lines\n###\ny"
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "y")
}

func TestBlockCommentDoesNotMatchMidLine(t *testing.T) {
	// ### not at column 1 should be treated as line comment
	tokens := lexAll("x ### comment")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
}

func TestCommentOnly(t *testing.T) {
	tokens := lexAll("# just a comment")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

func TestBlockCommentOnly(t *testing.T) {
	tokens := lexAll("###\nblock\n###")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

// --- Binding Power Adjacency ---

func TestBindingPowerAdjacency(t *testing.T) {
	// 50{...}60 — adjacency should preserve position info
	tokens := lexAll("50{left + right}60")
	assertTokenCount(t, tokens, 8)
	assertToken(t, tokens, 0, token.INTEGER, "50")
	assertToken(t, tokens, 1, token.LBRACE, "{")
	assertToken(t, tokens, 2, token.KEYWORD, "left")
	assertToken(t, tokens, 3, token.IDENTIFIER, "+")
	assertToken(t, tokens, 4, token.KEYWORD, "right")
	assertToken(t, tokens, 5, token.RBRACE, "}")
	assertToken(t, tokens, 6, token.INTEGER, "60")
}

// --- Whitespace ---

func TestWhitespace(t *testing.T) {
	tokens := lexAll("  x  \t  y  \n  z  ")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "y")
	assertToken(t, tokens, 2, token.IDENTIFIER, "z")
}

func TestCarriageReturn(t *testing.T) {
	tokens := lexAll("x\r\ny")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.IDENTIFIER, "y")
}

// --- Edge Cases ---

func TestEmptyInput(t *testing.T) {
	tokens := lexAll("")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

func TestOnlyWhitespace(t *testing.T) {
	tokens := lexAll("   \n\t  \n  ")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

func TestIllegalCharacter(t *testing.T) {
	// Backslash outside a string is illegal
	tokens := lexAll("\\")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.ILLEGAL, "\\")
}

// --- Position Tracking ---

func TestPositionTracking(t *testing.T) {
	tokens := lexAll("a b\nc")
	// a: line 1, col 1
	// b: line 1, col 3
	// c: line 2, col 1
	if tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Errorf("token 'a': expected 1:1, got %d:%d", tokens[0].Line, tokens[0].Column)
	}
	if tokens[1].Line != 1 || tokens[1].Column != 3 {
		t.Errorf("token 'b': expected 1:3, got %d:%d", tokens[1].Line, tokens[1].Column)
	}
	if tokens[2].Line != 2 || tokens[2].Column != 1 {
		t.Errorf("token 'c': expected 2:1, got %d:%d", tokens[2].Line, tokens[2].Column)
	}
}

// --- Structural Breaking ---

func TestStructuralBreaking(t *testing.T) {
	tests := []struct {
		name  string
		input string
		types []token.TokenType
		lits  []string
	}{
		{
			"colon adjacent",
			"x:42",
			[]token.TokenType{token.IDENTIFIER, token.COLON, token.INTEGER, token.EOF},
			[]string{"x", ":", "42", ""},
		},
		{
			"function call",
			"f(x)",
			[]token.TokenType{token.IDENTIFIER, token.LPAREN, token.IDENTIFIER, token.RPAREN, token.EOF},
			[]string{"f", "(", "x", ")", ""},
		},
		{
			"dot access",
			"list.0",
			[]token.TokenType{token.IDENTIFIER, token.DOT, token.INTEGER, token.EOF},
			[]string{"list", ".", "0", ""},
		},
		{
			"resource",
			"@stdout",
			[]token.TokenType{token.AT, token.IDENTIFIER, token.EOF},
			[]string{"@", "stdout", ""},
		},
		{
			"semicolon adjacent",
			"x;y",
			[]token.TokenType{token.IDENTIFIER, token.SEMICOLON, token.IDENTIFIER, token.EOF},
			[]string{"x", ";", "y", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := lexAll(tt.input)
			assertTokenCount(t, tokens, len(tt.types))
			for i, expectedType := range tt.types {
				assertToken(t, tokens, i, expectedType, tt.lits[i])
			}
		})
	}
}

// --- Complex Expressions ---

func TestComplexExpression(t *testing.T) {
	input := `main: {@args -> "Hello, World!" -> @stdout};`
	tokens := lexAll(input)
	// main : { @args -> "Hello, World!" -> @stdout } ;
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.IDENTIFIER, "main"},
		{token.COLON, ":"},
		{token.LBRACE, "{"},
		{token.AT, "@"},
		{token.IDENTIFIER, "args"},
		{token.IDENTIFIER, "->"},
		{token.STRING, "Hello, World!"},
		{token.IDENTIFIER, "->"},
		{token.AT, "@"},
		{token.IDENTIFIER, "stdout"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	assertTokenCount(t, tokens, len(expected))
	for i, exp := range expected {
		assertToken(t, tokens, i, exp.typ, exp.lit)
	}
}

func TestPipeAndCompose(t *testing.T) {
	input := `10 |> + o -`
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 6)
	assertToken(t, tokens, 0, token.INTEGER, "10")
	assertToken(t, tokens, 1, token.IDENTIFIER, "|>")
	assertToken(t, tokens, 2, token.IDENTIFIER, "+")
	assertToken(t, tokens, 3, token.IDENTIFIER, "o")
	assertToken(t, tokens, 4, token.IDENTIFIER, "-")
	assertToken(t, tokens, 5, token.EOF, "")
}

func TestResourceDefinition(t *testing.T) {
	input := `stdout @: [next: { right }];`
	tokens := lexAll(input)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.IDENTIFIER, "stdout"},
		{token.AT_COLON, "@:"},
		{token.LBRACKET, "["},
		{token.IDENTIFIER, "next"},
		{token.COLON, ":"},
		{token.LBRACE, "{"},
		{token.KEYWORD, "right"},
		{token.RBRACE, "}"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	assertTokenCount(t, tokens, len(expected))
	for i, exp := range expected {
		assertToken(t, tokens, i, exp.typ, exp.lit)
	}
}

func TestTableLiteral(t *testing.T) {
	input := `[1 2 3 name:"hello"]`
	tokens := lexAll(input)
	expected := []struct {
		typ token.TokenType
		lit string
	}{
		{token.LBRACKET, "["},
		{token.INTEGER, "1"},
		{token.INTEGER, "2"},
		{token.INTEGER, "3"},
		{token.IDENTIFIER, "name"},
		{token.COLON, ":"},
		{token.STRING, "hello"},
		{token.RBRACKET, "]"},
		{token.EOF, ""},
	}
	assertTokenCount(t, tokens, len(expected))
	for i, exp := range expected {
		assertToken(t, tokens, i, exp.typ, exp.lit)
	}
}

// --- Docstring Indent Stripping ---

func TestStripDocIndent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no indent", "hello\nworld", "hello\nworld"},
		{"uniform indent", "\n  hello\n  world\n", "hello\nworld"},
		{"mixed indent", "\n    hello\n  world\n", "  hello\nworld"},
		{"empty lines", "\n  hello\n\n  world\n", "hello\n\nworld"},
		{"empty", "", ""},
		{"single line", "\n  hello\n", "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripDocIndent(tt.input)
			if result != tt.expected {
				t.Errorf("stripDocIndent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// --- Integration: Example Files ---

func TestExampleFiles(t *testing.T) {
	examplesDir := filepath.Join("..", "..", "examples")
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Skipf("examples directory not found: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".org" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(examplesDir, entry.Name()))
			if err != nil {
				t.Fatalf("failed to read %s: %v", entry.Name(), err)
			}
			tokens := lexAll(string(content))
			for i, tok := range tokens {
				if tok.Type == token.ILLEGAL {
					t.Errorf("ILLEGAL token at index %d (line %d, col %d): %q",
						i, tok.Line, tok.Column, tok.Literal)
				}
			}
			// Should end with EOF
			last := tokens[len(tokens)-1]
			if last.Type != token.EOF {
				t.Errorf("expected last token to be EOF, got %s", last.Type)
			}
		})
	}
}

// --- Sign gluing edge cases ---

func TestSignAfterDot(t *testing.T) {
	tokens := lexAll("x. -3")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.DOT, ".")
	assertToken(t, tokens, 2, token.INTEGER, "-3")
}

func TestSignAfterLBracket(t *testing.T) {
	tokens := lexAll("[-1]")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.LBRACKET, "[")
	assertToken(t, tokens, 1, token.INTEGER, "-1")
	assertToken(t, tokens, 2, token.RBRACKET, "]")
}

func TestSignAfterElvis(t *testing.T) {
	tokens := lexAll("x ?: -1")
	assertTokenCount(t, tokens, 4)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
	assertToken(t, tokens, 1, token.ELVIS, "?:")
	assertToken(t, tokens, 2, token.INTEGER, "-1")
}

// --- Multiple tokens per line ---

func TestMultipleStatementsOneLine(t *testing.T) {
	tokens := lexAll("a : 1; b : 2;")
	assertTokenCount(t, tokens, 9)
	assertToken(t, tokens, 0, token.IDENTIFIER, "a")
	assertToken(t, tokens, 1, token.COLON, ":")
	assertToken(t, tokens, 2, token.INTEGER, "1")
	assertToken(t, tokens, 3, token.SEMICOLON, ";")
	assertToken(t, tokens, 4, token.IDENTIFIER, "b")
	assertToken(t, tokens, 5, token.COLON, ":")
	assertToken(t, tokens, 6, token.INTEGER, "2")
	assertToken(t, tokens, 7, token.SEMICOLON, ";")
}

// --- Block comment unterminated ---

func TestBlockCommentUnterminated(t *testing.T) {
	// Block comment that never closes just consumes to EOF
	tokens := lexAll("###\nthis never closes")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

// --- Consecutive comments ---

func TestConsecutiveComments(t *testing.T) {
	tokens := lexAll("# comment 1\n# comment 2\nx")
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.IDENTIFIER, "x")
}

// --- Token.String coverage for LookupIdent ---

func TestLookupIdentNonKeyword(t *testing.T) {
	result := token.LookupIdent("myVar")
	if result != token.IDENTIFIER {
		t.Errorf("expected IDENTIFIER, got %s", result)
	}
}

func TestLookupIdentKeyword(t *testing.T) {
	result := token.LookupIdent("this")
	if result != token.KEYWORD {
		t.Errorf("expected KEYWORD, got %s", result)
	}
}

func TestLookupIdentBoolean(t *testing.T) {
	result := token.LookupIdent("true")
	if result != token.BOOLEAN {
		t.Errorf("expected BOOLEAN, got %s", result)
	}
}

// --- Sign gluing: minus as identifier when not followed by digit ---

func TestMinusAsIdentifier(t *testing.T) {
	// At start of file, - not followed by digit = identifier
	tokens := lexAll("- x")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "-")
	assertToken(t, tokens, 1, token.IDENTIFIER, "x")
}

func TestPlusAsIdentifier(t *testing.T) {
	tokens := lexAll("+ x")
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "+")
	assertToken(t, tokens, 1, token.IDENTIFIER, "x")
}

// --- Additional Coverage Tests ---

func TestHexValLowerCase(t *testing.T) {
	tokens := lexAll(`"\u0061"`) // 'a'
	assertTokenCount(t, tokens, 2)
	assertToken(t, tokens, 0, token.STRING, "a")
}

func TestBlockCommentPartial(t *testing.T) {
	// ## at EOF - treated as line comment #
	tokens := lexAll("##")
	assertTokenCount(t, tokens, 1)
	assertToken(t, tokens, 0, token.EOF, "")
}

func TestUnicodeNumberIdentifier(t *testing.T) {
	// Roman Numeral One Ⅰ (U+2160) is Number, Letter
	// ½ (U+00BD) is Number, Other
	input := "Ⅰ ½"
	tokens := lexAll(input)
	assertTokenCount(t, tokens, 3)
	assertToken(t, tokens, 0, token.IDENTIFIER, "Ⅰ")
	assertToken(t, tokens, 1, token.IDENTIFIER, "½")
}
