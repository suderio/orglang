package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + Literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	FLOAT  = "FLOAT"  // 3.14
	STRING = "STRING" // "foobar"

	// Operators and Delimiters
	// Structural
	ASSIGN    = ":"
	COLON     = ":" // Alias for ASSIGN, used in parser/lexer for structural context
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	DOT       = "."

	// Keywords / Reserved (if any, for now purely contextual)
	TRUE     = "true"
	FALSE    = "false"
	ERROR    = "Error"
	RESOURCE = "resource"
	THIS     = "this"
	LEFT     = "left"
	RIGHT    = "right"

	// Special Operators
	AT          = "@"
	ARROW       = "->"
	DARROW      = "-<"
	JOIN_ARROW  = "-<>"
	ELVIS       = "?:"
	ERROR_CHECK = "??"
	PIPE        = "|>"
	COMPOSE     = "o"
	BINDING_TAG = "BINDING_TAG" // !123

	// Conditional / Access
	QUESTION = "?"

	// Comparison
	EQ     = "="
	NOT_EQ = "<>"
	LT     = "<"
	GT     = ">"
	LT_EQ  = "<="
	GT_EQ  = ">="

	// Arithmetic
	PLUS        = "+"
	PLUS_PLUS   = "++"
	MINUS       = "-"
	MINUS_MINUS = "--"
	NOT         = "~" // Negation
	ASTERISK    = "*"
	SLASH       = "/"
	MODULO      = "%"
	POWER       = "**"

	// Logic
	AND     = "&&"
	OR      = "||"
	BIT_AND = "&"
	BIT_OR  = "|"
	BIT_XOR = "^"
	// NOT     = "!" // Removed !
)

var keywords = map[string]TokenType{
	"true":     TRUE,
	"false":    FALSE,
	"resource": RESOURCE,
	"this":     THIS,
	"left":     LEFT,
	"right":    RIGHT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
