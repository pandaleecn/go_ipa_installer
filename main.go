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
)

type Payload struct {
	Version string `json:"version"`
	Build   string `json:"build"`
}

func main() {
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

	fmt.Println("payload.Version:", payload.Version)
	fmt.Println("payload.Build:", payload.Build)

	if version == "" || build == "" {
		http.Error(w, "Missing version or build", http.StatusBadRequest)
		return
	}

	ipaURL := fmt.Sprintf("http://172.16.0.94:9000/static/ipa/HelloTalk_Binary_%s_%s.ipa", version, build)
	ipaFile := filepath.Join(os.TempDir(), fmt.Sprintf("HelloTalk_Binary_%s_%s.ipa", version, build))

	// Download the IPA file
	err = downloadFile(ipaFile, ipaURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading IPA file: %v", err), http.StatusInternalServerError)
		return
	}

	// Get device ID using idevice_id
	cmd := createCommandWithEnv("idevice_id", "-l")
	deviceIDOutput, err := cmd.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting device ID: %v", err), http.StatusInternalServerError)
		return
	}

	deviceID := strings.TrimSpace(string(deviceIDOutput))
	if deviceID == "" {
		http.Error(w, "No device found", http.StatusInternalServerError)
		return
	}

	// Install the app using ideviceinstaller
	cmd = createCommandWithEnv("ideviceinstaller", "-u", deviceID, "-i", ipaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing ideviceinstaller: %v", err)
		http.Error(w, fmt.Sprintf("Error installing app: %s", output), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "App installed successfully: %s", output)
}

func createCommandWithEnv(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/bin:/opt/homebrew/bin:"+os.Getenv("PATH"))
	return cmd
}

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
