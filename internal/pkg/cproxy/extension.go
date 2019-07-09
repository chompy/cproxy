package cproxy

import (
	"log"
	"net/http"
	"path"
	"plugin"
)

// Extension - app extension
type Extension struct {
	OnUnload             func() error
	OnRequest            func(req *http.Request) (*http.Response, error)
	OnCollectSubRequests func(resp *http.Response) ([]*http.Request, error)
	OnResponse           func(resp *http.Response, subResps []*http.Response) (*http.Response, error)
}

// LoadExtensions - load extensions
func LoadExtensions(config *Config) ([]Extension, error) {
	exts := make([]Extension, 0)
	for _, name := range config.Extensions.Enabled {
		ext := Extension{}
		plugin, err := plugin.Open(
			path.Join(config.Extensions.Path, name),
		)
		if err != nil {
			return nil, err
		}
		log.Println("EXTENSION ::", name, "loaded.")
		// on load
		extOnLoad, err := plugin.Lookup("OnLoad")
		if err != nil {
			return nil, err
		}
		err = extOnLoad.(func() error)()
		if err != nil {
			return nil, err
		}
		// on unload
		extOnUnload, err := plugin.Lookup("OnUnload")
		if err != nil {
			return nil, err
		}
		ext.OnUnload = extOnUnload.(func() error)
		// on request
		extOnRequest, err := plugin.Lookup("OnRequest")
		if err != nil {
			return nil, err
		}
		ext.OnRequest = extOnRequest.(func(req *http.Request) (*http.Response, error))
		// on collect sub requests
		extOnCollectSubRequests, err := plugin.Lookup("OnCollectSubRequests")
		if err != nil {
			return nil, err
		}
		ext.OnCollectSubRequests = extOnCollectSubRequests.(func(resp *http.Response) ([]*http.Request, error))
		// on response
		extOnResponse, err := plugin.Lookup("OnResponse")
		if err != nil {
			return nil, err
		}
		ext.OnResponse = extOnResponse.(func(resp *http.Response, subResps []*http.Response) (*http.Response, error))

		exts = append(exts, ext)
	}
	return exts, nil
}

// UnloadExtensions - unload extensions
func UnloadExtensions(exts []Extension) error {
	for index := range exts {
		exts[index].OnUnload()
		// not sure if we want to handle errors here or not
		/*if err != nil {
			return err
		}*/
	}
	return nil
}
