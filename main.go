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
	"strconv"
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
	flag.BoolVar(&redirect, "redirect", false, `Enable dynamic redirect.
Format: /redir/<required_method>/<base64_redirect>
	/redir/<required_method>/<status_code>/<base64_redirect>
	/redir?method=<required_method>&status=<status_code>&redir=<redirect>

required_method: The method required to activate the redirect. Use * for any method.
status_code: The desired redirect status code to use. Must be in the range of 300-399. Status code 307 is the default.

Examples: https://localhost:8000/redir/POST/aHR0cHM6Ly9nb29nbGUuY29t // Redirects POST requests to https://google.com
	  https://localhost:8000/redir/*/303/aHR0cHM6Ly9nb29nbGUuY29t // Redirects any request to https://google.com using the 303 status code
	  https://localhost:8000/redir?method=*&status=302&redir=https://google.com // Redirects any request to https://google.com using the 302 status code`)
	flag.Parse()

	listenAddress := fmt.Sprintf("%s:%s", listenAddress, port)
	fmt.Printf("[+] Starting HTTP server on %s\n", listenAddress)

	var handler http.Handler
	if redirect {
		log.Println("Setting up the redirection handler.")
		http.Handle("/redir", handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(redirHandler)))
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
	trimmed := r.URL.Path[len("/redir"):]
	trimmed = strings.TrimPrefix(trimmed, "/")

	status := http.StatusTemporaryRedirect
	method := ""
	redirect := ""
	if r.URL.Query().Has("redir") {
		method = r.URL.Query().Get("method")
		redirect = r.URL.Query().Get("redir")

		s := r.URL.Query().Get("status")
		if s != "" {
			i, err := strconv.Atoi(s)
			if err == nil && i >= 300 && i <= 399 { // Only use the status code if it is in the 300's
				status = i
			}
		}
	} else {
		var statusStr string
		var parts string
		var encoded string
		method, parts, _ = strings.Cut(trimmed, "/")
		statusStr, encoded, _ = strings.Cut(parts, "/")

		if encoded == "" { // This means they didn't specify a status code to use
			encoded = statusStr
			statusStr = ""
		} else {
			encoded, _, _ = strings.Cut(encoded, "/")
		}

		bytes, err := base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			log.Printf("Error base64 decoding %s: %v", encoded, err)
			http.NotFound(w, r)
			return
		}

		redirect = string(bytes)

		if statusStr != "" {
			i, err := strconv.Atoi(statusStr)
			if err == nil && i >= 300 && i <= 399 { // Only use the status code if it is in the 300's
				status = i
			}
		}
	}

	if method != "*" && !strings.EqualFold(method, r.Method) {
		log.Printf("Redirect hit but method %s doesn't match expected %s", r.Method, method)
		fmt.Fprintln(w, "Success")
		return
	}

	log.Printf("[+] Redirecting %s to %s", r.RemoteAddr, redirect)
	http.Redirect(w, r, redirect, status)
}
