package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] <input> [args...]",
	Short: "Compile and execute OrgLang program (TBD)",
	Long:  `Compiles the OrgLang program and executes it immediately.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		progArgs := args[1:]

		fmt.Println(headerStyle.Render("Run"))
		printInfo("Input", input)
		if len(progArgs) > 0 {
			printInfo("Args", strings.Join(progArgs, " "))
		}
		printInfo("Status", "TBD - Run logic not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringSliceP("args", "a", []string{}, "Arguments to pass to the program")
}
