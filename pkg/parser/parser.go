package parser

import (
	"fmt"
	"strconv"
	"strings"

	"orglang/pkg/ast"
	"orglang/pkg/lexer"
	"orglang/pkg/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	prevToken token.Token // Track previous token for adjacency checks
	errors    []string
	bpTable   *BindingTable
	inTable   bool
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:       l,
		errors:  []string{},
		bpTable: NewBindingTable(),
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) peek() token.Token {
	return p.peekToken
}

func (p *Parser) cur() token.Token {
	return p.curToken
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d:%d: %s", p.curToken.Line, p.curToken.Column, msg))
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{
		Statements: []ast.Statement{},
	}

	for p.curToken.Type != token.EOF {
		if p.curToken.Type == token.SEMICOLON {
			p.nextToken()
			continue
		}

		stmt := p.parseExpression(0)
		if stmt != nil {
			if s, ok := stmt.(ast.Statement); ok {
				prog.Statements = append(prog.Statements, s)
			}
		}
	}
	return prog
}

const (
	LOWEST      = 0
	COMMA       = 60
	BINDING     = 80
	SUM         = 200
	PRODUCT     = 300
	COMPOSITION = 400
	EXPONENT    = 500
	PREFIX      = 900
)

func (p *Parser) getBindingPower(t token.Token) int {
	switch t.Type {
	case token.DOT, token.AT:
		return 900
	case token.ELVIS:
		return 750
	case token.AT_COLON, token.COLON:
		return 80
	case token.COMMA:
		return 60
	case token.EOF, token.SEMICOLON, token.RPAREN, token.RBRACE, token.RBRACKET:
		return 0
	case token.IDENTIFIER, token.KEYWORD, token.BOOLEAN:
		if p.inTable {
			return 0
		}
		if entry, ok := p.bpTable.Lookup(t.Literal); ok && entry.IsInfix {
			return entry.LBP
		}
		return 0
	}
	return 0
}

func (p *Parser) parseExpression(minBP int) ast.Expression {
	t := p.curToken
	p.nextToken() // Consume NUD

	left := p.nud(t)
	if left == nil {
		return &ast.ErrorExpr{Message: fmt.Sprintf("unexpected token %s (%q)", t.Type, t.Literal)}
	}

	for {
		lbp := p.getBindingPower(p.curToken)
		if lbp <= minBP {
			break
		}

		ledOp := p.curToken
		p.nextToken() // Consume Operator
		left = p.led(ledOp, left)
	}

	return left
}

func (p *Parser) nud(t token.Token) ast.Expression {
	switch t.Type {
	case token.INTEGER:
		if p.curToken.Type == token.LBRACE && p.areAdjacent(t, p.curToken) {
			lbpVal, _ := strconv.Atoi(t.Literal)
			return p.parseFunctionLiteral(&lbpVal)
		}
		return &ast.IntegerLiteral{Value: t.Literal}
	case token.DECIMAL:
		return &ast.DecimalLiteral{Value: t.Literal}
	case token.RATIONAL:
		parts := strings.Split(t.Literal, "/")
		if len(parts) != 2 {
			return &ast.RationalLiteral{Numerator: t.Literal, Denominator: "1"}
		}
		return &ast.RationalLiteral{Numerator: parts[0], Denominator: parts[1]}
	case token.STRING, token.DOCSTRING, token.RAWSTRING, token.RAWDOC:
		isDoc := t.Type == token.DOCSTRING || t.Type == token.RAWDOC
		isRaw := t.Type == token.RAWSTRING || t.Type == token.RAWDOC
		return &ast.StringLiteral{Value: t.Literal, IsDoc: isDoc, IsRaw: isRaw}
	case token.BOOLEAN:
		val := t.Literal == "true"
		return &ast.BooleanLiteral{Value: val}
	case token.IDENTIFIER, token.KEYWORD, token.AT:
		return p.nudIdentifier(t)
	case token.LPAREN:
		expr := p.parseExpression(0)
		if p.curToken.Type == token.RPAREN {
			p.nextToken()
		} else {
			p.addError("expected ')'")
		}
		return &ast.GroupExpr{Inner: expr}
	case token.LBRACE:
		return p.parseFunctionLiteral(nil)
	case token.LBRACKET:
		return p.parseTableLiteral()
	case token.ILLEGAL:
		return &ast.ErrorExpr{Message: t.Literal}
	}
	return nil
}

