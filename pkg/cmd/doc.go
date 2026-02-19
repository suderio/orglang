package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var docCmd = &cobra.Command{
	Use:   "doc <input>",
	Short: "Generate documentation (TBD)",
	Long:  `Generates documentation from docstrings.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Doc"))
		printInfo("Input", args[0])
		printInfo("Status", "TBD - Documentation generator not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(docCmd)
	docCmd.Flags().Bool("html", false, "Output HTML")
	docCmd.Flags().Bool("json", false, "Output JSON")
}
