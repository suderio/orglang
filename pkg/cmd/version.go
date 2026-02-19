package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	versionShort bool
	versionJSON  bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of OrgLang",
	Run: func(cmd *cobra.Command, args []string) {
		if versionJSON {
			fmt.Println(`{"version": "0.1.0-alpha"}`)
			return
		}
		if versionShort {
			fmt.Println("0.1.0-alpha")
			return
		}
		fmt.Println("OrgLang version 0.1.0-alpha")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "Print only the version number")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Print version info in JSON format")
}
