// Package lexer implements the OrgLang lexical analyzer.
//
// It reads UTF-8 source text and produces a stream of tokens as defined
// in the token package. The lexer handles sign gluing, rational literal
// detection, string escape sequences, raw strings, docstrings, Unicode
// identifiers, and compound structural operators.
package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"orglang/pkg/token"
)

// Lexer holds the state for scanning a single source input.
type Lexer struct {
	input         []byte
	pos           int             // current byte position
	line          int             // current line (1-indexed)
	col           int             // current column (1-indexed)
	prevTokenType token.TokenType // type of the last emitted token (for sign gluing)
}

// New creates a new Lexer for the given input bytes.
func New(input []byte) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

// Tokenize returns all tokens from the input, including the final EOF.
func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}

// NextToken scans and returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.input) {
		return l.makeToken(token.EOF, "")
	}

	r, _ := l.peekRune()
	startLine := l.line
	startCol := l.col

	var tok token.Token

	switch {
	// Structural delimiters
	case r == '(':
		l.readRune()
		tok = token.Token{Type: token.LPAREN, Literal: "(", Line: startLine, Column: startCol}
	case r == ')':
		l.readRune()
		tok = token.Token{Type: token.RPAREN, Literal: ")", Line: startLine, Column: startCol}
	case r == '[':
		l.readRune()
		tok = token.Token{Type: token.LBRACKET, Literal: "[", Line: startLine, Column: startCol}
	case r == ']':
		l.readRune()
		tok = token.Token{Type: token.RBRACKET, Literal: "]", Line: startLine, Column: startCol}
	case r == '{':
		l.readRune()
		tok = token.Token{Type: token.LBRACE, Literal: "{", Line: startLine, Column: startCol}
	case r == '}':
		l.readRune()
		tok = token.Token{Type: token.RBRACE, Literal: "}", Line: startLine, Column: startCol}
	case r == ';':
		l.readRune()
		tok = token.Token{Type: token.SEMICOLON, Literal: ";", Line: startLine, Column: startCol}

	// Structural operators
	case r == '@':
		tok = l.readAt(startLine, startCol)
	case r == ':':
		tok = l.readColon(startLine, startCol)
	case r == '.':
		l.readRune()
		tok = token.Token{Type: token.DOT, Literal: ".", Line: startLine, Column: startCol}
	case r == ',':
		l.readRune()
		tok = token.Token{Type: token.COMMA, Literal: ",", Line: startLine, Column: startCol}

	// Strings
	case r == '"':
		tok = l.readString(startLine, startCol)
	case r == '\'':
		tok = l.readRawString(startLine, startCol)

	// Numbers or sign-glued numbers
	case isASCIIDigit(r):
		tok = l.readNumber(0, startLine, startCol)
	case (r == '+' || r == '-') && l.shouldGlueSign():
		// Check if next char (after sign) is a digit
		if l.peekDigitAfterSign() {
			sign, _ := l.readRune()
			tok = l.readNumber(sign, startLine, startCol)
		} else {
			tok = l.readIdentifier(startLine, startCol)
		}

	// Identifiers (including operators like ->, ++, <=, etc.)
	case l.isIdentStart(r):
		tok = l.readIdentifier(startLine, startCol)

	default:
		// Check if it's a Unicode identifier start
		if l.isIdentStart(r) {
			tok = l.readIdentifier(startLine, startCol)
		} else {
			ch, _ := l.readRune()
			tok = token.Token{Type: token.ILLEGAL, Literal: string(ch), Line: startLine, Column: startCol}
		}
	}

	l.prevTokenType = tok.Type
	return tok
}

// --- Rune reading ---

func (l *Lexer) readRune() (rune, int) {
	if l.pos >= len(l.input) {
		return 0, 0
	}
	r, size := utf8.DecodeRune(l.input[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r, size
}

func (l *Lexer) peekRune() (rune, int) {
	if l.pos >= len(l.input) {
		return 0, 0
	}
	return utf8.DecodeRune(l.input[l.pos:])
}

func (l *Lexer) peekRuneAt(offset int) (rune, int) {
	p := l.pos + offset
	if p >= len(l.input) {
		return 0, 0
	}
	return utf8.DecodeRune(l.input[p:])
}

// --- Whitespace and comments ---

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			l.readRune()
			continue
		}
		if r == '#' {
			if l.isBlockComment() {
				l.skipBlockComment()
			} else {
				l.skipLineComment()
			}
			continue
		}
		break
	}
}

func (l *Lexer) skipLineComment() {
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '\n' {
			break
		}
	}
}

