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

package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"./internal/pkg/cproxy"
)

// getTestConfig - get config suitable for testing
func getTestConfig() cproxy.Config {
	config := cproxy.GetDefaultConfig()
	config.ProxyType = cproxy.ProxyTypeDummy
	return config
}

// TestRequest - test basic request, no extensions loaded
func TestRequest(t *testing.T) {

	// get config for testing
	config := getTestConfig()
	// create a new request
	req, err := http.NewRequest(
		http.MethodGet,
		"http://127.0.0.1/test",
		nil,
	)
	if err != nil {
		t.Errorf("Error while creating request, %s", err)
	}
	// handle the request, ensure no errors
	resp, err := cproxy.HandleRequest(req, &config, nil)
	if err != nil {
		t.Errorf("Error while handling request, %s", err)
	}
	// TEST: content-type text/plain
	if resp.Header.Get("Content-Type") != "text/plain" {
		t.Errorf("'Content-Type' response header was expected to be 'text/plain' go '%s' instead", resp.Header.Get("Content-Type"))
	}
	// convert response body in to bytes
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error while reading response body, %s", err)
	}
	bodyString := string(bodyBytes)
	// TEST: response body contains "PATH_INFO=/test"
	if !strings.Contains(bodyString, "PATH_INFO=/test") {
		t.Errorf("Response body was expected to contain string 'PATH_INFO=/test'")
	}
	// TEST: response body contains "SERVER_NAME=127.0.0.1"
	if !strings.Contains(bodyString, "SERVER_NAME=127.0.0.1") {
		t.Errorf("Response body was expected to contain string 'SERVER_NAME=127.0.0.1'")
	}

}

// TestRequestExtension - test request with a test extension
func TestRequestExtension(t *testing.T) {

	// get config for testing
	config := getTestConfig()
	// set response body value
	respBodyStr := "Testing response body"
	// create test ext
	ext := cproxy.Extension{
		Name: "CProxy-Test",
		OnUnload: func() {

		},
		OnRequest: func(req *http.Request) (*http.Response, error) {
			req.Header.Add("X-Test", "TESTING")
			return nil, nil
		},
		OnResponse: func(resp *http.Response) (*http.Response, error) {
			resp.Header.Add("X-Test", "TESTING")
			reader := bytes.NewReader([]byte(respBodyStr))
			resp.Body = ioutil.NopCloser(reader)
			return resp, nil
		},
	}
	// create a new request
	req, err := http.NewRequest(
		http.MethodGet,
		"http://127.0.0.1/test",
		nil,
	)
	if err != nil {
		t.Errorf("Error while creating request, %s", err)
	}
	// handle the request, ensure no errors
	resp, err := cproxy.HandleRequest(req, &config, &[]cproxy.Extension{ext})
	if err != nil {
		t.Errorf("Error while handling request, %s", err)
	}
	// TEST: x-test request header
	if req.Header.Get("X-Test") != "TESTING" {
		t.Errorf("'X-Test' request header was expected to be 'X-Test' go '%s' instead", req.Header.Get("X-Test"))
	}
	// TEST: x-test response header
	if resp.Header.Get("X-Test") != "TESTING" {
		t.Errorf("'X-Test' response header was expected to be 'X-Test' go '%s' instead", resp.Header.Get("X-Test"))
	}
	// convert response body in to bytes
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error while reading response body, %s", err)
	}
	bodyString := string(bodyBytes)
	// TEST: response body equals respBodyStr
	if bodyString != respBodyStr {
		t.Errorf("Response body was expected to be '%s' got '%s' instead", respBodyStr, bodyString)
	}

}
