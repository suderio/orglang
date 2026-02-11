package parser

import (
	"orglang/internal/ast"
	"orglang/pkg/lexer"
	"testing"
)

func TestIntegerLiteralExpression(t *testing.T) {
	input := "10;"

	l := lexer.NewCustom(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "10" {
		t.Errorf("literal.Value not %s. got=%s", "10", literal.Value)
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    string
	}{
		{"- 15;", "-", "15"},
		{"~ 15;", "~", "15"},
		{"@stdout;", "@", "stdout"},
	}

	for _, tt := range prefixTests {
		l := lexer.NewCustom(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}

		// Right side check simplistic for now
		if exp.Right.String() != tt.value {
			// For @stdout, Right is Identifier "stdout". String() returns "stdout".
			// For - 15, Right is IntegerLiteral "15".
			t.Errorf("exp.Right is not %s. got=%s", tt.value, exp.Right.String())
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  string
		operator   string
		rightValue string
	}{
		{"5 + 5;", "5", "+", "5"},
		{"5 - 5;", "5", "-", "5"},
		{"5 * 5;", "5", "*", "5"},
		{"5 / 5;", "5", "/", "5"},
		{"5 > 5;", "5", ">", "5"},
		{"5 < 5;", "5", "<", "5"},
		{"5 = 5;", "5", "=", "5"},
		{"5 <> 5;", "5", "<>", "5"},
	}

	for _, tt := range infixTests {
		l := lexer.NewCustom(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp is not ast.InfixExpression. got=%T", stmt.Expression)
		}

		if exp.Left.String() != tt.leftValue {
			t.Errorf("exp.Left is not %s. got=%s", tt.leftValue, exp.Left.String())
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}

		if exp.Right.String() != tt.rightValue {
			t.Errorf("exp.Right is not %s. got=%s", tt.rightValue, exp.Right.String())
		}
	}
}

func TestComplexBindingPower(t *testing.T) {
	// Tests new binding power syntax for blocks
	input := "700{ x + 1 }800;"

	l := lexer.NewCustom(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	block, ok := stmt.Expression.(*ast.BlockLiteral)
	if !ok {
		t.Fatalf("exp not BlockLiteral. got %T", stmt.Expression)
	}

	if block.LeftBP != "700" {
		t.Errorf("LeftBP wrong. exp=700 got=%s", block.LeftBP)
	}
	if block.RightBP != "800" {
		t.Errorf("RightBP wrong. exp=800 got=%s", block.RightBP)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