func (l *Lexer) isBlockComment() bool {
	// ### at column 1
	if l.col != 1 {
		return false
	}
	if l.pos+2 >= len(l.input) {
		return false
	}
	return l.input[l.pos] == '#' && l.input[l.pos+1] == '#' && l.input[l.pos+2] == '#'
}

func (l *Lexer) skipBlockComment() {
	// Consume the opening ###
	l.readRune() // #
	l.readRune() // #
	l.readRune() // #
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '\n' {
			// Check if next line starts with ###
			if l.pos+2 < len(l.input) &&
				l.input[l.pos] == '#' && l.input[l.pos+1] == '#' && l.input[l.pos+2] == '#' {
				l.readRune() // #
				l.readRune() // #
				l.readRune() // #
				return
			}
		}
	}
}

// --- Number scanning ---

func (l *Lexer) peekDigitAfterSign() bool {
	// The sign is at l.pos. Check l.pos+1 for a digit.
	r, _ := l.peekRuneAt(1)
	return isASCIIDigit(r)
}

// shouldGlueSign returns true if a +/- should be treated as a sign
// (prefix position: after operator, delimiter, or at start of input).
func (l *Lexer) shouldGlueSign() bool {
	switch l.prevTokenType {
	case "": // start of file
		return true
	case token.LPAREN, token.LBRACKET, token.LBRACE,
		token.SEMICOLON, token.COMMA,
		token.AT, token.AT_COLON, token.COLON, token.DOT,
		token.ELVIS:
		return true
	case token.IDENTIFIER, token.KEYWORD:
		// Identifiers that are operators would mean prefix position.
		// But we can't easily distinguish — play it safe: identifiers generally
		// return values, so +/- after them is infix. However, if the identifier
		// is a known operator (like `+`, `-`, `->`, etc.), it's prefix position.
		// For simplicity: treat identifier as NOT prefix position.
		return false
	default:
		return false
	}
}

func (l *Lexer) readNumber(sign rune, startLine, startCol int) token.Token {
	var buf strings.Builder
	if sign != 0 {
		buf.WriteRune(sign)
	}

	// Read integer digits
	l.readDigits(&buf)

	// Check for decimal point: digit.digit
	if l.pos < len(l.input) {
		r, rSize := l.peekRune()
		if r == '.' {
			// Peek ahead: if next after dot is a digit, it's a decimal
			r2, _ := l.peekRuneAt(rSize)
			if isASCIIDigit(r2) {
				l.readRune() // consume '.'
				buf.WriteRune('.')
				l.readDigits(&buf)
				return token.Token{Type: token.DECIMAL, Literal: buf.String(), Line: startLine, Column: startCol}
			}
			// Otherwise: 1. -> INTEGER + DOT (dot stays for next token)
		}
	}

	// Check for rational: integer/integer (no whitespace)
	if l.pos < len(l.input) {
		r, rSize := l.peekRune()
		if r == '/' {
			r2, _ := l.peekRuneAt(rSize)
			if isASCIIDigit(r2) || r2 == '+' || r2 == '-' {
				l.readRune() // consume '/'
				buf.WriteRune('/')
				// Optional sign on denominator
				r3, _ := l.peekRune()
				if r3 == '+' || r3 == '-' {
					s, _ := l.readRune()
					buf.WriteRune(s)
				}
				l.readDigits(&buf)
				return token.Token{Type: token.RATIONAL, Literal: buf.String(), Line: startLine, Column: startCol}
			}
		}
	}

	return token.Token{Type: token.INTEGER, Literal: buf.String(), Line: startLine, Column: startCol}
}

func (l *Lexer) readDigits(buf *strings.Builder) {
	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if !isASCIIDigit(r) {
			break
		}
		l.readRune()
		buf.WriteRune(r)
	}
}

// --- String scanning ---

func (l *Lexer) readString(startLine, startCol int) token.Token {
	l.readRune() // consume opening "

	// Check for docstring """
	if l.matchString("\"\"") {
		return l.readDocstring(startLine, startCol)
	}

	var buf strings.Builder
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '"' {
			return token.Token{Type: token.STRING, Literal: buf.String(), Line: startLine, Column: startCol}
		}
		if r == '\\' {
			escaped, err := l.readEscape()
			if err != "" {
				return token.Token{Type: token.ILLEGAL, Literal: err, Line: startLine, Column: startCol}
			}
			buf.WriteRune(escaped)
			continue
		}
		buf.WriteRune(r)
	}

	// Unterminated string
	return token.Token{Type: token.ILLEGAL, Literal: "unterminated string", Line: startLine, Column: startCol}
}

