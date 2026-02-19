package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildOutput   string
	buildTarget   string
	buildOptimize int
	buildStatic   bool
	buildDebug    bool
	buildVerbose  bool
)

var buildCmd = &cobra.Command{
	Use:   "build [flags] <input>",
	Short: "Compile OrgLang source code",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TBD: build command not implemented yet")
		if buildVerbose {
			fmt.Printf("Building %s...\n", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "Output file name")
	buildCmd.Flags().StringVarP(&buildTarget, "target", "t", "", "Target architecture (e.g., linux/amd64)")
	buildCmd.Flags().IntVarP(&buildOptimize, "optimize", "O", 1, "Optimization level (0-3)")
	buildCmd.Flags().BoolVar(&buildStatic, "static", false, "Link statically")
	buildCmd.Flags().BoolVar(&buildDebug, "debug", false, "Include debug information")
	buildCmd.Flags().BoolVarP(&buildVerbose, "verbose", "v", false, "Verbose output during compilation")
}
