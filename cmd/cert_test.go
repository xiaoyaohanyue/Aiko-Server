// cert_test.go

package cmd

import (
	"testing"

	"github.com/xtls/xray-core/common/protocol/tls/cert" // Import the package that defines Certificate type
)

func TestExecuteCert(t *testing.T) {
	certCommand.Flags().Set("domain", "example.com")
	certCommand.Flags().Set("name", "Test Cert")
	certCommand.Flags().Set("org", "Test Org")
	certCommand.Flags().Set("ca", "false")
	certCommand.Flags().Set("file", "test_cert")
	certCommand.Flags().Set("expire", "24h")
	certCommand.Flags().Set("output", "/Users/aiko/Documents/GitHub/Aiko-Server/test/cert")

	executeCert(certCommand, nil)
}

func TestSaveCertificateAndKey(t *testing.T) {
	// Create a mock certificate for testing
	mockCert := &cert.Certificate{}

	err := saveCertificateAndKey(mockCert, "/Users/aiko/Documents/GitHub/Aiko-Server/test/cert/aiko-server.pem", "/Users/aiko/Documents/GitHub/Aiko-Server/test/cert/aiko-server.privkey")
	if err != nil {
		t.Errorf("saveCertificateAndKey failed with error: %v", err)
	}
}