func (l *Lexer) readDocstring(startLine, startCol int) token.Token {
	// Opening """ already consumed (first " by readString, next "" by matchString)
	var buf strings.Builder
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '"' && l.matchString("\"\"") {
			content := stripDocIndent(buf.String())
			return token.Token{Type: token.DOCSTRING, Literal: content, Line: startLine, Column: startCol}
		}
		if r == '\\' {
			escaped, err := l.readEscape()
			if err != "" {
				return token.Token{Type: token.ILLEGAL, Literal: err, Line: startLine, Column: startCol}
			}
			buf.WriteRune(escaped)
			continue
		}
		buf.WriteRune(r)
	}
	return token.Token{Type: token.ILLEGAL, Literal: "unterminated docstring", Line: startLine, Column: startCol}
}

func (l *Lexer) readRawString(startLine, startCol int) token.Token {
	l.readRune() // consume opening '

	// Check for raw docstring '''
	if l.matchString("''") {
		return l.readRawDocstring(startLine, startCol)
	}

	var buf strings.Builder
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '\'' {
			return token.Token{Type: token.RAWSTRING, Literal: buf.String(), Line: startLine, Column: startCol}
		}
		buf.WriteRune(r)
	}
	return token.Token{Type: token.ILLEGAL, Literal: "unterminated raw string", Line: startLine, Column: startCol}
}

func (l *Lexer) readRawDocstring(startLine, startCol int) token.Token {
	// Opening ''' already consumed
	var buf strings.Builder
	for l.pos < len(l.input) {
		r, _ := l.readRune()
		if r == '\'' && l.matchString("''") {
			content := stripDocIndent(buf.String())
			return token.Token{Type: token.RAWDOC, Literal: content, Line: startLine, Column: startCol}
		}
		buf.WriteRune(r)
	}
	return token.Token{Type: token.ILLEGAL, Literal: "unterminated raw docstring", Line: startLine, Column: startCol}
}

// matchString checks if the next bytes match s, and if so, consumes them.
func (l *Lexer) matchString(s string) bool {
	if l.pos+len(s) > len(l.input) {
		return false
	}
	if string(l.input[l.pos:l.pos+len(s)]) == s {
		// Advance through each rune for proper line/col tracking
		for range len(s) {
			l.readRune()
		}
		return true
	}
	return false
}

// --- Escape sequences ---

func (l *Lexer) readEscape() (rune, string) {
	if l.pos >= len(l.input) {
		return 0, "unterminated escape sequence"
	}
	r, _ := l.readRune()
	switch r {
	case 'n':
		return '\n', ""
	case 't':
		return '\t', ""
	case 'r':
		return '\r', ""
	case '\\':
		return '\\', ""
	case '"':
		return '"', ""
	case '0':
		return 0, ""
	case 'u':
		return l.readUnicodeEscape()
	default:
		return 0, fmt.Sprintf("unknown escape: \\%c", r)
	}
}

func (l *Lexer) readUnicodeEscape() (rune, string) {
	if l.pos >= len(l.input) {
		return 0, "unterminated unicode escape"
	}
	r, _ := l.peekRune()
	if r == '{' {
		// \u{XXXXXX} — 1 to 6 hex digits
		l.readRune() // consume '{'
		var val rune
		count := 0
		for l.pos < len(l.input) {
			r, _ := l.peekRune()
			if r == '}' {
				l.readRune()
				if count == 0 {
					return 0, "empty unicode escape \\u{}"
				}
				if val > 0x10FFFF {
					return 0, fmt.Sprintf("unicode codepoint out of range: U+%X", val)
				}
				return val, ""
			}
			d := hexVal(r)
			if d < 0 {
				return 0, fmt.Sprintf("invalid hex digit in unicode escape: %c", r)
			}
			l.readRune()
			val = val*16 + rune(d)
			count++
			if count > 6 {
				return 0, "unicode escape too long (max 6 hex digits)"
			}
		}
		return 0, "unterminated unicode escape \\u{...}"
	}

	// \uXXXX — exactly 4 hex digits
	var val rune
	for i := range 4 {
		if l.pos >= len(l.input) {
			return 0, "unterminated unicode escape \\uXXXX"
		}
		r, _ := l.readRune()
		d := hexVal(r)
		if d < 0 {
			return 0, fmt.Sprintf("invalid hex digit in unicode escape at position %d: %c", i+1, r)
		}
		val = val*16 + rune(d)
	}
	return val, ""
}

// --- Identifier scanning ---

