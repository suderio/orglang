package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [files...]",
	Short: "Format source code (TBD)",
	Long:  `Formats OrgLang source files to standard style.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(headerStyle.Render("Format"))
		printInfo("Status", "TBD - Formatter not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().BoolP("write", "w", false, "Write result to file")
	fmtCmd.Flags().Bool("check", false, "Check if file is formatted")
}
