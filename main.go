package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"sync"
)

var mutex sync.Mutex

var port uint

var baseOptions = []string{"-n", "-v", "0", "-o", "-", "-t", "1000"}

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

	options := append(baseOptions, "--width", "640", "--height", "480")
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

	addr := ":" + strconv.FormatUint(uint64(port), 10)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func main() {
	flag.UintVar(&port, "port", 9000, "http service address")
	flag.Parse()

	go handleHTTP()

	for {
		select {}
	}
}