func (l *Lexer) readIdentifier(startLine, startCol int) token.Token {
	var buf strings.Builder

	for l.pos < len(l.input) {
		r, _ := l.peekRune()
		if isASCIIDigit(r) || l.isIdentContinue(r) {
			l.readRune()
			buf.WriteRune(r)
		} else {
			break
		}
	}

	lit := buf.String()

	// Check for ?:  (ELVIS)
	if lit == "?" {
		r, _ := l.peekRune()
		if r == ':' {
			l.readRune()
			return token.Token{Type: token.ELVIS, Literal: "?:", Line: startLine, Column: startCol}
		}
	}

	// Look up keyword/boolean
	tokType := token.LookupIdent(lit)
	return token.Token{Type: tokType, Literal: lit, Line: startLine, Column: startCol}
}

// --- Structural operator helpers ---

func (l *Lexer) readAt(startLine, startCol int) token.Token {
	l.readRune() // consume '@'
	r, _ := l.peekRune()
	if r == ':' {
		l.readRune()
		return token.Token{Type: token.AT_COLON, Literal: "@:", Line: startLine, Column: startCol}
	}
	return token.Token{Type: token.AT, Literal: "@", Line: startLine, Column: startCol}
}

func (l *Lexer) readColon(startLine, startCol int) token.Token {
	l.readRune() // consume ':'
	// Check for extended assignment operators (reserved)
	if l.pos < len(l.input) {
		r, _ := l.peekRune()
		switch r {
		case '+', '-', '*', '/', '%', '&', '^', '|', '~':
			l.readRune()
			return token.Token{Type: token.IDENTIFIER, Literal: ":" + string(r), Line: startLine, Column: startCol}
		case '>':
			// Check for :>>
			l.readRune()
			r2, _ := l.peekRune()
			if r2 == '>' {
				l.readRune()
				return token.Token{Type: token.IDENTIFIER, Literal: ":>>", Line: startLine, Column: startCol}
			}
			// Just :> (not a defined operator, but emit as identifier)
			return token.Token{Type: token.IDENTIFIER, Literal: ":>", Line: startLine, Column: startCol}
		case '<':
			// Check for :<<
			l.readRune()
			r2, _ := l.peekRune()
			if r2 == '<' {
				l.readRune()
				return token.Token{Type: token.IDENTIFIER, Literal: ":<<", Line: startLine, Column: startCol}
			}
			return token.Token{Type: token.IDENTIFIER, Literal: ":<", Line: startLine, Column: startCol}
		}
	}
	return token.Token{Type: token.COLON, Literal: ":", Line: startLine, Column: startCol}
}

// --- Docstring indent stripping ---

// stripDocIndent removes the common leading whitespace from a docstring.
// It strips the leading newline and trailing newline if present, then
// finds the minimum indentation across non-empty lines and removes it.
func stripDocIndent(s string) string {
	// Strip leading newline
	if len(s) > 0 && s[0] == '\n' {
		s = s[1:]
	}
	// Strip trailing newline
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}

	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Find minimum indentation across non-empty lines
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.Join(lines, "\n")
	}

	// Strip common indent
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}

// --- Character classification ---

func (l *Lexer) isIdentStart(r rune) bool {
	if r == '_' {
		return true
	}
	if isASCIIDigit(r) {
		return false
	}
	if isStructural(r) {
		return false
	}
	// Explicitly allow ASCII operators that are otherwise punctuation
	// +, -, *, /, %, ?, !, &, |, ^, ~, <, >, =, $
	switch r {
	case '+', '-', '*', '/', '%', '?', '!', '&', '|', '^', '~', '<', '>', '=', '$':
		return true
	}
	// \p{Letter}, \p{Symbol}, or \p{Number} but not \p{Punctuation}
	return unicode.IsLetter(r) || unicode.IsSymbol(r) || unicode.IsNumber(r)
}

func (l *Lexer) isIdentContinue(r rune) bool {
	return l.isIdentStart(r)
}

func isStructural(r rune) bool {
	switch r {
	case '@', ':', '.', ',', ';',
		'(', ')', '[', ']', '{', '}',
		'"', '\'', '\\', '#':
		return true
	}
	// Explicitly exclude ASCII operators from being structural (they are identifiers)
	switch r {
	case '+', '-', '*', '/', '%', '?', '!', '&', '|', '^', '~', '<', '>', '=', '$':
		return false
	}
	// Any OTHER \p{Punctuation} character is also structural (excluded from identifiers)
	return unicode.IsPunct(r)
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func hexVal(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'a' && r <= 'f':
		return int(r-'a') + 10
	case r >= 'A' && r <= 'F':
		return int(r-'A') + 10
	default:
		return -1
	}
}

// --- Token construction helper ---

func (l *Lexer) makeToken(tt token.TokenType, lit string) token.Token {
	return token.Token{Type: tt, Literal: lit, Line: l.line, Column: l.col}
}
