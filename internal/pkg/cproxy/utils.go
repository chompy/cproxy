package cproxy

import (
	"log"
	"net"
	"net/http"
)

// GetListener - get listener for incomming requests
func GetListener(config *Config) (net.Listener, error) {
	// attempt tcp listener
	listener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		// attempt unix listener
		listener, err = net.Listen("unix", config.Listen)
	}
	return listener, err
}

// error500HTML - HTML for error 500 page
const error500HTML = `<!DOCTYPE html><html><head><title>Error 500</title><meta charset="UTF-8"/><style type="text/css"> html, body{font-family: sans-serif; text-align: center; margin-top: 40px;}h1{color: #000; font-size: 36px;}</style></head><body><h1>Error 500</h1></body></html>`

// RenderErrorPage - render an error page
func RenderErrorPage(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("ERROR ::", err.Error())
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(500)
	w.Write([]byte(error500HTML))
}