func (p *Parser) nudIdentifier(t token.Token) ast.Expression {
	name := t.Literal
	entry, ok := p.bpTable.Lookup(name)

	if ok && entry.IsPrefix {
		bp := entry.PrefixBP
		if bp == 0 {
			bp = PREFIX
		}
		right := p.parseExpression(bp)
		return &ast.PrefixExpr{Op: name, Right: right}
	}

	if !ok {
		// allow if defining
		if p.curToken.Type == token.COLON || p.curToken.Type == token.AT_COLON {
			return &ast.Name{Value: name}
		}
		if name == "left" || name == "right" || name == "this" {
			return &ast.Name{Value: name}
		}
		return &ast.ErrorExpr{Message: fmt.Sprintf("undefined identifier: %s", name)}
	}

	return &ast.Name{Value: name}
}

func (p *Parser) led(t token.Token, left ast.Expression) ast.Expression {
	switch t.Type {
	case token.COLON:
		return p.ledBinding(left, false)
	case token.AT_COLON:
		return p.ledBinding(left, true)
	case token.DOT:
		right := p.parseExpression(p.getBindingPower(t))
		return &ast.DotExpr{Left: left, Key: right}
	case token.ELVIS:
		right := p.parseExpression(750)
		return &ast.ElvisExpr{Left: left, Right: right}
	case token.COMMA:
		right := p.parseExpression(60)
		return &ast.CommaExpr{Left: left, Right: right}
	case token.IDENTIFIER:
		if t.Literal == "|>" {
			opExpr := &ast.InfixExpr{Left: left, Op: "|>", Right: p.parseAtom()}
			// If the next token starts an expression, treating the result as a user-defined operator implies
			// we should consume the argument with high precedence (like a prefix operator).
			// We use 900 (same as prefix operators in binding_powers.go).
			// We check if the next token has a NUD (is start of expression).
			// However, parseExpression handles this check. But we only want to consume if LBP allows?
			// No, prefix operators consume regardless of previous LBP context if they are in NUD position.
			// Here we are in LED position of |>. We finished parsing |>.
			// Now we are effectively in NUD position of the *result* of |>.
			// So we blindly try to parse an expression?
			// If we do, we might consume something that belongs to a higher level construct?
			// E.g. `a |> b`. Next is `;` or `)`. `parseExpression(900)` will fail or return nil?
			// parseExpression expects a NUD.
			// So we check if current token can start an expression.
			// Simple check: is it NOT a delimiter that ends an expression?
			// Better: check if NUD is defined? But NUD logic is inside parseExpression.
			// Let's iterate: if we can parse an expression with BP 900, we do.
			// But parseExpression errors if no NUD.
			// We can peek.
			if p.isPossibleExpressionStart() {
				// We must ensure we don't consume tokens that should bind weaker than 900?
				// Actually, parseExpression(900) will stop if it hits an infix with LBP <= 900.
				// But we need to know if we SHOULD start parsing.
				// If next is `;`, parseExpression errors.
				// If next is `)`, errors.
				// So we need a check.
				arg := p.parseExpression(900)
				return &ast.ApplyExpr{Func: opExpr, Arg: arg}
			}
			return opExpr
		}
		if t.Literal == "o" {
			opExpr := &ast.InfixExpr{Left: left, Op: "o", Right: p.parseAtom()}
			if p.isPossibleExpressionStart() {
				arg := p.parseExpression(900)
				return &ast.ApplyExpr{Func: opExpr, Arg: arg}
			}
			return opExpr
		}

		entry, _ := p.bpTable.Lookup(t.Literal)
		rbp := entry.RBP
		right := p.parseExpression(rbp)
		return &ast.InfixExpr{Left: left, Op: t.Literal, Right: right}

	case token.AT:
		bp := 900
		right := p.parseExpression(bp)
		return &ast.InfixExpr{Left: left, Op: "@", Right: right}
	}

	return left
}

