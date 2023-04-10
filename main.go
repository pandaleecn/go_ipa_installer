package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"strings"
	"errors"
)

type Payload struct {
	Version 	string `json:"version"`
	Build		string `json:"build"`
	Url			string `json:"url"`
	Platform   	int `json:"platform"`
}

type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func main() {

	// 获取当前环境变量
	originalPath := os.Getenv("PATH")

	// 将 /usr/local/bin 和 /opt/homebrew/bin 路径添加到 PATH
	os.Setenv("PATH", "/usr/local/bin:/opt/homebrew/bin:"+originalPath)

	http.HandleFunc("/", handleRequest)
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {

	// Set CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    // Handle preflight requests
    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("2http ListenAndServe:9001")
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var payload Payload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		fmt.Println("3http ListenAndServe:9001")
		http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		return
	}

	version := payload.Version
	build := payload.Build
	platform := payload.Platform
	url := payload.Url

	if version == "" || build == "" || url == "" {
		http.Error(w, "Missing version or build", http.StatusBadRequest)
		return
	}

	// 获取 URL 的最后一个参数
    parts := strings.Split(url, "/")
	ipaFile := filepath.Join(os.TempDir(), fmt.Sprintf(parts[len(parts)-1]))

	// Download the IPA file
	err = downloadFile(ipaFile, url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading IPA file: %v", err), http.StatusInternalServerError)
		return
	}

	if platform == 1 {
		err = installApk(ipaFile)
	} else {
		err = installIpa(ipaFile)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf(err.Error()), http.StatusInternalServerError)
		return
	}
	responseSuccessToFront(w)
}

// 安装成功，返回给前端
func responseSuccessToFront(w http.ResponseWriter){
	// 处理安装请求
    response := Response{
        Code:    200,
        Message: fmt.Sprintf("App installed successfully"),
    }
    json, err := json.Marshal(response)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(response.Code)
    w.Write(json)
}

// 安装 Android 应用
func installApk(apkFile string)(errRes error) {
	cmd := exec.Command("adb", "install", apkFile)
	err := cmd.Run()
	if err != nil {
		errRes = errors.New(fmt.Sprintf("Failed to install APK: %v", err))
		return
	}
	return
}

// 安装 iOS 应用
func installIpa(ipaFile string)(errRes error) {
	// Get device ID using idevice_id
	deviceIDOutput, err := exec.Command("idevice_id", "-l").Output()
	if err != nil {
		errRes = errors.New(fmt.Sprintf("An error occurred: %s", err))
		return
	}

	deviceID := strings.TrimSpace(string(deviceIDOutput))
	if deviceID == "" {
		errRes = errors.New("No device found")
		return
	}

	// Install the app using ideviceinstaller
	cmd := exec.Command("ideviceinstaller", "-u", deviceID, "-i", ipaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing ideviceinstaller: %v", err)
		errRes = errors.New(fmt.Sprintf("Error installing app: %s", output))
	}
	return
}

// 下载安装包
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
