package parser

import (
	"strings"
	"testing"

	"orglang/pkg/lexer"
)

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedAST    string
		expectedErrors []string
	}{
		{
			name:           "Unbalanced Parenthesis",
			input:          "( 1 + 1",
			expectedAST:    "((1 + 1))",
			expectedErrors: []string{"expected ')'"},
		},
		{
			name:           "Unbalanced Bracket",
			input:          "[ 1 2",
			expectedAST:    "[1 2]",
			expectedErrors: []string{"expected ']'"},
		},
		{
			name:           "Unbalanced Brace",
			input:          "{ 1 + 1",
			expectedAST:    "{ (1 + 1) }",
			expectedErrors: []string{"expected '}'"},
		},
		{
			name:           "Undefined Identifier",
			input:          "unknown_id",
			expectedAST:    "<Error: undefined identifier: unknown_id>",
			expectedErrors: nil, // ErrorExpr in AST, not in p.errors list (design choice?)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]byte(tt.input))
			p := New(l)
			prog := p.ParseProgram()

			// Check AST string
			astStr := strings.TrimSpace(prog.String())
			if astStr != tt.expectedAST {
				t.Errorf("AST mismatch.\nExpected: %q\nGot:      %q", tt.expectedAST, astStr)
			}

			// Check specific errors
			errors := p.Errors()
			if len(tt.expectedErrors) == 0 {
				if len(errors) > 0 {
					t.Errorf("expected no errors, got %d: %v", len(errors), errors)
				}
			} else {
				if len(errors) != len(tt.expectedErrors) {
					t.Errorf("expected %d errors, got %d: %v", len(tt.expectedErrors), len(errors), errors)
				} else {
					for i, expectedMsg := range tt.expectedErrors {
						if !strings.Contains(errors[i], expectedMsg) {
							t.Errorf("error %d mismatch.\nExpected to contain: %q\nGot: %q", i, expectedMsg, errors[i])
						}
					}
				}
			}
		})
	}
}
