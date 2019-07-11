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
		"Path to configuration file.",
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
	var exts []cproxy.Extension
	exts, err = cproxy.LoadExtensions(
		&config,
		func(req *http.Request) (*http.Response, error) {
			return cproxy.HandleRequest(req, &config, &exts)
		},
	)
	if err != nil {
		panic(err)
	}

	// create listener
	listener, err := cproxy.GetListener(&config)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// handle incoming request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// handle request
		resp, err := cproxy.HandleRequest(
			r,
			&config,
			&exts,
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

	// determine proxy type and begin listening on configured port
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
