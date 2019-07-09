package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strings"

	"./internal/pkg/cproxy"
)

// HandleRequest - handle a request
func HandleRequest(r *http.Request, config *cproxy.Config, exts []cproxy.Extension) (*http.Response, error) {

	log.Println("REQUEST ::", r.Method, r.URL.String())

	// call 'OnRequest'
	log.Println("EVENT :: OnRequest")
	var resp *http.Response
	for index := range exts {
		err := exts[index].OnRequest(r, resp)
		if err != nil {
			return nil, err
		}
	}

	// backend fetch
	var err error
	if resp == nil {
		resp, err = cproxy.BackendFetch(r, config)
		if err != nil {
			return nil, err
		}
	}

	// call 'OnCollectSubRequest' , mostly used for cache esi
	log.Println("EVENT :: OnCollectSubRequest")
	subResps := make([][]*http.Response, 0)
	for index := range exts {
		subReqs, err := exts[index].OnCollectSubRequests(resp)
		if err != nil {
			return nil, err
		}
		resps := make([]*http.Response, 0)
		for subIndex := range subReqs {
			resp, err := HandleRequest(subReqs[subIndex], config, exts)
			if err != nil {
				return nil, err
			}
			resps = append(
				resps,
				resp,
			)
		}
		subResps = append(
			subResps,
			resps,
		)
	}

	// call 'OnResponse'
	log.Println("EVENT :: OnResponse")
	for index := range exts {
		err := exts[index].OnResponse(
			resp,
			subResps[index],
		)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

func main() {

	// display app name + version
	log.Printf("%s v%.2f", cproxy.AppName, cproxy.VersionNo/100.0)

	// command line args
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	configFilePath := flag.String(
		"config-path",
		filepath.Join(filepath.Dir(execPath), cproxy.DefaultConfigFilePath),
		"Path to store cache files in for file system cache.",
	)
	enableExts := flag.String(
		"extensions",
		"",
		"Comma delimited list of extensions to enable.",
	)
	flag.Parse()

	// load config
	config := cproxy.LoadConfigFile(*configFilePath)
	if *enableExts != "" {
		config.Extensions.Enabled = strings.Split(*enableExts, ",")
	}

	// load extensions
	exts, err := cproxy.LoadExtensions(&config)
	if err != nil {
		panic(err)
	}
	defer cproxy.UnloadExtensions(exts)

	// create listener
	listener, err := cproxy.GetListener(&config)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// handle incoming request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// handle request
		resp, err := HandleRequest(
			r,
			&config,
			exts,
		)
		if err != nil {
			panic(err)
		}

		// set response headers
		for k, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(k, value)
			}
		}
		w.Header().Add("X-Proxy-Name", cproxy.AppName)

		// write status code
		w.WriteHeader(resp.StatusCode)
		// set response body
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			cproxy.RenderErrorPage(w, r, err)
		}

	})

	switch config.ProxyType {
	case cproxy.ProxyTypeFCGI:
		{
			// listen for cgi requests
			log.Println("INIT :: Listen for FastCGI requests on", config.Listen+".")
			err := fcgi.Serve(listener, handler)
			if err != nil {
				panic(err)
			}
			break
		}
	case cproxy.ProxyTypeHTTP:
		{
			// listen for http requests
			log.Println("INIT :: Listen for HTTP requests on", config.Listen+".")
			err := http.Serve(listener, handler)
			if err != nil {
				panic(err)
			}
			break
		}
	}

}
