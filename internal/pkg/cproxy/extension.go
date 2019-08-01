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
	"log"
	"net/http"
	"path"
	"plugin"
)

// Extension - cproxy extension data
type Extension struct {
	Name       string
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
		// get config
		rawConfig := []byte{}
		if val, ok := config.Extensions.Config[name]; ok {
			rawConfig = val
		}
		// on load
		extOnLoad, err := plugin.Lookup("OnLoad")
		if err != nil {
			return nil, err
		}
		extOnLoad.(func(subRequestCallback func(req *http.Request) (*http.Response, error), rawConfig []byte) error)(subRequestCallback, rawConfig)
		// get name
		extName := name
		extGetName, err := plugin.Lookup("GetName")
		if err == nil {
			extName = extGetName.(func() string)()
		}
		log.Println("EXTENSION ::", extName, "loaded.")
		// create ext reference
		ext := Extension{
			Name: extName,
		}
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
