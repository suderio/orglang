package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// Styles
	logoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	headerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true) // Blue accent
	subtextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // Dim gray
)

var rootCmd = &cobra.Command{
	Use:   "org",
	Short: "OrgLang Compiler",
	Long: logoStyle.Render("OrgLang") + ` - A dynamic, resource-oriented programming language.

Design: Distinct, yet Sober.`,
	// Silence usages on error to keep output clean
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be defined here
}

// Helper for printing section headers
func printHeader(title string) {
	fmt.Println(headerStyle.Render(title))
}

// Helper for printing info
func printInfo(label, value string) {
	fmt.Printf("%s: %s\n", subtextStyle.Render(label), value)
}
