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
var configFormat string // Thêm biến để lưu định dạng cấu hình

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Aiko-Server",
	Run:   serverHandle,
}

func init() {
	// serverCmd.Flags().StringVarP(&configFile, "config", "c", "/etc/Aiko-Server/aiko.yml", "Custom configuration file path.")
	serverCmd.Flags().StringVarP(&configFormat, "format", "f", "yml", "Configuration file format (json or yml).")
	command.AddCommand(serverCmd)
}

func serverHandle(_ *cobra.Command, _ []string) {
	showVersion()
	// Kiểm tra xem người dùng đã chỉ định file cấu hình hay không
	if configFile == "" {
		// Nếu không, sử dụng đường dẫn mặc định dựa trên định dạng được chỉ định
		configFile = "/etc/Aiko-Server/aiko." + configFormat
	}
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
		// test with default config file name
		config.SetConfigName("aiko")
		config.SetConfigType(configFormat)
		config.AddConfigPath(".") // look for config in the working directory
	}

	if err := config.ReadInConfig(); err != nil {
		fmt.Printf("Config file error: %s\n", err)
	}

	config.WatchConfig()

	return config
}
