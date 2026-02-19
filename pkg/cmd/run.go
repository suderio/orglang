package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	runArgs  string // This might need to be []string if capturing remaining args
	runDebug bool
)

var runCmd = &cobra.Command{
	Use:   "run [flags] <input> [args...]",
	Short: "Compile and execute an OrgLang program",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TBD: run command not implemented yet")
		if len(args) > 0 {
			fmt.Printf("Running %s...\n", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runArgs, "args", "a", "", "Pass arguments to the program")
	runCmd.Flags().BoolVar(&runDebug, "debug", false, "Run in debug mode")
}
