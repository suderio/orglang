package codegen

import (
	"bytes"
	"fmt"
	"orglang/internal/ast" // Added this import
	"strings"
	"text/template"
)

// Added this global variable
var primitives = map[string]bool{
	"sys": true,
	"mem": true,
	"org": true,
}

const cTemplate = `#include "orglang.h"
#include <stdio.h>

// Globals
{{ .Globals }}

// Auxiliary Functions
{{ .AuxFunctions }}

int main(int argc, char **argv) {
    org_argc = argc;
    org_argv = argv;
    Arena *arena = arena_create(1024 * 1024);
    
    OrgContext ctx;
    org_sched_init(&ctx, arena);

    // Program start
{{ .Body }}
    // Program end
    
    org_sched_run(&ctx);

    arena_free(arena);
    return 0;
}
`

type CEmitter struct {
	tmpl           *template.Template
	auxFunctions   []string
	funcCounter    int
	ModuleLoader   func(string) (*ast.Program, error)
	ModuleRegistry map[string]string // Path -> FunctionName
	bindings       map[string]bool
	ScopePrefix    string
}

func NewCEmitter(loader func(string) (*ast.Program, error)) *CEmitter {
	t, _ := template.New("c").Parse(cTemplate)
	return &CEmitter{
		tmpl:           t,
		auxFunctions:   []string{},
		funcCounter:    0,
		ModuleLoader:   loader,
		ModuleRegistry: make(map[string]string),
		bindings:       make(map[string]bool),
		ScopePrefix:    "",
	}
}

