package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	testVerbose  bool
	testFilter   string
	testCoverage bool
)

var testCmd = &cobra.Command{
	Use:   "test [flags] [files...]",
	Short: "Run tests defined in OrgLang files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TBD: test command not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Verbose test output")
	testCmd.Flags().StringVar(&testFilter, "filter", "", "Run only tests matching regex")
	testCmd.Flags().BoolVar(&testCoverage, "coverage", false, "Generate coverage report")
}
