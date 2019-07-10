package cproxy

import (
	"log"
	"net/http"
	"path"
	"plugin"
)

// Extension - cproxy extension data
type Extension struct {
	OnUnload   func()
	OnRequest  func(req *http.Request) (*http.Response, error)
	OnResponse func(resp *http.Response) (*http.Response, error)
}

// LoadExtensions - load extensions and initalize
func LoadExtensions(config *Config, subRequestCallback func(req *http.Request) (*http.Response, error)) ([]Extension, error) {
	exts := make([]Extension, 0)
	for _, name := range config.Extensions.Enabled {
		plugin, err := plugin.Open(
			path.Join(config.Extensions.Path, name),
		)
		if err != nil {
			return nil, err
		}
		// on load
		extOnLoad, err := plugin.Lookup("OnLoad")
		if err != nil {
			return nil, err
		}
		extOnLoad.(func(subRequestCallback func(req *http.Request) (*http.Response, error)) error)(subRequestCallback)
		log.Println("EXTENSION ::", name, "loaded.")
		// create ext reference
		ext := Extension{}
		// on unload
		extOnUnload, err := plugin.Lookup("OnUnload")
		if err != nil {
			return nil, err
		}
		ext.OnUnload = extOnUnload.(func())
		// on request
		extOnRequest, err := plugin.Lookup("OnRequest")
		if err != nil {
			return nil, err
		}
		ext.OnRequest = extOnRequest.(func(req *http.Request) (*http.Response, error))
		// on response
		extOnResponse, err := plugin.Lookup("OnResponse")
		if err != nil {
			return nil, err
		}
		ext.OnResponse = extOnResponse.(func(resp *http.Response) (*http.Response, error))
		// add ext to list
		exts = append(exts, ext)
	}
	return exts, nil
}

// UnloadExtensions - unload all extensions
func UnloadExtensions(exts *[]Extension) {
	for _, ext := range *exts {
		ext.OnUnload()
	}
}
