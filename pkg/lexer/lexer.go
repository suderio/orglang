package lexer

import (
	"orglang/pkg/token"
	"strings"
	"text/scanner"
)

type Lexer struct {
	scanner scanner.Scanner
	src     string
	file    string
}

func New(file string, input string) *Lexer {
	l := &Lexer{src: input, file: file}
	// Initialize the text/scanner
	l.scanner.Init(strings.NewReader(input))
	l.scanner.Filename = file

	// Configure scanner mode
	// We want to handle identifiers, numbers, and strings mostly by ourselves to match spec
	// But text/scanner is convenient. Let's start with basic mode and customize.
	// We will likely need ScanInts, ScanFloats, ScanStrings.
	// However, OrgLang identifiers allow symbols. This breaks default ScanIdents.
	// So we might need to rely on ScanChars mostly.

	// Let's try raw ScanChars logic for identifiers to support `+`, `-`, etc.
	l.scanner.Mode = scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanComments | scanner.SkipComments

	// Customize IsIdentRune? text/scanner doesn't expose this easily in public API without overriding.
	// Actually, text/scanner is quite rigid about what constitutes an identifier.
	// Given our `IDENTIFIER` rule: [a-zA-Z_!$%&*+\-=\^~?/<|>][a-zA-Z0-9_!$%&*+\-=\^~?/<|>]*
	// logic, `text/scanner`'s default `ScanIdents` will fail us.
	// We might acturally be better off writing a custom loop.

	// Re-evaluating strategy: Writing a simple custom lexer is often easier than fighting `text/scanner`.
	// Let's modify this file to use a custom simple lexer derived from the "Writing an Interpreter in Go" book style,
	// which fits the Pratt Parser pattern perfectly and gives us full control.

	return l
}

// TODO: Replace text/scanner with custom logic in next iteration if needed.
// For now, let's implement NextToken using a custom simple approach since text/scanner is restrictive.

type CustomLexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int
}

func NewCustom(input string) *CustomLexer {
	l := &CustomLexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *CustomLexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
	l.column += 1
}

