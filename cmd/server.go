package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/AikoPanel/Aiko-Server/panel"
)

var configFile string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Aiko-Server",
	Run:   serverHandle,
}

func init() {
	serverCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file for Aiko-Server.")
	command.AddCommand(serverCmd)
}

func serverHandle(_ *cobra.Command, _ []string) {
	showVersion()
	config := getConfig()
	panelConfig := &panel.Config{}
	if err := config.Unmarshal(panelConfig); err != nil {
		fmt.Printf("Parse config file %v failed: %s\n", configFile, err)
		return
	}
	p := panel.New(panelConfig)
	lastTime := time.Now()
	config.OnConfigChange(func(e fsnotify.Event) {
		if time.Now().After(lastTime.Add(3 * time.Second)) {
			fmt.Println("Config file changed:", e.Name)
			p.Close()
			runtime.GC()
			if err := config.Unmarshal(panelConfig); err != nil {
				fmt.Printf("Parse config file %v failed: %s\n", configFile, err)
				return
			}
			p.Start()
			lastTime = time.Now()
		}
	})
	p.Start()
	defer p.Close()

	runtime.GC()
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignals
	}
}

func getConfig() *viper.Viper {
	config := viper.New()

	if configFile != "" {
		configName := path.Base(configFile)
		configFileExt := path.Ext(configFile)
		configNameOnly := strings.TrimSuffix(configName, configFileExt)
		configPath := path.Dir(configFile)
		config.SetConfigName(configNameOnly)
		config.SetConfigType(strings.TrimPrefix(configFileExt, "."))
		config.AddConfigPath(configPath)
		os.Setenv("XRAY_LOCATION_ASSET", configPath)
		os.Setenv("XRAY_LOCATION_CONFIG", configPath)
	} else {
		config.SetConfigName("aiko")
		config.SetConfigType("yml")
		config.AddConfigPath(".")
	}

	if err := config.ReadInConfig(); err != nil {
		fmt.Printf("Config file error: %s\n", err)
	}

	config.WatchConfig()

	return config
}