func (p *Parser) ledBinding(left ast.Expression, isResource bool) ast.Expression {
	// Colon is Right-associative. RBP = 79.
	val := p.parseExpression(79)

	if name, ok := left.(*ast.Name); ok {
		if funcLit, ok := val.(*ast.FunctionLiteral); ok {
			p.registerBinding(name.Value, funcLit, isResource)
		} else {
			p.bpTable.RegisterValue(name.Value)
		}
	}

	if isResource {
		return &ast.ResourceDef{Name: left, Value: val}
	}
	return &ast.BindingExpr{Name: left, Value: val}
}

func (p *Parser) registerBinding(name string, fl *ast.FunctionLiteral, isRes bool) {
	usesLeft := bodyContainsName(fl.Body, "left")
	usesRight := bodyContainsName(fl.Body, "right")

	lbp := 100
	if fl.LBP != nil {
		lbp = *fl.LBP
	}

	if usesLeft && usesRight {
		if fl.RBP != nil {
			p.bpTable.RegisterCustomInfix(name, lbp, *fl.RBP)
		} else {
			p.bpTable.RegisterInfix(name, lbp)
		}
	} else if usesRight {
		p.bpTable.RegisterPrefix(name, lbp)
	} else {
		p.bpTable.RegisterValue(name)
	}
}

func bodyContainsName(stmts []ast.Statement, name string) bool {
	for _, s := range stmts {
		if nodeContainsName(s, name) {
			return true
		}
	}
	return false
}

func nodeContainsName(n ast.Node, name string) bool {
	if n == nil {
		return false
	}
	switch v := n.(type) {
	case *ast.Name:
		return v.Value == name
	case *ast.PrefixExpr:
		return nodeContainsName(v.Right, name)
	case *ast.InfixExpr:
		return nodeContainsName(v.Left, name) || nodeContainsName(v.Right, name)
	case *ast.BindingExpr:
		return nodeContainsName(v.Value, name)
	case *ast.DotExpr:
		return nodeContainsName(v.Left, name) || nodeContainsName(v.Key, name)
	case *ast.GroupExpr:
		return nodeContainsName(v.Inner, name)
	case *ast.ResourceDef:
		return nodeContainsName(v.Value, name)
	case *ast.ResourceInst:
		return nodeContainsName(v.Name, name)
	case *ast.ElvisExpr:
		return nodeContainsName(v.Left, name) || nodeContainsName(v.Right, name)
	case *ast.CommaExpr:
		return nodeContainsName(v.Left, name) || nodeContainsName(v.Right, name)
	case *ast.TableLiteral:
		for _, e := range v.Elements {
			if nodeContainsName(e, name) {
				return true
			}
		}
	case *ast.FunctionLiteral:
		return false
	case *ast.Program:
		return bodyContainsName(v.Statements, name)
	}
	return false
}

