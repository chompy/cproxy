package cproxy

import (
	"log"
	"net/http"
)

// HandleRequest - handle a request
func HandleRequest(req *http.Request, config *Config, exts *[]Extension) (*http.Response, error) {

	log.Println("REQUEST ::", req.Method, req.URL.String())

	// call 'OnRequest'
	log.Println("EVENT :: OnRequest")
	var resp *http.Response
	for _, ext := range *exts {
		var err error
		resp, err = ext.OnRequest(req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			break
		}
	}

	// backend fetch, only if response is nil
	if resp == nil {
		var err error
		resp, err = BackendFetch(req, config)
		if err != nil {
			return nil, err
		}
	}

	// call 'OnResponse'
	log.Println("EVENT :: OnResponse")
	for _, ext := range *exts {
		var err error
		resp, err = ext.OnResponse(resp)
		resp.Request = req
		if err != nil {
			return nil, err
		}
	}

	return resp, nil

}
