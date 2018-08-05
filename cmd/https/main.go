package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/kokukuma/gcr-proxy/proxy"
)

func main() {

	// new proxy
	logger := log.New(os.Stdout, "[GCR Proxy] ", log.LstdFlags)

	proxy, err := getProxy()
	if err != nil {
		panic(err)
	}
	proxy.SetLogger(logger)

	// https
	keyFile := os.Getenv("KEY_PATH")
	crtFile := os.Getenv("CRT_PATH")
	logger.Print("Start GCR Proxy")
	logger.Fatal(http.ListenAndServeTLS(":8000", crtFile, keyFile, proxy))
}

func getProxy() (*proxy.Proxy, error) {
	// load json key
	jsonKeyPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	jsonKey, err := ioutil.ReadFile(jsonKeyPath)
	if err != nil {
		return nil, err
	}

	// proxyUrl
	proxyUrl := os.Getenv("PROXY_URL")
	if proxyUrl == "" {
		return nil, fmt.Errorf("Invalid PROXY_URL %s", proxyUrl)
	}

	proxyAuth := os.Getenv("PROXY_AUTH")
	if proxyAuth == "" {
		return nil, fmt.Errorf("Invalid PROXY_AUTH %s", proxyAuth)
	}

	registryUrl := os.Getenv("REGISTRY_URL")
	if registryUrl == "" {
		return nil, fmt.Errorf("Invalid REGISTRY_URL %s", registryUrl)
	}

	proxy, err := proxy.NewProxy(proxyAuth, jsonKey, proxyUrl, registryUrl)
	if err != nil {
		return nil, err
	}
	return proxy, nil
}
