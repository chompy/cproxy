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
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// AppName - name of this application
const AppName = "CProxy"

// VersionNo - current version of this application
const VersionNo = 1

// DefaultConfigFilePath - default path to config file
const DefaultConfigFilePath = "cproxy.json"

// ProxyTypeHTTP - denotes HTTP proxy
const ProxyTypeHTTP = "http"

// ProxyTypeFCGI - denotes FastCGI proxy
const ProxyTypeFCGI = "fcgi"

// ProxyTypeDummy - denotes dummy test proxy
const ProxyTypeDummy = "dummy"

// Config - app configuration struct
type Config struct {
	ProxyType  string `json:"proxy_type"`
	Listen     string `json:"listen"`  // 8081, /app/listen.sock
	Backend    string `json:"backend"` // 127.0.0.1:9000, /app/run.sock, https://www.example.com
	Extensions struct {
		Path    string                     `json:"path"`
		Enabled []string                   `json:"enabled"`
		Config  map[string]json.RawMessage `json:"config"`
	} `json:"extensions"`
}

// GetDefaultConfig - get default configuration values
func GetDefaultConfig() Config {
	// get listen port from env var (platform.sh)
	listenPort := os.Getenv("PORT")
	proxyType := ProxyTypeFCGI // assume fcgi if port env var
	if listenPort == "" {
		listenPort = "8081"
		proxyType = ProxyTypeHTTP
	}
	listenPort = ":" + listenPort
	config := Config{
		ProxyType: proxyType,
		Listen:    listenPort,
		Backend:   "/run/app.sock",
	}
	config.Extensions.Path = "ext"
	execPath, err := os.Executable()
	if err == nil {
		config.Extensions.Path = filepath.Join(filepath.Dir(execPath), "ext")
	}
	return config
}

// LoadConfigFile - load configuration from file
func LoadConfigFile(configFilePath string) Config {
	if configFilePath == "" {
		configFilePath = DefaultConfigFilePath
		execPath, err := os.Executable()
		if err == nil {
			configFilePath = filepath.Join(filepath.Dir(execPath), DefaultConfigFilePath)
		}
	}
	config := GetDefaultConfig()
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Println("CONFIG :: Warning,", err)
		return config
	}
	defer f.Close()
	configBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("CONFIG :: Warning,", err)
		return config
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Println("CONFIG :: Warning,", err)
	}
	return config
}
