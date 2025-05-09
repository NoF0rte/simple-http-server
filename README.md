# Simple HTTP Server
This is a simple HTTP server that tries to act like python's HTTP server and be slightly more configurable.

## Install
Use go to install the binary
```
go install github.com/NoF0rte/simple-http-server@latest
```

## Usage
```
Usage of simple-http-server:
  -cors
        Enable CORS
  -d string
        The directory where the files are served (default ".")
  -l string
        The address to listen on (default "0.0.0.0")
  -p string
        The port to listen on (default "8000")
  -redirect
        Enable dynamic redirect.
        Format: /redir/<required_method>/<base64_redirect>
                /redir/<required_method>/<status_code>/<base64_redirect>
                /redir?method=<required_method>&status=<status_code>&redir=<redirect>
    
        required_method: The method required to activate the redirect. Use * for any method.
        status_code: The desired redirect status code to use. Must be in the range of 300-399. Status code 307 is the default.
    
        Examples: https://localhost:8000/redir/POST/aHR0cHM6Ly9nb29nbGUuY29t // Redirects POST requests to https://google.com
                  https://localhost:8000/redir/*/303/aHR0cHM6Ly9nb29nbGUuY29t // Redirects any request to https://google.com using the 303 status code
                  https://localhost:8000/redir?method=*&status=302&redir=https://google.com // Redirects any request to https://google.com using the 302 status code
  -verbose
        Enable verbose logging. Logs out the request headers and body

```