type TemplateData struct {
	Globals      string
	Body         string
	AuxFunctions string
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

	// Pre-declarations for globals
	var globalsBuilder strings.Builder
	for b := range c.bindings {
		// FIX: Use mangleIdentifier(b) to ensure valid C identifiers
		globalsBuilder.WriteString(fmt.Sprintf("static OrgValue *org_var_%s = NULL;\n", mangleIdentifier(b)))
	}

	data := TemplateData{
		Globals:      globalsBuilder.String(),
		Body:         bodyBuilder.String(),
		AuxFunctions: strings.Join(c.auxFunctions, "\n"),
	}

	var output bytes.Buffer
	if err := c.tmpl.Execute(&output, data); err != nil {
		return "", err
	}

	// Post-process logic for main enforcement
	// We want to add this to the Body, but since Body is already stringified in data,
	// and we are using a template, it's cleaner to append to Body in the template data
	// OR just append to the output if the template structure allows.
	// Actually, let's look at the template:
	// {{ .Body }}
	// // Program end
	// org_sched_run(&ctx);
	//
	// We want to execute 'main' BEFORE org_sched_run.
	// So we can append to bodyBuilder before creating data.

	// Append main execution logic to bodyBuilder
	mainCheck := `
    if (org_var_main) {
        org_call(arena, org_var_main, NULL, NULL);
    } else {
        fprintf(stderr, "No main key in the org file\n");
        exit(1);
    }
`
	bodyBuilder.WriteString(mainCheck)

	// Re-create data with updated Body
	data = TemplateData{
		Globals:      globalsBuilder.String(),
		Body:         bodyBuilder.String(),
		AuxFunctions: strings.Join(c.auxFunctions, "\n"),
	}

	output.Reset()
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
		if e.Operator == "@" {
			if ident, ok := e.Right.(*ast.Identifier); ok && ident.Value == "args" {
				return "org_resource_args_create_wrap(arena)", nil
			}
		}
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
			// REMOVED SPECIAL CASE. Handled by runtime.
			// if rightPrefix, ok := e.Right.(*ast.PrefixExpression); ok && rightPrefix.Operator == "@" {
			// 	if rightIdent, ok := rightPrefix.Right.(*ast.Identifier); ok && rightIdent.Value == "stdout" {
			// 		return fmt.Sprintf("org_print(arena, %s)", left), nil
			// 	}
			// }
		}
		if e.Operator == "@" {
			// Infix @ : Primitives or Modules
			// Left @ Right

			// 1. Check for Module Import: "path" @ org
			if rightIdent, ok := e.Right.(*ast.Identifier); ok && rightIdent.Value == "org" {
				if pathLit, ok := e.Left.(*ast.StringLiteral); ok {
					return c.compileModule(pathLit.Value)
				}
			}

			rightIdent, ok := e.Right.(*ast.Identifier)
			if !ok {
				// Maybe it's @(sys) or similar?
				// For now expect identifier.
				return "", fmt.Errorf("expected identifier on right side of @, got %T", e.Right)
			}

			if rightIdent.Value == "sys" {
				// org_syscall(arena, left)
				return fmt.Sprintf("org_syscall(arena, %s)", left), nil
			} else if rightIdent.Value == "mem" {
				// org_malloc(arena, org_value_to_long(left))
				return fmt.Sprintf("org_malloc(arena, org_value_to_long(%s))", left), nil
			} else if rightIdent.Value == "args" {
				return "org_resource_args_create_wrap(arena)", nil
			} else if rightIdent.Value == "stdout" {
				return "org_resource_stdout_create_wrap(arena)", nil
			} else if rightIdent.Value == "org" {
				// Handled above if Left is String.
				// If Left is not string, runtime error or dynamic load?
				// For now error.
				return "", fmt.Errorf("left side of @ org must be a string literal (path)")
			} else {
				return "", fmt.Errorf("unknown resource primitive: @%s", rightIdent.Value)
			}
		}

		if e.Operator == "." || e.Operator == "?" {
			// Table Access
			// Left is Table, Right is Key (Expression)
			// We need to pass the Key as an OrgValue to org_table_get.
			// Currently we emitExpression(Right) which generates code that returns OrgValue*.
			// This works for Indices (IntegerLiteral -> org_int_from_str).
			// For Identifiers (table.key), parsed as Identifier "key". emitExpression("key") -> returns value of variable "key".
			// BUT `.` implies literal key.
			// We need to handle this divergence.

			var keyCode string

			// Check if Right is Identifier or StringLiteral acting as Key
			if ident, ok := e.Right.(*ast.Identifier); ok {
				// Treat as String Key: org_string_from_c(arena, "ident")
				keyCode = fmt.Sprintf("org_string_from_c(arena, \"%s\")", ident.Value)
			} else if _, ok := e.Right.(*ast.IntegerLiteral); ok {
				// Integer Literal: standard emit
				var err error
				keyCode, err = c.emitExpression(e.Right)
				if err != nil {
					return "", err
				}
			} else {
				// Fallback: evaluate expression (dynamic key)
				var err error
				keyCode, err = c.emitExpression(e.Right)
				if err != nil {
					return "", err
				}
			}

			getCall := fmt.Sprintf("org_table_get(arena, %s, %s)", left, keyCode)

			// Clean logic:
			if e.Operator == "?" {
				// Left is Cond (Key), Right is Table
				tableCode, err := c.emitExpression(e.Right)
				if err != nil {
					return "", err
				}

				keyCode, err := c.emitExpression(e.Left)
				if err != nil {
					return "", err
				}

				// If Key is bool? (n > 0). BooleanLiteral or Expression.
				// We need key as OrgValue*.

				getCall = fmt.Sprintf("org_table_get(arena, %s, %s)", tableCode, keyCode)
				return fmt.Sprintf("org_value_evaluate(arena, %s)", getCall), nil
			}

			// Fallthrough for . (Dot)
			// Left is Table, Right is Key.
			return getCall, nil
		}

		if e.Operator == ":" {
			// Pair: left : right -> [left, right]
			// Pair: left : right -> [left, right]
			// If left is Identifier, it's a binding.
			if ident, ok := e.Left.(*ast.Identifier); ok {
				name := ident.Value
				if c.ScopePrefix != "" {
					name = c.ScopePrefix + "_" + name
				}
				c.bindings[name] = true
				mangled := mangleIdentifier(name)
				return fmt.Sprintf("(org_var_%s = %s, org_pair_make(arena, org_string_from_c(arena, \"%s\"), org_var_%s))",
					mangled, right, ident.Value, mangled), nil
			}
			return fmt.Sprintf("org_pair_make(arena, %s, %s)", left, right), nil
		}

		return fmt.Sprintf("org_op_infix(arena, \"%s\", %s, %s)", e.Operator, left, right), nil

	case *ast.CallExpression:
		// 1. Emit Function Expression
		fnCode, err := c.emitExpression(e.Function)
		if err != nil {
			return "", err
		}

		// 2. Emit Arguments
		var argCode string
		if len(e.Arguments) == 0 {
			argCode = "NULL"
		} else if len(e.Arguments) == 1 {
			argCode, err = c.emitExpression(e.Arguments[0])
			if err != nil {
				return "", err
			}
		} else {
			// Wrap multiple arguments in a list
			var argBuilder strings.Builder
			for _, arg := range e.Arguments {
				val, err := c.emitExpression(arg)
				if err != nil {
					return "", err
				}
				argBuilder.WriteString(", ")
				argBuilder.WriteString(val)
			}
			argCode = fmt.Sprintf("org_list_make(arena, %d%s)", len(e.Arguments), argBuilder.String())
		}

		// 3. Emit Call
		return fmt.Sprintf("org_call(arena, %s, NULL, %s)", fnCode, argCode), nil

	case *ast.BlockLiteral:
		return c.emitBlockLiteral(e)

	case *ast.ResourceLiteral:
		return c.emitResourceLiteral(e)

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
		if e.Value == "left" {
			return "left", nil
		}
		if e.Value == "right" {
			return "right", nil
		}
		if e.Value == "this" {
			return "func", nil
		}
		if primitives[e.Value] {
			return fmt.Sprintf("org_string_from_c(arena, \"%s\")", e.Value), nil
		}
		// Apply scope prefix if present (except for 'main' in the root scope? NO, main is special)
		// Actually, if we mangle 'main', the entry point check needs to know the mangled name.
		// For the ROOT program, ScopePrefix is empty. So 'org_var_main'.
		// For modules, ScopePrefix is set. So 'org_var_mod1_main'.
		// This handles the collision perfectly!

		name := e.Value

		// Global exceptions that should NOT be scoped/mangled with module prefix
		// These are defined in stdlib or runtime and are global.
		if name == "stdout" || name == "stdin" || name == "stderr" || name == "Error" || name == "print" {
			// Do not prefix.
		} else if c.ScopePrefix != "" {
			name = c.ScopePrefix + "_" + name
		}

		return fmt.Sprintf("org_var_%s", mangleIdentifier(name)), nil

	case *ast.BooleanLiteral:
		if e.Value {
			return "org_bool(arena, 1)", nil
		}
		return "org_bool(arena, 0)", nil

	default:
		return "", fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (c *CEmitter) emitBlockLiteral(bl *ast.BlockLiteral) (string, error) {
	fnName := fmt.Sprintf("org_fn_%d", c.funcCounter)
	c.funcCounter++

	var bodyBuilder strings.Builder

	// Iterate statements
	for i, stmt := range bl.Statements {
		stmtCode, err := c.emitStatement(stmt)
		if err != nil {
			return "", err
		}

		if i == len(bl.Statements)-1 {
			bodyBuilder.WriteString("return " + stmtCode + ";\n")
		} else {
			bodyBuilder.WriteString(stmtCode + ";\n")
		}
	}

	// If empty body
	if len(bl.Statements) == 0 {
		bodyBuilder.WriteString("return NULL;\n")
	}

	fnParams := "Arena *arena, OrgValue *func, OrgValue *left, OrgValue *right"
	fnDef := fmt.Sprintf("static OrgValue *%s(%s) {\n%s}\n", fnName, fnParams, bodyBuilder.String())

	c.auxFunctions = append(c.auxFunctions, fnDef)

	return fmt.Sprintf("org_func_create(arena, %s)", fnName), nil
}

