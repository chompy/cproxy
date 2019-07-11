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
func GetFCGIEnvVars(r *http.Request, config *Config) map[string]string {
	p := fcgi.ProcessEnv(r)
	if p == nil {
		p = map[string]string{}
	}
	p["REQUEST_METHOD"] = r.Method
	p["REQUEST_URI"] = r.URL.Path
	p["QUERY_STRING"] = r.URL.Query().Encode()
	p["CONTENT_LENGTH"] = strconv.FormatInt(r.ContentLength, 10)
	p["CONTENT_TYPE"] = r.Header.Get("Content-Type")
	p["HTTP_HOST"] = r.Host
	// TODO make configurable in plugin
	/*if config.UseESI {
		p["HTTP_SURROGATE_CAPABILITY"] = "content=ESI/1.0"
	}*/
	for k, values := range r.Header {
		k = "HTTP_" + strings.Replace(strings.ToUpper(k), "-", "_", -1)
		p[k] = values[0]
	}
	return p
}
