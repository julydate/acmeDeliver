package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	appName  = "acmeDeliver"
	describe = "An acme.sh certificate distribution service"
	version  = "2.0.0"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s %s. %s", appName, version, describe)
	},
}
