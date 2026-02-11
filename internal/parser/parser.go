package parser

import (
	"fmt"
	"orglang/internal/ast"
	"orglang/pkg/lexer"
	"orglang/pkg/token"
)

// Precedence levels
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

// BindingPower defines the left and right binding power for an operator
// Exported so users or tools can introspect it
func PrecedenceLevels() map[string]int {
	return map[string]int{
		"LOWEST":      LOWEST,
		"EQUALS":      EQUALS,
		"LESSGREATER": LESSGREATER,
		"SUM":         SUM,
		"PRODUCT":     PRODUCT,
		"PREFIX":      PREFIX,
		"CALL":        CALL,
		"INDEX":       INDEX,
	}
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l  *lexer.Lexer
	cl *lexer.CustomLexer // Using CustomLexer directly if needed, or wrap interface

	curToken  token.Token
	peekToken token.Token
	errors    []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn

	bindingPowers map[token.TokenType]OperatorPower // Cache
}

func New(l *lexer.CustomLexer) *Parser {
	p := &Parser{
		cl:            l,
		errors:        []string{},
		bindingPowers: OperatorBindingPowers(),
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseDecimalLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.AT, p.parsePrefixExpression) // @sys
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACE, p.parseBlockLiteral)
	p.registerPrefix(token.LBRACKET, p.parseListLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LBRACE, p.parseBlockLiteralInfix)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.cl.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	// OrgLang statement is Expression + Semicolon usually.
	// Or assignments which are also Expressions?
	// "Assignments are table insertions... which return values"
	// So essentially everything is an ExpressionStatement.
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	// Store as string for arbitrary precision
	return &ast.IntegerLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseDecimalLiteral() ast.Expression {
	// Store as string
	return &ast.DecimalLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// Blocks: (INTEGER)? "{" Expression "}" (INTEGER)?
func (p *Parser) parseBlockLiteral() ast.Expression {
	// This is called when curToken is '{'.
	// If there was an integer before it, it would have been parsed as an IntegerLiteral?
	// Wait. `700 { ... }`.
	// Parser sees `700`. Calls `parseIntegerLiteral`. Returns `IntegerLiteral`.
	// Then loop checks `infixParseFns`. `{` is LBRACE. Is LBRACE an infix operator?
	// Usually no.
	// So `parseExpression(LOWEST)` returns `IntegerLiteral(700)`.
	// Then `parseStatement` sees `IntegerLiteral`. Then `peekToken` is `{`.
	// It's not semicolon.

	// We need to handle `Integer LBRACE` sequence.
	// Ideally, LBRACE acts as an Infix operator for `Integer`?
	// Or we handle it at statement level?
	// "An integer just before the { ... Optional"
	// "Spaces not allowed between integer and brace" -> This is a Lexer rule or Parser check?
	// If lexer emits `INT(700)` then `LBRACE({)`, there might have been space.
	// Parser operates on tokens. It can check `p.curToken.EndPos == p.peekToken.StartPos` if we tracked positions?
	// Or we assume if syntax allows `700{`, it parses as Block with BP 700.

	// If we want `700 {` to be Block, then `{` must be an Infix operator binding to the integer?
	// `token.LBRACE` as infix?
	// If so, `700 { ... }` -> InfixExpression(Left=700, Op="{", Right=Body) ?
	// But BlockLiteral is a distinct node.

	// Alternative: `parseIntegerLiteral` looks ahead.
	// If next is `{`, it consumes it and parses Block.
	// But that breaks Pratt modularity.

	// Alternative: Register `{` as Infix.
	// `p.registerInfix(token.LBRACE, p.parseBlockLiteralInfix)`
	// And also `p.registerPrefix(token.LBRACE, p.parseBlockLiteralPrefix)` for `{ ... }` without BP.

	// Let's implement this strategy.

	block := &ast.BlockLiteral{Token: p.curToken}
	// This function is for Prefix `{`. So no Left BP.

	p.nextToken() // eat {
	block.Body = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	// Check for trailing BP: `} 701`
	// Again, "Spaces not allowed"?
	// If next token is INT, is it suffix BP or next statement?
	// ` { ... } 701; ` vs ` { ... }; 701; `
	// If no semicolon, `701` could be next expression.
	// But `Block` ends the expression usually?
	// If we are inside `parseExpression`, `}` returns.
	// The expression for Block is done.
	// Loop in `parseExpression` checks infix.
	// `INT` is not infix.

	// So `parseBlockLiteral` must consume the suffix integer if present?
	// Yes.

	if p.peekTokenIs(token.INT) {
		// consumes integer as RightBP.
		// We should strictly check for no-space if we could, but let's allow loosely for now or check offsets.
		p.nextToken()
		block.RightBP = p.curToken.Literal
	}

	return block
}

// Infix handler for `700 { ... }`
func (p *Parser) parseBlockLiteralInfix(left ast.Expression) ast.Expression {
	// Left is the IntegerLiteral
	// Ensure Left is actually an IntegerLiteral
	intLit, ok := left.(*ast.IntegerLiteral)
	if !ok {
		// Error or fallback?
		// `{` applied to non-integer? Function call? `func { ... }` (Ruby style?)
		// Spec says `INTEGER? {`.
		// If left is not integer, maybe it is a function call with lambda?
		// `user { x + 1 }`.
		// If OrgLang supports `func { block }`, then `{` is infix call?
		// For now, assume binding power syntax is only for Integers.
		// But what if `f { ... }`?
		p.errors = append(p.errors, fmt.Sprintf("Expected integer for block binding power, got %T", left))
		return nil
	}

	block := &ast.BlockLiteral{Token: p.curToken} // Token is {
	block.LeftBP = intLit.Value

	p.nextToken() // eat {
	block.Body = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	// Suffix BP
	if p.peekTokenIs(token.INT) {
		p.nextToken()
		block.RightBP = p.curToken.Literal
	}

	return block
}

func (p *Parser) parseListLiteral() ast.Expression {
	lit := &ast.ListLiteral{Token: p.curToken}

	// [ ]
	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return lit
	}

	p.nextToken() // valid start of expression

	// Parse list content as a single expression (if comma is operator, it handles list-ness)
	lit.Content = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return lit
}

// Helpers

func (p *Parser) curPrecedence() int {
	if power, ok := p.bindingPowers[p.curToken.Type]; ok {
		if power.Infix != nil {
			return power.Infix.Left
		}
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if power, ok := p.bindingPowers[p.peekToken.Type]; ok {
		if power.Infix != nil {
			return power.Infix.Left
		}
	}
	return LOWEST
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
