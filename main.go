package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/rs/cors"
)

var port string
var listenAddress string
var redirect bool
var verbose bool
var directory string
var enableCors bool

func main() {
	flag.StringVar(&port, "p", "8000", "The port to listen on")
	flag.StringVar(&listenAddress, "l", "0.0.0.0", "The address to listen on")
	flag.StringVar(&directory, "d", ".", "The directory where the files are served")
	flag.BoolVar(&enableCors, "cors", false, "Enable CORS")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging. Logs out the request headers and body")
	flag.BoolVar(&redirect, "redirect", false, "Enable dynamic redirect. Format: /redir/<required_method>/<base64_redirect> or /redir?method=<required_method>&redir=<redirect>")
	flag.Parse()

	listenAddress := fmt.Sprintf("%s:%s", listenAddress, port)
	fmt.Printf("[+] Starting HTTP server on %s\n", listenAddress)

	var handler http.Handler
	if redirect {
		log.Println("Setting up the redirection handler.")
		http.Handle("/redir/", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(redirHandler)))
	}

	fmt.Printf("[+] Serving files at: %s\n", directory)
	handler = handlers.CombinedLoggingHandler(os.Stdout, http.FileServer(http.Dir(directory)))

	if enableCors {
		handler = cors.AllowAll().Handler(handler)
	}

	if verbose {
		handler = verboseHandler(handler)
	}

	http.Handle("/", handler)

	http.ListenAndServe(listenAddress, nil)
}

func verboseHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("-----------------------------")
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		handler.ServeHTTP(w, r)

		var headers []string
		for header := range r.Header {
			headers = append(headers, header)
		}

		sort.Strings(headers)

		for _, header := range headers {
			fmt.Printf("%s: %s\n", header, r.Header.Get(header))
		}

		if len(body) > 0 {
			fmt.Println()
			fmt.Println(string(body))
		}
		fmt.Println("-----------------------------")
	})
}

func redirHandler(w http.ResponseWriter, r *http.Request) {
	trimmed := r.URL.Path[len("/redir/"):]

	method := ""
	redirect := ""
	if r.URL.Query().Has("redir") {
		method = r.URL.Query().Get("method")
		redirect = r.URL.Query().Get("redir")
	} else {
		var encoded string
		method, encoded, _ = strings.Cut(trimmed, "/")
		encoded, _, _ = strings.Cut(encoded, "/")

		bytes, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			log.Printf("Error base64 decoding %s: %v", encoded, err)
			http.NotFound(w, r)
			return
		}

		redirect = string(bytes)
	}

	if method != "*" && !strings.EqualFold(method, r.Method) {
		log.Printf("Redirect hit but method %s doesn't match expected %s", r.Method, method)
		fmt.Fprintln(w, "Success")
		return
	}

	log.Printf("[+] Redirecting %s to %s", r.RemoteAddr, redirect)
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}
