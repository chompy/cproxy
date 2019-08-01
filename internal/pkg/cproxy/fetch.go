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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"strconv"
	"strings"

	fcgiclient "github.com/alash3al/go-fastcgi-client"
)

// BackendFetch - fetch content from backend
func BackendFetch(req *http.Request, config *Config) (*http.Response, error) {
	switch config.ProxyType {
	case ProxyTypeHTTP:
		{
			return httpBackendFetch(req, config)
		}
	case ProxyTypeFCGI:
		{
			return fcgiBackendFetch(req, config)
		}
	case ProxyTypeDummy:
		{
			return dummyBackendFetch(req, config)
		}
	}
	return nil, fmt.Errorf("no fetcher found for proxy type '%s'", config.ProxyType)
}

// httpBackendFetch - fetch content from http backend
func httpBackendFetch(req *http.Request, config *Config) (*http.Response, error) {
	connectURL, err := url.Parse(config.Backend)
	if err != nil {
		return nil, err
	}
	req.Host = connectURL.Host
	req.URL.Scheme = connectURL.Scheme
	req.URL.Host = connectURL.Host
	httpConn := http.Client{}
	req.RequestURI = ""
	oResp, err := httpConn.Do(req)
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
		req,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// fcgiBackendFetch - fetch content from fcgi backend
func fcgiBackendFetch(req *http.Request, config *Config) (*http.Response, error) {
	p := GetFCGIEnvVars(req, config)
	// open connection to backend
	fcgiConn, err := fcgiclient.Dial("tcp", config.Backend)
	if err != nil {
		fcgiConn, err = fcgiclient.Dial("unix", config.Backend)
		if err != nil {
			return nil, err
		}
	}
	defer fcgiConn.Close()
	// send request
	oResp, err := fcgiConn.Request(p, req.Body)
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
		req,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// dummyBackendFetch - dummy fetch function used for testing
func dummyBackendFetch(req *http.Request, config *Config) (*http.Response, error) {
	p := GetFCGIEnvVars(req, config)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Request:    req,
		Header:     make(http.Header, 0),
	}
	resp.Header.Add("Content-Type", "text/plain")
	// create a body of FCGI vars
	bodyBytes := make([]byte, 0)
	for key, val := range p {
		bodyBytes = append(
			bodyBytes,
			[]byte(key+"="+val+"\n")...,
		)
	}
	reader := bytes.NewReader(bodyBytes)
	resp.Body = ioutil.NopCloser(reader)
	resp.ContentLength = int64(len(bodyBytes))
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
