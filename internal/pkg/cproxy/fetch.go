package cproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"strconv"
	"strings"

	fcgiclient "github.com/alash3al/go-fastcgi-client"
)

// BackendFetch - fetch content from backend
func BackendFetch(r *http.Request, config *Config) (*http.Response, error) {
	switch config.ProxyType {
	case ProxyTypeHTTP:
		{
			return httpBackendFetch(r, config)
		}
	case ProxyTypeFCGI:
		{
			return fcgiBackendFetch(r, config)
		}
	}
	return nil, fmt.Errorf("no fetcher found for proxy type '%s'", config.ProxyType)
}

// httpBackendFetch - fetch content from http backend
func httpBackendFetch(r *http.Request, config *Config) (*http.Response, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Connect)
	if err != nil {
		return nil, err
	}
	r.Host = tcpAddr.String()
	httpConn := http.Client{}
	oResp, err := httpConn.Do(r)
	if err != nil {
		return nil, err
	}
	// read response
	// read now so network connection can be closed
	buf := bytes.NewBuffer(nil)
	bufW := bufio.NewWriter(buf)
	oResp.Write(bufW)
	bufW.Flush()
	// close the original response
	oResp.Body.Close()
	respBytes := buf.Bytes()
	resp, err := http.ReadResponse(
		bufio.NewReader(bytes.NewReader(respBytes)),
		r,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// fcgiBackendFetch - fetch content from fcgi backend
func fcgiBackendFetch(r *http.Request, config *Config) (*http.Response, error) {
	p := GetFCGIEnvVars(r, config)
	// open connection to backend
	fcgiConn, err := fcgiclient.Dial("tcp", config.Connect)
	if err != nil {
		fcgiConn, err = fcgiclient.Dial("unix", config.Connect)
		if err != nil {
			return nil, err
		}
	}
	defer fcgiConn.Close()
	// send request
	oResp, err := fcgiConn.Request(p, r.Body)
	if err != nil {
		return nil, err
	}
	// read response
	// read now so network connection can be closed
	buf := bytes.NewBuffer(nil)
	bufW := bufio.NewWriter(buf)
	oResp.Write(bufW)
	bufW.Flush()
	// close the original response
	oResp.Body.Close()
	respBytes := buf.Bytes()
	resp, err := http.ReadResponse(
		bufio.NewReader(bytes.NewReader(respBytes)),
		r,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetFCGIEnvVars - retrieve FCGI env vars
func GetFCGIEnvVars(req *http.Request, config *Config) map[string]string {
	p := fcgi.ProcessEnv(req)
	if p == nil {
		p = map[string]string{}
	}
	p["SERVER_SOFTWARE"] = "go"
	p["SERVER_NAME"] = req.Host
	p["SERVER_PROTOCOL"] = "HTTP/1.1"
	p["HTTP_HOST"] = req.Host
	p["GATEWAY_INTERFACE"] = "CGI/1.1"
	p["REQUEST_METHOD"] = req.Method
	p["QUERY_STRING"] = req.URL.RawQuery
	p["REQUEST_URI"] = req.URL.RequestURI()
	p["PATH_INFO"] = req.URL.Path
	//p["SCRIPT_NAME"] = "/"
	//p["SCRIPT_FILENAME"] = h.Path
	p["SERVER_PORT"] = req.URL.Port()
	p["CONTENT_LENGTH"] = strconv.FormatInt(req.ContentLength, 10)
	p["CONTENT_TYPE"] = req.Header.Get("Content-Type")
	for k, values := range req.Header {
		k = "HTTP_" + strings.Replace(strings.ToUpper(k), "-", "_", -1)
		p[k] = values[0]
	}
	return p
}
