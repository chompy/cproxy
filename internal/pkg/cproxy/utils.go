/*
This file is part of CProxy.

CProxy is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

CProxy is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with CProxy.  If not, see <https://www.gnu.org/licenses/>.
*/

package cproxy

import (
	"bufio"
	"bytes"
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

// HTTPResponseToBytes - convert http response to bytes
func HTTPResponseToBytes(r *http.Response) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	bufW := bufio.NewWriter(buf)
	err := r.Write(bufW)
	if err != nil {
		return nil, err
	}
	bufW.Flush()
	r.Body.Close()
	return buf.Bytes(), nil
}

// HTTPResponseFromBytes - convert bytes to http response
func HTTPResponseFromBytes(b []byte) (*http.Response, error) {
	return http.ReadResponse(
		bufio.NewReader(bytes.NewReader(b)),
		nil,
	)
}
