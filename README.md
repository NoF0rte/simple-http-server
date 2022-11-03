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

```