package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/curve25519"
)

var x25519Command = cobra.Command{
	Use:   "x25519",
	Short: "Generate key pair for x25519 key exchange",
	Run: func(cmd *cobra.Command, args []string) {
		executeX25519()
	},
}

func init() {
	command.AddCommand(&x25519Command)
	x25519Command.Flags().StringVarP(&input_base64, "i", "", "", "Input private key in base64 format")
}

var input_base64 string

func executeX25519() {
	var output string
	var err error
	var privateKey []byte
	var publicKey []byte

	if len(input_base64) > 0 {
		privateKey, err = base64.RawURLEncoding.DecodeString(input_base64)
		if err != nil {
			output = err.Error()
			goto out
		}
		if len(privateKey) != curve25519.ScalarSize {
			output = "Invalid length of private key."
			goto out
		}
	} else {
		privateKey = make([]byte, curve25519.ScalarSize)
		_, err = rand.Read(privateKey)
		if err != nil {
			output = err.Error()
			goto out
		}

		// Modify random bytes using algorithm described at:
		// https://cr.yp.to/ecdh.html.
		privateKey[0] &= 248
		privateKey[31] &= 127
		privateKey[31] |= 64
	}

	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		output = err.Error()
		goto out
	}

	output = fmt.Sprintf("Created by: AikoCute\nPrivate key: %v\nPublic key: %v",
		base64.RawURLEncoding.EncodeToString(privateKey),
		base64.RawURLEncoding.EncodeToString(publicKey))

out:
	fmt.Println(output)
}
