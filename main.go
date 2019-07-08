package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/fcgi"

	"./internal/pkg/cproxy"
)

func main() {

	// display app name + version
	log.Printf("%s v%.2f", cproxy.AppName, cproxy.VersionNo/100.0)

	// command line args
	configFilePath := flag.String(
		"config-path",
		cproxy.DefaultConfigFilePath,
		"Path to store cache files in for file system cache.",
	)
	flag.Parse()

	// load config
	config := cproxy.LoadConfigFile(*configFilePath)

	// create listener
	listener, err := cproxy.GetListener(&config)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	// handle incoming request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("REQUEST ::", r.Method, r.URL.String())

		// TODO fire OnRequest event

		// TODO if OnRequest event provides response then don't perform backend fetch

		// backend fetch
		resp, err := cproxy.BackendFetch(r, &config)
		if err != nil {
			panic(err)
		}

		// TODO fire OnResponse event

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
