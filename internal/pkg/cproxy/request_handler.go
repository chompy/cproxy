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
	"fmt"
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
	if exts != nil {
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
	if exts != nil {
		for _, ext := range *exts {
			log.Println("REQUEST", requestNumber, ":: EVENT :: OnResponse ::", ext.Name)
			var err error
			resp, err = ext.OnResponse(resp)
			if err != nil {
				return nil, err
			}
			if resp == nil {
				return nil, fmt.Errorf("extension %s OnResponse returned nil response", ext.Name)
			}
			resp.Request = req
		}
	}

	log.Println("REQUEST", requestNumber, ":: Completed")
	return resp, nil

}
