package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/xtls/xray-core/transport/internet/tls"
)

var certChainHashCmd = &cobra.Command{
	Use:   "certChainHash",
	Short: "Calculate TLS certificates hash.",
	Long: `
    Calculate TLS certificate chain hash.
    `,
	Run: executeCertChainHash,
}

var input string

func init() {
	command.AddCommand(certChainHashCmd)
	certChainHashCmd.Flags().StringVarP(&input, "cert", "c", "fullchain.pem", "The file path of the certificates chain")
	// You can add more flags if needed.
	// certChainHashCmd.Flags().StringVarP(&output, "output", "o", "output.txt", "Output file")
	// certChainHashCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	// ...

	// Add the certChainHashCmd to the root command if you have one.
	// rootCmd.AddCommand(certChainHashCmd)
}

func executeCertChainHash(cmd *cobra.Command, args []string) {
	certContent, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	certChainHashB64 := tls.CalculatePEMCertChainSHA256Hash(certContent)
	fmt.Println(certChainHashB64)
}

// You might also want to create a rootCmd if your application has more commands.
// var rootCmd = &cobra.Command{...}

// func Execute() {
//     if err := rootCmd.Execute(); err != nil {
//         fmt.Println(err)
//         os.Exit(1)
//     }
// }
