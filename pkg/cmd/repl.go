package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	replNoBanner bool
	replHistory  string
)

var replCmd = &cobra.Command{
	Use:   "repl [flags]",
	Short: "Start an interactive Read-Eval-Print Loop",
	Run: func(cmd *cobra.Command, args []string) {
		if !replNoBanner {
			fmt.Println("Welcome to OrgLang REPL (v0.1.0-alpha)")
			fmt.Println("Type 'exit' or Ctrl+D to quit.")
		}
		fmt.Println("TBD: REPL not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(replCmd)

	replCmd.Flags().BoolVar(&replNoBanner, "no-banner", false, "Hide welcome message")
	replCmd.Flags().StringVar(&replHistory, "history", "", "Path to history file")
}
