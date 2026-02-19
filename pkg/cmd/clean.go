package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove build artifacts (TBD)",
	Long:  `Removes build artifacts.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Clean"))
		printInfo("Status", "TBD - Clean logic not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
