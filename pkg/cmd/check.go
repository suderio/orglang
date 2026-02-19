package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:     "check <input>",
	Short:   "Static analysis (TBD)",
	Long:    `Performs static analysis without compiling/running.`,
	Aliases: []string{"vet"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Check"))
		printInfo("Input", args[0])
		printInfo("Status", "TBD - Static analysis not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
