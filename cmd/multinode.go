package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func checkAndInstallSocat() error {
	cmd := exec.Command("which", "socat")
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		fmt.Println("Multinode is not installed. Installing...")

		// Kiểm tra hệ thống và sử dụng trình quản lý gói phù hợp
		var installCmd *exec.Cmd
		switch detectOS() {
		case "ubuntu":
			installCmd = exec.Command("apt-get", "install", "-y", "socat")
		case "centos":
			installCmd = exec.Command("yum", "install", "-y", "socat")
		default:
			fmt.Println("Unsupported OS")
			return nil
		}

		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return err
		}
		fmt.Println("Multinode has been installed.")
	} else {
		fmt.Println("Multinode is already installed.")
	}
	return nil
}

func detectOS() string {
	cmd := exec.Command("bash", "-c", "cat /etc/os-release")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error detecting OS:", err)
		return ""
	}

	osInfo := string(output)
	if strings.Contains(osInfo, "Ubuntu") {
		return "ubuntu"
	} else if strings.Contains(osInfo, "CentOS") {
		return "centos"
	}

	fmt.Println("Unsupported OS:", osInfo)
	return ""
}

var sourcePort string
var destinationPort string

var cmdNode = &cobra.Command{
	Use:   "multinode",
	Short: "Run multinode command for port mapping",
	Run:   runSocat,
}

func init() {
	cmdNode.Flags().StringVarP(&sourcePort, "source", "s", "", "Source port")
	cmdNode.Flags().StringVarP(&destinationPort, "destination", "d", "", "Destination port")
	command.AddCommand(cmdNode)
}

func runSocat(cmd *cobra.Command, args []string) {
	if err := checkAndInstallSocat(); err != nil {
		fmt.Printf("Error installing socat: %v\n", err)
		return
	}

	if sourcePort == "" || destinationPort == "" {
		fmt.Println("Both source and destination ports must be specified")
		return
	}

	cmdString := fmt.Sprintf("socat TCP-LISTEN:%s,fork TCP:localhost:%s", destinationPort, sourcePort)
	fmt.Printf("Running command: %s\n", cmdString)

	cmdExec := exec.Command("sh", "-c", cmdString)
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr

	if err := cmdExec.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
