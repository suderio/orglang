package ast

import (
	"fmt"
	"strings"
)

// Node interface for all AST nodes
type Node interface {
	String() string
}

// Statement interface (for now, same as Node, but semantic distinction)
type Statement interface {
	Node
	statementNode()
}

// Expression interface
type Expression interface {
	Node
	expressionNode()
}

// Program node
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// --- Literals ---

type IntegerLiteral struct {
	Value string
}

func (il *IntegerLiteral) String() string  { return il.Value }
func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) statementNode()  {}

type DecimalLiteral struct {
	Value string
}

func (dl *DecimalLiteral) String() string  { return dl.Value }
func (dl *DecimalLiteral) expressionNode() {}
func (dl *DecimalLiteral) statementNode()  {}

type RationalLiteral struct {
	Numerator   string
	Denominator string
}

func (rl *RationalLiteral) String() string  { return fmt.Sprintf("%s/%s", rl.Numerator, rl.Denominator) }
func (rl *RationalLiteral) expressionNode() {}
func (rl *RationalLiteral) statementNode()  {}

type StringLiteral struct {
	Value string
	IsDoc bool
	IsRaw bool
}

func (sl *StringLiteral) String() string {
	if sl.IsDoc {
		return fmt.Sprintf(`"""%s"""`, sl.Value)
	}
	if sl.IsRaw {
		return fmt.Sprintf(`'%s'`, sl.Value)
	}
	return fmt.Sprintf(`"%s"`, sl.Value)
}
func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) statementNode()  {}

type BooleanLiteral struct {
	Value bool
}

func (bl *BooleanLiteral) String() string  { return fmt.Sprintf("%t", bl.Value) }
func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) statementNode()  {}

// FunctionLiteral represents { ... } or N{ ... }M
type FunctionLiteral struct {
	LBP  *int // Leading Binding Power (optional)
	Body []Statement
	RBP  *int // Right Binding Power (optional)
}

func (fl *FunctionLiteral) String() string {
	var out strings.Builder
	if fl.LBP != nil {
		out.WriteString(fmt.Sprintf("%d", *fl.LBP))
	}
	out.WriteString("{ ")
	for i, s := range fl.Body {
		if i > 0 {
			out.WriteString("; ")
		}
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	if fl.RBP != nil {
		out.WriteString(fmt.Sprintf("%d", *fl.RBP))
	}
	return out.String()
}
func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) statementNode()  {}

// TableLiteral represents [...]
type TableLiteral struct {
	Elements []Expression
}

func (tl *TableLiteral) String() string {
	var out strings.Builder
	out.WriteString("[")
	for i, e := range tl.Elements {
		if i > 0 {
			out.WriteString(" ")
		}
		out.WriteString(e.String())
	}
	out.WriteString("]")
	return out.String()
}
func (tl *TableLiteral) expressionNode() {}
func (tl *TableLiteral) statementNode()  {}

// --- Expressions ---

type Name struct {
	Value string
}

func (n *Name) String() string  { return n.Value }
func (n *Name) expressionNode() {}
func (n *Name) statementNode()  {}

type PrefixExpr struct {
	Op    string
	Right Expression
}

func (pe *PrefixExpr) String() string {
	return fmt.Sprintf("(%s%s)", pe.Op, pe.Right.String())
}
func (pe *PrefixExpr) expressionNode() {}
func (pe *PrefixExpr) statementNode()  {}

type InfixExpr struct {
	Left  Expression
	Op    string
	Right Expression
}

func (ie *InfixExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Op, ie.Right.String())
}
func (ie *InfixExpr) expressionNode() {}
func (ie *InfixExpr) statementNode()  {}

// DotExpr represents left.key
type DotExpr struct {
	Left Expression
	Key  Expression
}

func (de *DotExpr) String() string {
	return fmt.Sprintf("(%s.%s)", de.Left.String(), de.Key.String())
}
func (de *DotExpr) expressionNode() {}
func (de *DotExpr) statementNode()  {}

// BindingExpr represents name : value
type BindingExpr struct {
	Name  Expression
	Value Expression
}

func (be *BindingExpr) String() string {
	return fmt.Sprintf("(%s : %s)", be.Name.String(), be.Value.String())
}
func (be *BindingExpr) expressionNode() {}
func (be *BindingExpr) statementNode()  {}

// ResourceDef represents name @: value
type ResourceDef struct {
	Name  Expression
	Value Expression
}

func (rd *ResourceDef) String() string {
	return fmt.Sprintf("(%s @: %s)", rd.Name.String(), rd.Value.String())
}
func (rd *ResourceDef) expressionNode() {}
func (rd *ResourceDef) statementNode()  {}

// ResourceInst represents @name
type ResourceInst struct {
	Name Expression
}

func (ri *ResourceInst) String() string {
	return fmt.Sprintf("@%s", ri.Name.String())
}
func (ri *ResourceInst) expressionNode() {}
func (ri *ResourceInst) statementNode()  {}

// ElvisExpr represents left ?: right
type ElvisExpr struct {
	Left  Expression
	Right Expression
}

func (ee *ElvisExpr) String() string {
	return fmt.Sprintf("(%s ?: %s)", ee.Left.String(), ee.Right.String())
}
func (ee *ElvisExpr) expressionNode() {}
func (ee *ElvisExpr) statementNode()  {}

// CommaExpr represents left, right
type CommaExpr struct {
	Left  Expression
	Right Expression
}

func (ce *CommaExpr) String() string {
	return fmt.Sprintf("(%s , %s)", ce.Left.String(), ce.Right.String())
}
func (ce *CommaExpr) expressionNode() {}
func (ce *CommaExpr) statementNode()  {}

// GroupExpr represents (inner)
type GroupExpr struct {
	Inner Expression
}

func (ge *GroupExpr) String() string {
	return fmt.Sprintf("(%s)", ge.Inner.String())
}
func (ge *GroupExpr) expressionNode() {}
func (ge *GroupExpr) statementNode()  {}

// BlockExpr represents { ... } when used as an expression (same as FunctionLiteral but explicit naming if needed)
// We treat { ... } as FunctionLiteral.

// ErrorExpr represents a parsing error or undefined identifier
type ErrorExpr struct {
	Message string
}

func (ee *ErrorExpr) String() string {
	return fmt.Sprintf("<Error: %s>", ee.Message)
}
func (ee *ErrorExpr) expressionNode() {}
func (ee *ErrorExpr) statementNode()  {}
