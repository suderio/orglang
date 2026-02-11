package ast

import (
	"bytes"
	"orglang/pkg/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// Literals

type IntegerLiteral struct {
	Token token.Token
	Value string // Arbitrary precision
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type DecimalLiteral struct {
	Token token.Token
	Value string // Arbitrary precision
}

func (dl *DecimalLiteral) expressionNode()      {}
func (dl *DecimalLiteral) TokenLiteral() string { return dl.Token.Literal }
func (dl *DecimalLiteral) String() string       { return dl.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Token.Literal + "\"" } // Re-quote?

type ErrorLiteral struct {
	Token token.Token
}

func (el *ErrorLiteral) expressionNode()      {}
func (el *ErrorLiteral) TokenLiteral() string { return el.Token.Literal }
func (el *ErrorLiteral) String() string       { return el.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. - or ~ or @
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	if pe.Operator == "@" {
		// usually @resource
	} else {
		out.WriteString(" ") // Space required for ~ and - based on updated spec
	}
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type GroupExpression struct {
	Token      token.Token // '('
	Expression Expression
}

func (ge *GroupExpression) expressionNode()      {}
func (ge *GroupExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupExpression) String() string {
	return "(" + ge.Expression.String() + ")"
}

type ListLiteral struct {
	Token    token.Token  // '['
	Elements []Expression // Changed from Content to Elements
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	for i, el := range ll.Elements {
		out.WriteString(el.String())
		if i < len(ll.Elements)-1 {
			out.WriteString(" ")
		}
	}
	out.WriteString("]")
	return out.String()
}

type BlockLiteral struct {
	Token token.Token // '{'
	Body  Expression  // Single expression/statement list?
	// Spec says Block ::= (INTEGER)? "{" Expression "}" (INTEGER)?
	// Usually Expression is a Statement list or a single Expression?
	// "Program ::= Statement*", "Statement ::= Expression ';'"
	// So Block body is likely Expression (which can be a Sequence/BlockExpression?)
	// For now, single Expression.
	LeftBP  string // Optional binding power (as string for now "700")
	RightBP string // Optional
}

func (bl *BlockLiteral) expressionNode()      {}
func (bl *BlockLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BlockLiteral) String() string {
	var out bytes.Buffer
	if bl.LeftBP != "" {
		out.WriteString(bl.LeftBP)
	}
	out.WriteString("{ ")
	if bl.Body != nil {
		out.WriteString(bl.Body.String())
	}
	out.WriteString(" }")
	if bl.RightBP != "" {
		out.WriteString(bl.RightBP)
	}
	return out.String()
}

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	// arguments separated by comma?
	// Spec: "f x, y" or "f(x, y)"?
	// If parser uses comma, we join by comma.
	// We will implement parser to use comma.
	for i, arg := range args {
		out.WriteString(arg)
		if i < len(args)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")

	return out.String()
}