func (c *CEmitter) flattenComma(expr ast.Expression) []ast.Expression {
	if infix, ok := expr.(*ast.InfixExpression); ok && infix.Operator == "," {
		return append(c.flattenComma(infix.Left), c.flattenComma(infix.Right)...)
	}
	return []ast.Expression{expr}
}

func (c *CEmitter) emitResourceLiteral(rl *ast.ResourceLiteral) (string, error) {
	// Parse Body (ListLiteral) for setup, step, teardown, next keys
	var setup, step, teardown, nextStr string = "NULL", "NULL", "NULL", "NULL"

	// Iterate over elements
	if rl.Body != nil {
		for _, el := range rl.Body.Elements {
			for _, item := range c.flattenComma(el) {
				// Look for InfixExpression with operator ":"
				if infix, ok := item.(*ast.InfixExpression); ok && infix.Operator == ":" {
					// Left should be identifier
					if ident, ok := infix.Left.(*ast.Identifier); ok {
						valCode, err := c.emitExpression(infix.Right)
						if err != nil {
							return "", err
						}

						switch ident.Value {
						case "setup":
							setup = valCode
						case "step":
							step = valCode
						case "teardown":
							teardown = valCode
						case "next":
							nextStr = valCode
						}
					}
				}
			}
		}
	}

	return fmt.Sprintf("org_resource_create(arena, %s, %s, %s, %s)", setup, step, teardown, nextStr), nil
}

