package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var createConfigCmd = &cobra.Command{
	Use:   "createconfig",
	Short: "Create a configuration file from a remote URL and save it to /etc/Aiko-Server/",
	Run:   createConfig,
}

var configFileType string // Flag to specify the config file type (yaml, yml, json)

func init() {
	createConfigCmd.Flags().StringVarP(&configFileType, "type", "t", "yml", "Config file type (yaml, yml, json)")
	command.AddCommand(createConfigCmd)
}

func createConfig(_ *cobra.Command, _ []string) {
	configURL := "https://raw.githubusercontent.com/AikoPanel/Aiko-Server/master/Aiko-Server/config/aiko." + strings.ToLower(configFileType) + ".example"
	response, err := http.Get(configURL)
	if err != nil {
		fmt.Printf("Error downloading configuration from URL: %s\n", err)
		os.Exit(1)
	}
	defer response.Body.Close()

	configData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading configuration data: %s\n", err)
		os.Exit(1)
	}

	configFileName := "/etc/Aiko-Server/aiko." + strings.ToLower(configFileType)
	err = ioutil.WriteFile(configFileName, configData, 0644)
	if err != nil {
		fmt.Printf("Error creating configuration file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration file aiko.%s created successfully, Saved to %s.\n", strings.ToLower(configFileType), configFileName)
}
