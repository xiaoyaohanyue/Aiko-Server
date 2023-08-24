package cmd

import (
	"testing"
)

func Test_Version(t *testing.T) {
	version = "AikoCute Version Test"
	showVersion()
}
