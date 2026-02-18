// Package token defines the token types and structures used by the OrgLang lexer.
package token

// TokenType represents the type of a lexical token.
type TokenType string

const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Literals
	INTEGER    TokenType = "INTEGER"
	DECIMAL    TokenType = "DECIMAL"
	RATIONAL   TokenType = "RATIONAL"
	STRING     TokenType = "STRING"
	DOCSTRING  TokenType = "DOCSTRING"
	RAWSTRING  TokenType = "RAWSTRING"
	RAWDOC     TokenType = "RAWDOC"
	BOOLEAN    TokenType = "BOOLEAN"

	// Identifiers and keywords
	IDENTIFIER TokenType = "IDENTIFIER"
	KEYWORD    TokenType = "KEYWORD"

	// Structural delimiters
	LPAREN    TokenType = "LPAREN"    // (
	RPAREN    TokenType = "RPAREN"    // )
	LBRACKET  TokenType = "LBRACKET"  // [
	RBRACKET  TokenType = "RBRACKET"  // ]
	LBRACE    TokenType = "LBRACE"    // {
	RBRACE    TokenType = "RBRACE"    // }
	SEMICOLON TokenType = "SEMICOLON" // ;

	// Structural operators
	AT       TokenType = "AT"       // @
	AT_COLON TokenType = "AT_COLON" // @:
	COLON    TokenType = "COLON"    // :
	DOT      TokenType = "DOT"      // .
	COMMA    TokenType = "COMMA"    // ,

	// Compound structural operators
	ELVIS TokenType = "ELVIS" // ?:
)

// Token represents a single lexical token with its type, literal value,
// and source position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int // 1-indexed
	Column  int // 1-indexed
}

// keywords maps reserved words to their token type.
var keywords = map[string]TokenType{
	"true":  BOOLEAN,
	"false": BOOLEAN,
	"this":  KEYWORD,
	"left":  KEYWORD,
	"right": KEYWORD,
}

// LookupIdent checks if an identifier is a keyword or boolean literal.
// Returns the appropriate TokenType.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENTIFIER
}
