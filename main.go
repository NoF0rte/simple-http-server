package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/rs/cors"
)

var port string
var listenAddress string
var directory string
var enableCors bool

func main() {
	flag.StringVar(&port, "p", "8000", "The port to listen on")
	flag.StringVar(&listenAddress, "l", "0.0.0.0", "The address to listen on")
	flag.StringVar(&directory, "d", ".", "The directory where the files are served")
	flag.BoolVar(&enableCors, "cors", false, "Enable CORS")
	flag.Parse()

	listenAddress := fmt.Sprintf("%s:%s", listenAddress, port)
	fmt.Printf("[+] Starting HTTP server on %s\n", listenAddress)
	fmt.Printf("[+] Serving files at: %s\n", directory)

	handler := handlers.CombinedLoggingHandler(os.Stdout, http.FileServer(http.Dir(directory)))

	if enableCors {
		handler = cors.Default().Handler(handler)
	}

	http.ListenAndServe(listenAddress, handler)
}
