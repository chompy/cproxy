package cproxy

import (
	"log"
	"net/http"
)

// requestCount - request counter
var requestCount = 0

// HandleRequest - handle a request
func HandleRequest(req *http.Request, config *Config, exts *[]Extension) (*http.Response, error) {

	// increment request count
	requestCount++
	requestNumber := requestCount

	// output to log
	log.Println("REQUEST", requestNumber, "::", req.Method, req.URL.String())

	// call 'OnRequest'
	var resp *http.Response
	for _, ext := range *exts {
		log.Println("REQUEST", requestNumber, ":: EVENT :: OnRequest ::", ext.Name)
		var err error
		resp, err = ext.OnRequest(req)
		if err != nil {
			return nil, err
		}
		if resp != nil {
			// if response returned then assume it is a cached
			// response and no further manipulation is needed
			log.Println("REQUEST", requestNumber, ":: Completed")
			return resp, nil
		}
	}

	// backend fetch, only if response is nil
	if resp == nil {
		log.Println("REQUEST", requestNumber, ":: Backend fetch")
		var err error
		resp, err = BackendFetch(req, config)
		if err != nil {
			return nil, err
		}
	}

	// call 'OnResponse'
	for _, ext := range *exts {
		log.Println("REQUEST", requestNumber, ":: EVENT :: OnResponse ::", ext.Name)
		var err error
		resp, err = ext.OnResponse(resp)
		if err != nil {
			return nil, err
		}
		resp.Request = req
	}

	log.Println("REQUEST", requestNumber, ":: Completed")
	return resp, nil

}
