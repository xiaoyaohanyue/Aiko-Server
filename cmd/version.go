package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version  = "AikoCuteHotMe" //use ldflags replace
	codename = "Aiko-Server"
	intro    = "A backend based on multi Panel"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of Aiko-Server",
	Run: func(_ *cobra.Command, _ []string) {
		showVersion()
	},
}

func init() {
	command.AddCommand(versionCmd)
}

func showVersion() {
	fmt.Printf("%s %s (%s)\n", codename, version, intro)
}
