package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [flags] [files...]",
	Short: "Run tests (TBD)",
	Long:  `Runs tests defined in OrgLang files.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Test"))
		printInfo("Status", "TBD - Test framework not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
	testCmd.Flags().Bool("coverage", false, "Generate coverage report")
}
