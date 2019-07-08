package cproxy

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
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

// Config - app configuration struct
type Config struct {
	ProxyType string `json:"proxy_type"`
	Listen    string `json:"listen"`  // 8081, /app/listen.sock
	Connect   string `json:"connect"` // 127.0.0.1:9000, /app/run.sock
	Plugins   struct {
		Enabled []string               `json:"enabled"`
		Config  map[string]interface{} `json:"config"`
	} `json:"plugins"`
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
	return Config{
		ProxyType: proxyType,
		Listen:    listenPort,
		Connect:   "/run/app.sock",
	}
}

// LoadConfigFile - load configuration from file
func LoadConfigFile(configFilePath string) Config {
	if configFilePath == "" {
		configFilePath = DefaultConfigFilePath
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
