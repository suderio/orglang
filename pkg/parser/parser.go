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
		return 800
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
			return &ast.InfixExpr{Left: left, Op: "|>", Right: p.parseAtom()}
		}
		if t.Literal == "o" {
			return &ast.InfixExpr{Left: left, Op: "o", Right: p.parseAtom()}
		}

		entry, _ := p.bpTable.Lookup(t.Literal)
		rbp := entry.RBP
		right := p.parseExpression(rbp)
		return &ast.InfixExpr{Left: left, Op: t.Literal, Right: right}

	case token.AT:
		bp := 800
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
		p.bpTable.RegisterInfix(name, lbp)
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
	case token.INTEGER, token.DECIMAL, token.RATIONAL, token.STRING, token.BOOLEAN:
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
	}

	var rbp *int
	if p.curToken.Type == token.INTEGER && p.areAdjacent(p.curToken, p.peekToken) {
		// Adjacency check for RBP
		// This implementation is still simplified.
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
		expr := p.parseExpression(0)
		if expr != nil {
			elements = append(elements, expr)
		}
	}

	if p.curToken.Type == token.RBRACKET {
		p.nextToken()
	}

	return &ast.TableLiteral{Elements: elements}
}

func (p *Parser) areAdjacent(t1, t2 token.Token) bool {
	return t1.Line == t2.Line && (t1.Column+len(t1.Literal) == t2.Column)
}
