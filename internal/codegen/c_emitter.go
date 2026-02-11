package codegen

import (
	"bytes"
	"fmt"
	"orglang/internal/ast"
	"strings"
	"text/template"
)

const cTemplate = `#include "orglang.h"
#include <stdio.h>

int main() {
    Arena *arena = arena_create(1024 * 1024);
    
    // Program start
{{ .Body }}
    // Program end
    
    arena_free(arena);
    return 0;
}
`

type CEmitter struct {
	tmpl *template.Template
}

func NewCEmitter() *CEmitter {
	t, _ := template.New("c").Parse(cTemplate)
	return &CEmitter{tmpl: t}
}

type TemplateData struct {
	Body string
}

func (c *CEmitter) Generate(program *ast.Program) (string, error) {
	var bodyBuilder strings.Builder

	for _, stmt := range program.Statements {
		code, err := c.emitStatement(stmt)
		if err != nil {
			return "", err
		}
		bodyBuilder.WriteString("    " + code + ";\n")
	}

	data := TemplateData{
		Body: bodyBuilder.String(),
	}

	var output bytes.Buffer
	if err := c.tmpl.Execute(&output, data); err != nil {
		return "", err
	}

	return output.String(), nil
}

func (c *CEmitter) emitStatement(stmt ast.Statement) (string, error) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return c.emitExpression(s.Expression)
	default:
		return "", fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *CEmitter) emitExpression(expr ast.Expression) (string, error) {
	if expr == nil {
		return "", nil
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("org_int_from_str(arena, \"%s\")", e.Value), nil

	case *ast.DecimalLiteral:
		return fmt.Sprintf("org_dec_from_str(arena, \"%s\")", e.Value), nil

	case *ast.StringLiteral:
		return fmt.Sprintf("org_string_from_c(arena, \"%s\")", e.Value), nil

	case *ast.PrefixExpression:
		right, err := c.emitExpression(e.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("org_op_prefix(arena, \"%s\", %s)", e.Operator, right), nil

	case *ast.InfixExpression:
		left, err := c.emitExpression(e.Left)
		if err != nil {
			return "", err
		}
		right, err := c.emitExpression(e.Right)
		if err != nil {
			return "", err
		}
		// `org_op_infix(arena, "+", left, right)`
		if e.Operator == "->" {
			// Check for @stdout
			// right side is PrefixExpression(@) -> Identifier(stdout)
			if rightPrefix, ok := e.Right.(*ast.PrefixExpression); ok && rightPrefix.Operator == "@" {
				if rightIdent, ok := rightPrefix.Right.(*ast.Identifier); ok && rightIdent.Value == "stdout" {
					return fmt.Sprintf("org_print(arena, %s)", left), nil
				}
			}
		}
		return fmt.Sprintf("org_op_infix(arena, \"%s\", %s, %s)", e.Operator, left, right), nil

	case *ast.CallExpression:
		fn, ok := e.Function.(*ast.Identifier)
		if !ok {
			return "", fmt.Errorf("complex function calls not supported yet")
		}
		if fn.Value == "print" {
			if len(e.Arguments) != 1 {
				return "", fmt.Errorf("print() expects exactly 1 argument")
			}
			arg, err := c.emitExpression(e.Arguments[0])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("org_print(arena, %s)", arg), nil
		}
		return "", fmt.Errorf("unknown function: %s", fn.Value)

	case *ast.BlockLiteral:
		return "NULL /* Block not implemented */", nil

	case *ast.ListLiteral:
		var argBuilder strings.Builder
		count := len(e.Elements)
		for _, el := range e.Elements {
			val, err := c.emitExpression(el)
			if err != nil {
				return "", err
			}
			argBuilder.WriteString(", ")
			argBuilder.WriteString(val)
		}
		// e.g. org_list_make(arena, 3, a, b, c)
		return fmt.Sprintf("org_list_make(arena, %d%s)", count, argBuilder.String()), nil

	case *ast.GroupExpression:
		return c.emitExpression(e.Expression)

	case *ast.Identifier:
		return e.Value, nil

	case *ast.BooleanLiteral:
		if e.Value {
			return "org_bool(arena, 1)", nil
		}
		return "org_bool(arena, 0)", nil

	default:
		return "", fmt.Errorf("unknown expression type: %T", expr)
	}
}
