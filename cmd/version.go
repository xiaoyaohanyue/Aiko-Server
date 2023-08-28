package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	version  = "AikoCuteHotMe" //use ldflags replace
	codename = "Aiko-Server"
	intro    = "A backend based on multi Panel"
)

func getIPAddress() (string, error) {
	resp, err := http.Get("http://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ipBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ipBytes), nil
}

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
	ipAddr, err := getIPAddress()
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}
	fmt.Printf("IP Server: %s\n", ipAddr)
}