func (p *Parser) parseAtom() ast.Expression {
	t := p.curToken
	switch t.Type {
	case token.LPAREN:
		p.nextToken()
		inner := p.parseExpression(0)
		if p.curToken.Type == token.RPAREN {
			p.nextToken()
		} else {
			p.addError("expected ) after atom group")
		}
		return &ast.GroupExpr{Inner: inner}
	case token.LBRACE:
		return p.parseFunctionLiteral(nil)
	case token.IDENTIFIER:
		p.nextToken()
		return &ast.Name{Value: t.Literal}
	case token.INTEGER, token.DECIMAL, token.RATIONAL, token.STRING, token.DOCSTRING, token.RAWSTRING, token.RAWDOC, token.BOOLEAN:
		p.nextToken()
		switch t.Type {
		case token.INTEGER:
			return &ast.IntegerLiteral{Value: t.Literal}
		case token.DECIMAL:
			return &ast.DecimalLiteral{Value: t.Literal}
		case token.RATIONAL:
			parts := strings.Split(t.Literal, "/")
			if len(parts) == 2 {
				return &ast.RationalLiteral{Numerator: parts[0], Denominator: parts[1]}
			}
			return &ast.RationalLiteral{Numerator: t.Literal, Denominator: "1"}
		case token.STRING, token.DOCSTRING, token.RAWSTRING, token.RAWDOC:
			isDoc := t.Type == token.DOCSTRING || t.Type == token.RAWDOC
			isRaw := t.Type == token.RAWSTRING || t.Type == token.RAWDOC
			return &ast.StringLiteral{Value: t.Literal, IsDoc: isDoc, IsRaw: isRaw}
		case token.BOOLEAN:
			return &ast.BooleanLiteral{Value: t.Literal == "true"}
		}
		return &ast.Name{Value: t.Literal}
	}
	p.addError("expected atom")
	return &ast.ErrorExpr{Message: "expected atom"}
}

func (p *Parser) parseFunctionLiteral(lbp *int) *ast.FunctionLiteral {
	if p.curToken.Type == token.LBRACE {
		p.nextToken()
	}

	body := []ast.Statement{}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.SEMICOLON {
			p.nextToken()
			continue
		}

		stmt := p.parseExpression(0)
		if stmt != nil {
			if s, ok := stmt.(ast.Statement); ok {
				body = append(body, s)
			}
		}
	}

	if p.curToken.Type == token.RBRACE {
		p.nextToken()
	} else {
		p.addError("expected '}'")
	}

	var rbp *int
	if p.curToken.Type == token.INTEGER && p.areAdjacent(p.prevToken, p.curToken) {
		val, _ := strconv.Atoi(p.curToken.Literal)
		rbp = &val
		p.nextToken()
	}

	return &ast.FunctionLiteral{LBP: lbp, Body: body, RBP: rbp}
}

func (p *Parser) parseTableLiteral() *ast.TableLiteral {
	elements := []ast.Expression{}

	if p.curToken.Type == token.RBRACKET {
		p.nextToken()
		return &ast.TableLiteral{Elements: elements}
	}

	prevInTable := p.inTable
	p.inTable = true
	defer func() { p.inTable = prevInTable }()

	for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.SEMICOLON {
			p.addError("semicolons are not valid inside table literals")
			p.nextToken()
			continue
		}
		expr := p.parseExpression(0)
		if expr != nil {
			elements = append(elements, expr)
		}
	}

	if p.curToken.Type == token.RBRACKET {
		p.nextToken()
	} else {
		p.addError("expected ']'")
	}

	return &ast.TableLiteral{Elements: elements}
}

func (p *Parser) areAdjacent(t1, t2 token.Token) bool {
	return t1.Line == t2.Line && (t1.Column+len(t1.Literal) == t2.Column)
}

func (p *Parser) isPossibleExpressionStart() bool {
	t := p.curToken.Type
	if t == token.IDENTIFIER {
		entry, ok := p.bpTable.Lookup(p.curToken.Literal)
		if !ok {
			return true // Unknown identifier -> Start of expression (variable)
		}
		return entry.IsPrefix // Must be prefix capable (e.g. prefix op or value)
	}
	return t == token.INTEGER || t == token.DECIMAL || t == token.RATIONAL ||
		t == token.STRING || t == token.DOCSTRING || t == token.RAWSTRING || t == token.RAWDOC ||
		t == token.BOOLEAN || t == token.LPAREN || t == token.LBRACE || t == token.LBRACKET ||
		t == token.AT || t == token.AT_COLON // Resources
}
