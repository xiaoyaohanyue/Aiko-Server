package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
)

var ip string

type IPInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Location string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Timezone string `json:"timezone"`
}

var cmdVPSInfo = &cobra.Command{
	Use:   "getVPSInfo",
	Short: "Get VPS information",
	Run: func(cmd *cobra.Command, args []string) {
		getVPSInfo()
	},
}

func init() {
	command.AddCommand(cmdVPSInfo)
	cmdVPSInfo.Flags().StringVarP(&ip, "ip", "i", "", "VPS IP address")
}

func getVPSInfo() {
	if ip == "" {
		ipAddr, err := getIPAddress()
		if err != nil {
			fmt.Println("Lỗi:", err)
			return
		}
		ip = ipAddr
	}

	url := fmt.Sprintf("https://ipinfo.io/%s/json", ip)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var info IPInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	// Lấy thông tin về RAM
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Lỗi khi lấy thông tin về RAM:", err)
		return
	}

	fmt.Println("-- Aiko-Server --")
	fmt.Printf("IPv4: %s\n", info.IP)
	fmt.Printf("Location: %s | %s | %s\n", info.City, info.Region, info.Country)
	fmt.Printf("Provider: %s\n", info.Org)
	if len(cpuInfo) > 0 {
		cpuName := cpuInfo[0].ModelName
		fmt.Println("CPU:", cpuName)
	}
	if memInfo != nil {
		fmt.Printf("RAM: %v GB\n", memInfo.Total/1024/1024/1024)
	}
	osName := runtime.GOOS
	fmt.Println("Hệ điều hành:", osName)
	// fmt.Printf("-- System: %s\n", "Ubuntu 22.04.2 LT") // Thay đổi thông tin hệ điều hành của bạn tại đây
}
