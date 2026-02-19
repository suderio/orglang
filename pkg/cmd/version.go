package cmd

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	Version   = "v0.1.0-dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of OrgLang",
	Long:  `All software has versions. This is OrgLang's.`,
	Run: func(cmd *cobra.Command, args []string) {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")) // Cyan
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // Grey

		fmt.Println(titleStyle.Render("OrgLang Compiler"))
		fmt.Printf("  %s %s\n", labelStyle.Render("Version:"), Version)
		fmt.Printf("  %s %s\n", labelStyle.Render("Go Version:"), runtime.Version())
		fmt.Printf("  %s %s/%s\n", labelStyle.Render("OS/Arch:"), runtime.GOOS, runtime.GOARCH)
		if Commit != "unknown" {
			fmt.Printf("  %s %s\n", labelStyle.Render("Commit:"), Commit)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
