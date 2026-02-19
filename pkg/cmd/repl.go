package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start interactive REPL (TBD)",
	Long:  `Starts an interactive Read-Eval-Print Loop for OrgLang.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("REPL"))
		printInfo("Status", "TBD - REPL logic not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(replCmd)
	replCmd.Flags().String("history", "", "Path to history file")
}
