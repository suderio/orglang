package codegen

import (
	"orglang/internal/parser"
	"orglang/pkg/lexer"
	"strings"
	"testing"
)

func TestCEmitter(t *testing.T) {
	input := `
    1 + 2;
    -3.14;
    ~true;
    "hello";
    `

	l := lexer.NewCustom(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	emitter := NewCEmitter(nil)
	output, err := emitter.Generate(program)
	if err != nil {
		t.Fatalf("emitter error: %v", err)
	}

	// Basic checks
	if !strings.Contains(output, "#include \"orglang.h\"") {
		t.Error("Output missing standard header")
	}
	if !strings.Contains(output, "int main() {") {
		t.Error("Output missing main function")
	}

	// Check for specific generated calls
	if !strings.Contains(output, "org_int_from_str(arena, \"1\")") {
		t.Error("Missing integer generation")
	}
	if !strings.Contains(output, "org_op_infix(arena, \"+\",") {
		t.Error("Missing infix op generation")
	}
	if !strings.Contains(output, "org_dec_from_str(arena, \"-3.14\")") {
		t.Error("Missing decimal generation")
	}
	// ~true -> prefix ~, IDENT true.
	// IDENT not handled in emitter yet?
	// Wait, `~true`. `true` is BOOLEAN Literal? Or Identifier?
	// Lexer: `true` -> BOOLEAN?
	// Lexer: `case "true": tok.Type = token.TRUE`?
	// Let's check token.go/lexer.go.
	// If Lexer returns BOOLEAN token, parser needs `parseBoolean`?
	// Parser has `registerPrefix(token.TRUE, p.parseBooleanLiteral)`?
	// I didn't verify Boolean support in Parser Phase.
	// If parser sees `true` as IDENT, then `~true` is `Prefix(~, Ident(true))`.
	// Emitter handles Prefix. Identifier?
	// Emitter `emitExpression` switch default error.
	// So `~true` might fail if `true` is Identifier and Identifier case missing.

}
