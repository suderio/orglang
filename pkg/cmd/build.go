package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [flags] <input>",
	Short: "Compile OrgLang source code (TBD)",
	Long:  `Compiles OrgLang source code into an executable or bytecode.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Build"))
		printInfo("Input", args[0])
		printInfo("Status", "TBD - Build logic not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	// Add flags here
	buildCmd.Flags().StringP("output", "o", "", "Output file name")
	buildCmd.Flags().StringP("target", "t", "", "Target architecture (future)")
	buildCmd.Flags().IntP("optimize", "O", 1, "Optimization level")
}