func (l *CustomLexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *CustomLexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	if l.ch == 0 {
		tok.Literal = ""
		tok.Type = token.EOF
		return tok
	}

	// Handle structural tokens first
	switch l.ch {
	case '(':
		tok.Type = token.LPAREN
		tok.Literal = "("
	case ')':
		tok.Type = token.RPAREN
		tok.Literal = ")"
	case '[':
		tok.Type = token.LBRACKET
		tok.Literal = "["
	case ']':
		tok.Type = token.RBRACKET
		tok.Literal = "]"
	case '{':
		tok.Type = token.LBRACE
		tok.Literal = "{"
	case '}':
		tok.Type = token.RBRACE
		tok.Literal = "}"
	case ',':
		tok.Type = token.COMMA
		tok.Literal = ","
	case ';':
		tok.Type = token.SEMICOLON
		tok.Literal = ";"
	case ':':
		tok.Type = token.COLON
		tok.Literal = ":"
	case '.':
		tok.Type = token.DOT
		tok.Literal = "."
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '@':
		tok.Type = token.AT
		tok.Literal = "@"
	case '#':
		// Comment handling
		if l.peekChar() == '#' && l.peekCharAhead(2) == '#' {
			// Block comment ### ... ###
			l.readChar() // #
			l.readChar() // #
			// consume until ###
			for {
				if l.ch == '#' && l.peekChar() == '#' && l.peekCharAhead(2) == '#' {
					l.readChar() // #
					l.readChar() // #
					l.readChar() // #
					break
				}
				if l.ch == 0 {
					break
				}
				l.readChar()
			}
		} else {
			// Single line comment
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
		}
		return l.NextToken()

	// Default case: Check for identifiers/symbols
	default:
		if isLetter(l.ch) || isSymbolChar(l.ch) {
			literal := l.readIdentifier()
			// Refine type
			tok.Literal = literal

			// Special case for ELVIS '?:' which mixes symbol '?' and structural ':'
			if literal == "?" && l.ch == ':' {
				l.readChar() // eat ':'
				tok.Type = token.ELVIS
				tok.Literal = "?:"
				return tok
			}

			// Special case for Signed Numbers (e.g., -1, +2.5) if they start with + or - and followed by digit
			// But wait, + and - are operators too.
			// The rule says: "unary negation (-) ... must also use spaces".
			// "Integer and Decimal definitions ... don't allow for a minus/plus sign in the beginning (this sign *must* not have a space between it and the number)."
			// So: `- 1` is Unary Minus, Number 1.
			// `-1` is Number -1.
			// The lexer needs to distinguish.
			// If literal is "+" or "-" and next char is digit, it's a number?
			// But what about `a-1`? Id `a`, operator `-`, number `1`.
			// Our readIdentifier reads symbols too. `a-1` is identifier `a-1` currently!
			// Wait, `IDENTIFIER` rule allows symbols inside.
			// If the user types `-1`, `readIdentifier` consumes `-` then... checks loop.
			// `isDigit` is true. `readIdentifier` loop continues!
			// So `-1` is parsed as IDENTIFIER "-1".
			// We need to check if an identifier looks like a number.

			// Actually, we should check for number start BEFORE identifier if it starts with digit.
			// But valid identifiers can't start with digit.
			// Signed numbers start with + or -.

			// If we are here, `literal` contains the consumed identifier/symbol sequence.
			// Check if it is a signed number.
			// Regex for signed number: ^[+-][0-9]+(\.[0-9]+)?$
			// But `readIdentifier` consumes symbols/letters/digits greedily.
			// If input is `-1`, literal is `-1`.
			// If input is `-1a`, literal is `-1a` (Identifier).
			// So we can check if `literal` matches number pattern.

			if isSignedNumber(literal) {
				// If it looks like a signed number, check for decimal part which readIdentifier stops at
				if l.ch == '.' && isDigit(l.peekChar()) {
					l.readChar() // eat dot
					literal += "." + l.readNumber()
					tok.Literal = literal
					tok.Type = token.FLOAT
				} else if strings.Contains(literal, ".") {
					// Already contains dot? (readIdentifier stops at dot, so maybe not unless logic changes)
					tok.Type = token.FLOAT
				} else {
					tok.Type = token.INT
				}
				return tok
			}

			switch literal {
			case "?":
				tok.Type = token.QUESTION
			case "!":
				tok.Type = token.LNOT
			case "~":
				tok.Type = token.NOT
			case "->":
				tok.Type = token.ARROW
			case "-<":
				tok.Type = token.DARROW
			case "-<>":
				tok.Type = token.JOIN_ARROW
			case "??":
				tok.Type = token.ERROR_CHECK
			case "?:":
				tok.Type = token.ELVIS
			case "|>":
				tok.Type = token.PIPE
			case "o":
				tok.Type = token.COMPOSE
			case "++":
				tok.Type = token.PLUS_PLUS
			case "--":
				tok.Type = token.MINUS_MINUS
			case "+":
				tok.Type = token.PLUS
			case "-":
				tok.Type = token.MINUS
			case "*":
				tok.Type = token.ASTERISK
			case "/":
				tok.Type = token.SLASH
			case "%":
				tok.Type = token.MODULO
			case "**":
				tok.Type = token.POWER
			case "=":
				tok.Type = token.EQ
			case "<>":
				tok.Type = token.NOT_EQ
			case "<":
				tok.Type = token.LT
			case ">":
				tok.Type = token.GT
			case "<=":
				tok.Type = token.LT_EQ
			case ">=":
				tok.Type = token.GT_EQ
			case "&&":
				tok.Type = token.AND
			case "||":
				tok.Type = token.OR
			case "&":
				tok.Type = token.BIT_AND
			case "|":
				tok.Type = token.BIT_OR
			case "^":
				tok.Type = token.BIT_XOR
			case "<<":
				tok.Type = token.LSHIFT
			case ">>":
				tok.Type = token.RSHIFT
			default:
				tok.Type = token.LookupIdent(literal)
			}
			return tok
		} else if isDigit(l.ch) {
			// Unsigned number starting with digit
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			// Check for float
			// readNumber stopped at non-digit. If it is dot, check for float.
			if l.ch == '.' {
				// Must check if next char after dot is digit
				if isDigit(l.peekChar()) {
					l.readChar() // eat dot
					tok.Literal += "." + l.readNumber()
					tok.Type = token.FLOAT
				}
			}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *CustomLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

func (l *CustomLexer) readIdentifier() string {
	position := l.position
	// Note: This reading loop treats symbols and letters as part of the same identifier
	// This allows "foo-bar" or "sign!" or "<=>"
	for isLetter(l.ch) || isDigit(l.ch) || isSymbolChar(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *CustomLexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *CustomLexer) readString() string {
	if l.peekChar() == '"' && l.peekCharAhead(2) == '"' {
		// Multiline string """
		l.readChar() // eat second "
		l.readChar() // eat third "
		l.readChar() // consume possible newline after opening quotes

		startPos := l.position
		for {
			if l.ch == '"' && l.peekChar() == '"' && l.peekCharAhead(2) == '"' {
				break
			}
			if l.ch == 0 {
				break
			}
			l.readChar()
		}

		raw := l.input[startPos:l.position]

		// Consume closing quotes
		l.readChar()
		l.readChar()
		l.readChar()

		return stripIndentation(raw)
	}

	// Single line string
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	// Check if we stopped at a quote and consume it
	content := l.input[position:l.position]
	if l.ch == '"' {
		l.readChar() // Consume the closing quote
	}
	return content
}

func stripIndentation(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return ""
	}

	// Find minimum common indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for i := 0; i < len(line) && (line[i] == ' ' || line[i] == '\t'); i++ {
			indent++
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.TrimSuffix(s, "\n")
	}

	var result strings.Builder
	for i, line := range lines {
		if len(line) >= minIndent {
			result.WriteString(line[minIndent:])
		} else {
			result.WriteString(strings.TrimSpace(line))
		}
		if i < len(lines)-1 {
			result.WriteByte('\n')
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

func (l *CustomLexer) peekCharAhead(n int) byte {
	if l.readPosition+n-1 >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+n-1]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isSymbolChar(ch byte) bool {
	// ! $ % & * - + = ^ ~ ? / < > |
	// ! $ % & * - + = ^ ~ ? / < > |
	switch ch {
	// ! is frequently a prefix operator (!true), so we should be careful.
	// If ! is followed by a letter, it should probably be a separate operator token
	// unless we want identifiers like `risk!` (Ruby style).
	// But `!true` tokenizing as `!true` identifier is bad.
	// For now, let's KEEP ! as a symbol char, but in readIdentifier, we might need a special check.
	// Actually, `readIdentifier` is the problem.
	// If the identifier STARTs with `!`, usage is likely prefix op.
	// But OrgLang allows `!valid` as var name?
	// Reference says: "must start with a letter ... or any of the following symbols: !"
	// So `!variable` IS a valid identifier.
	// `!false` IS a valid identifier.
	// But `false` is a keyword.
	// If `!false` is parsed as identifier, it's not `! false`.
	// The user must write `! false` (with space) if they mean prefix NOT.
	// UNLESS `!` is a delimiter.
	// Reference says: "Operator Philosophy... Operators are strictly unary (prefix) or binary (infix)."
	// And "Identifiers ... must start with ... ! ...".
	// So `!false` IS an identifier.
	// The user used `!false` in the test.
	// They must use `! false` or `( ! false )`.
	// I will fix `basic.org` to use `! false`.
	case '!', '$', '%', '&', '*', '-', '+', '=', '^', '~', '?', '/', '<', '>', '|':
		return true
	}
	return false
}
