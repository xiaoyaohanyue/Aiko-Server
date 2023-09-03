package cmd

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func TestCreateConfig(t *testing.T) {
	// Tạo một mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Trả về dữ liệu mẫu cho yêu cầu HTTP
		w.Write([]byte("sample_config_data"))
	}))
	defer server.Close()

	// Lưu lại giá trị biến môi trường ban đầu của $HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Thiết lập biến môi trường $HOME để đảm bảo /etc/Aiko-Server/ là một thư mục hợp lệ
	os.Setenv("HOME", "/")

	// Thiết lập giá trị configFileType cho kiểm thử
	configFileType = "yml"

	// Tạo một cobra.Command giả để truyền vào hàm createConfig
	cmd := &cobra.Command{}

	// Gọi hàm createConfig với các giá trị giả
	createConfig(cmd, []string{})

	// Kiểm tra xem tệp cấu hình đã được tạo thành công
	configFileName := "./aiko.yml"
	_, err := os.Stat(configFileName)
	if err != nil {
		t.Errorf("Expected configuration file to be created, but got error: %v", err)
	}

	// Kiểm tra nội dung của tệp cấu hình
	configData, err := ioutil.ReadFile(configFileName)
	if err != nil {
		t.Errorf("Error reading configuration file: %v", err)
	}

	expectedConfigData := "sample_config_data"
	if string(configData) != expectedConfigData {
		t.Errorf("Expected configuration data to be %s, but got %s", expectedConfigData, string(configData))
	}
}
