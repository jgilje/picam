//go:generate esc -o static.go -pkg main -prefix static static

package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
)

var port uint

// var homeTempl = template.Must(template.ParseFiles("home.html"))
var homeTmpl *template.Template

func indexHandler(c http.ResponseWriter, req *http.Request) {
	homeTmpl.Execute(c, req.Host)
}

var baseOptions = []string{"-o", "-", "-t", "100"}

func handleOptions(options []string, query url.Values) []string {
	if query["exposure"] != nil {
		options = append(options, "-ex", query["exposure"][0])
	}
	if query["awb"] != nil {
		options = append(options, "-awb", query["awb"][0])
	}
	if query["ifx"] != nil {
		options = append(options, "-ifx", query["ifx"][0])
	}
	if query["iso"] != nil {
		options = append(options, "-ISO", query["iso"][0])
	}
	if query["ss"] != nil {
		options = append(options, "-ss", query["ss"][0])
	}
	if (query["vf"] != nil) && (query["vf"][0] == "true") {
		options = append(options, "-vf")
	}
	if (query["hf"] != nil) && (query["hf"][0] == "true") {
		options = append(options, "-hf")
	}

	return options
}

func uncachedHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		handler(w, r)
	})
}

func fullHandler(c http.ResponseWriter, req *http.Request) {
	options := baseOptions
	options = handleOptions(options, req.URL.Query())

	cmd := exec.Command("raspistill", options...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Failed to execute preview")
	}

	c.Write(stdout)
}

func previewHandler(c http.ResponseWriter, req *http.Request) {
	options := append(baseOptions, "-w", "640", "-h", "480")
	options = handleOptions(options, req.URL.Query())

	cmd := exec.Command("raspistill", options...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Println("Failed to execute preview")
	}

	c.Write(stdout)
}

func readFile(fs http.FileSystem, name string) string {
	file, err := fs.Open(name)
	if err != nil {
		log.Fatal("Failed to open file")
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("Failed to read file", name)
	}

	return string(bytes)
}

func handleHTTP() {
	fs := FS(false)

	tmpl, err := template.New("home").Parse(readFile(fs, "/index.html"))
	if err != nil {
		log.Fatal("Template error in index.html")
	}
	homeTmpl = tmpl

	http.Handle("/", http.FileServer(fs))
	http.HandleFunc("/index.html", indexHandler)
	http.HandleFunc("/preview", uncachedHandler(previewHandler))
	http.HandleFunc("/full", uncachedHandler(fullHandler))

	// http.HandleFunc("/", homeHandler)
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
