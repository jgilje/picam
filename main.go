package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type Config struct {
	Port                 uint   `json:"port"`
	PreviewWidth         string `json:"preview-width"`
	PreviewHeight        string `json:"preview-height"`
	ExposureCompensation string `json:"exposure-compensation"`
}

type OptionalConfig struct {
	Port                 *uint   `json:"port"`
	PreviewWidth         *string `json:"preview-width"`
	PreviewHeight        *string `json:"preview-height"`
	ExposureCompensation *string `json:"exposure-compensation"`
}

var config Config = Config{
	Port:                 9000,
	PreviewWidth:         "1024",
	PreviewHeight:        "768",
	ExposureCompensation: "0",
}

var mutex sync.Mutex

var baseOptions = []string{"-n", "-v", "0", "-o", "-", "-t", "1000"}

func loadConfig(filename string) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var optionalConfig OptionalConfig

	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	if optionalConfig.Port != nil {
		config.Port = *optionalConfig.Port
	}
	if optionalConfig.PreviewWidth != nil {
		config.PreviewWidth = *optionalConfig.PreviewWidth
	}
	if optionalConfig.PreviewHeight != nil {
		config.PreviewHeight = *optionalConfig.PreviewHeight
	}
	if optionalConfig.ExposureCompensation != nil {
		config.ExposureCompensation = *optionalConfig.ExposureCompensation
	}
}

func handleOptions(options []string, query url.Values) []string {
	if query["ss"] != nil {
		options = append(options, "--shutter", query["ss"][0])
	}
	if (query["vf"] != nil) && (query["vf"][0] == "true") {
		options = append(options, "--vflip=1")
	}
	if (query["hf"] != nil) && (query["hf"][0] == "true") {
		options = append(options, "--hflip=1")
	}

	return options
}

func uncachedHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		handler(w, r)
	})
}

func fullHandler(c http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	options := baseOptions
	options = handleOptions(options, req.URL.Query())

	cmd := exec.Command("libcamera-still", options...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Failed to execute preview")
	}

	c.Write(stdout)
}

func previewHandler(c http.ResponseWriter, req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	options := append(baseOptions, "--width", config.PreviewWidth, "--height", config.PreviewHeight)
	if config.ExposureCompensation != "" {
		options = append(options, "--ev", config.ExposureCompensation)
	}
	options = handleOptions(options, req.URL.Query())

	cmd := exec.Command("libcamera-still", options...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Failed to execute preview")
	}

	c.Write(stdout)
}

func handleHTTP() {
	http.HandleFunc("/preview", uncachedHandler(previewHandler))
	http.HandleFunc("/full", uncachedHandler(fullHandler))

	addr := ":" + strconv.FormatUint(uint64(config.Port), 10)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func main() {
	configFile := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	loadConfig(*configFile)

	go handleHTTP()

	for {
		select {}
	}
}
