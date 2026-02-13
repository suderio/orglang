package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"orglang/internal/ast"
	"orglang/internal/codegen"
	"orglang/internal/parser"

	// "orglang/internal/analysis" // TODO: Enable when analysis pass is ready
	"orglang/pkg/lexer"
	"orglang/pkg/stdlib"
)

var (
	version = "0.1.0"

	// Style
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "org",
		Short: "OrgLang Compiler",
		Long:  `OrgLang is a new programming language focused on orthogonality and aesthetics.`,
	}

	var outputFlag string

	var buildCmd = &cobra.Command{
		Use:   "build [file]",
		Short: "Compile an OrgLang source file to an executable",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sourceFile := args[0]
			outputFile := outputFlag
			if outputFile == "" {
				outputFile = strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
			}

			if err := compile(sourceFile, outputFile); err != nil {
				fmt.Println(errorStyle.Render("Build Failed:"))
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println(successStyle.Render("Build Successful: " + outputFile))
		},
	}

	buildCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output binary name")

	var runCmd = &cobra.Command{
		Use:   "run [file] [args...]",
		Short: "Compile and run an OrgLang source file",
		Args:  cobra.MinimumNArgs(1),
		Run:   runRun,
	}

	rootCmd.AddCommand(buildCmd, runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runRun(cmd *cobra.Command, args []string) {
	sourceFile := args[0]
	// Use a temp file for output or just build locally
	outputFile := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))

	if err := compile(sourceFile, outputFile); err != nil {
		fmt.Println(errorStyle.Render("Build Failed:"))
		fmt.Println(err)
		os.Exit(1)
	}

	// Run the binary
	absPath, _ := filepath.Abs(outputFile)
	// Pass remaining args to the executable
	runArgs := args[1:]
	runCmd := exec.Command(absPath, runArgs...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	if err := runCmd.Run(); err != nil {
		fmt.Println(errorStyle.Render("Runtime Error:"))
		fmt.Println(err)
		os.Exit(1)
	}
}

func compile(sourcePath, key string) error {
	// 1. Read Source
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading source: %w", err)
	}

	// 2. Lex & Parse Default Library
	lDef := lexer.NewCustom(stdlib.DefaultLibrary)
	pDef := parser.New(lDef)
	progDef := pDef.ParseProgram()
	if len(pDef.Errors()) > 0 {
		return fmt.Errorf("default library parsing errors:\n%s", strings.Join(pDef.Errors(), "\n"))
	}

	// 3. Lex & Parse User Source
	l := lexer.NewCustom(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parsing errors:\n%s", strings.Join(p.Errors(), "\n"))
	}

	// Merge Statements: Default Lib first
	finalStmts := append(progDef.Statements, program.Statements...)
	program.Statements = finalStmts

	// 3. Semantic Analysis
	// (Symbol table and flow graph would happen here)
	// For now, simpler passes or skip if fully handled in parser/codegen

	// 4. Codegen (C99)
	moduleLoader := func(path string) (*ast.Program, error) {
		// Resolve relative to the current source file being compiled?
		// CEmitter currently doesn't pass the "parent" path.
		// For now, try as is, but if relative, try relative to sourcePath.
		resolvedPath := path
		if !filepath.IsAbs(path) {
			resolvedPath = filepath.Join(filepath.Dir(sourcePath), path)
		}

		// Read Source
		content, err := os.ReadFile(resolvedPath)
		if err != nil {
			// Fallback: try raw path (for stdlib or project root relative)
			content, err = os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("reading module %s: %w", path, err)
			}
		}

		// Lex & Parse
		l := lexer.NewCustom(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			return nil, fmt.Errorf("parsing module %s errors:\n%s", path, strings.Join(p.Errors(), "\n"))
		}
		return program, nil
	}

	emitter := codegen.NewCEmitter(moduleLoader)
	cCode, err := emitter.Generate(program)
	if err != nil {
		return fmt.Errorf("codegen error: %w", err)
	}

	// 5. Write C file
	cPath := key + ".c"
	if err := os.WriteFile(cPath, []byte(cCode), 0644); err != nil {
		return fmt.Errorf("writing C file: %w", err)
	}

	// 6. Write Runtime Header
	headerPath := filepath.Join(filepath.Dir(cPath), "orglang.h")
	if err := os.WriteFile(headerPath, []byte(codegen.RuntimeHeader), 0644); err != nil {
		return fmt.Errorf("writing runtime header: %w", err)
	}

	// 7. Compile C (gcc)
	// Output binary name is `key` (e.g. `main` from `main.org`)
	// We link the C file. orglang.h is in the same dir, so -I. is implicit or standard.
	// If output file is absolute, we need to be careful with include paths.
	// simple: gcc -o output input.c

	buildCmd := exec.Command("gcc", "-o", key, cPath, "-lm")
	// Ensure GCC can find orglang.h if it's not in CWD?
	// If cPath is relative, it's fine.

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gcc error:\n%s", string(out))
	}

	return nil
}