func (c *CEmitter) compileModule(path string) (string, error) {
	// TODO: Resolve absolute path relative to current module?
	// For now, assume relative to execution or absolute.
	// But duplicate check needs canonical path.
	// Since we don't have CWD easily in emitter, we rely on path as key.
	// main.go ModuleLoader receives path.

	// Check Registry
	if fnName, ok := c.ModuleRegistry[path]; ok {
		return fmt.Sprintf("%s(arena, NULL, NULL)", fnName), nil
	}

	// Load Module
	if c.ModuleLoader == nil {
		return "", fmt.Errorf("module loader not initialized")
	}
	program, err := c.ModuleLoader(path)
	if err != nil {
		return "", err
	}

	// Compile to Function
	// Use funcCounter as unique module ID for scoping
	scopeID := fmt.Sprintf("mod%d", c.funcCounter)

	fnName := fmt.Sprintf("org_module_%d", c.funcCounter)
	c.funcCounter++
	c.ModuleRegistry[path] = fnName

	// SCOPE FIX:
	// We use the global 'c.bindings' map to track ALL variables, so they are declared at the top of the file.
	// To prevent collisions, we use 'c.ScopePrefix' to mangle names within this module.

	parentPrefix := c.ScopePrefix
	c.ScopePrefix = scopeID

	var bodyBuilder strings.Builder
	var resultVars []string

	for i, stmt := range program.Statements {
		stmtCode, err := c.emitStatement(stmt)
		if err != nil {
			c.ScopePrefix = parentPrefix // Restore on error
			return "", err
		}

		resVar := fmt.Sprintf("stmt_%d", i)
		bodyBuilder.WriteString(fmt.Sprintf("    OrgValue *%s = %s;\n", resVar, stmtCode))
		resultVars = append(resultVars, resVar)
	}

	// We do NOT declare bindings locally anymore. They are global.

	finalBody := bodyBuilder.String()

	if len(resultVars) == 0 {
		finalBody += "    return org_list_make(arena, 0);\n"
	} else {
		args := strings.Join(resultVars, ", ")
		finalBody += fmt.Sprintf("    return org_list_make(arena, %d, %s);\n", len(resultVars), args)
	}

	fnParams := "Arena *arena, OrgValue *this_val, OrgValue *args"
	fnDef := fmt.Sprintf("static OrgValue *%s(%s) {\n%s}\n", fnName, fnParams, finalBody)

	c.auxFunctions = append(c.auxFunctions, fnDef)

	// Restore parent scope
	c.ScopePrefix = parentPrefix

	return fmt.Sprintf("%s(arena, NULL, NULL)", fnName), nil
}

func mangleIdentifier(ident string) string {
	var sb strings.Builder
	for _, ch := range ident {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			sb.WriteRune(ch)
		} else {
			sb.WriteString(fmt.Sprintf("_%X", ch))
		}
	}
	return sb.String()
}
