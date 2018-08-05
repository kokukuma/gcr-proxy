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
	// load json key
	jsonKeyPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	jsonKey, err := ioutil.ReadFile(jsonKeyPath)
	if err != nil {
		panic(err)
	}

	// proxyUrl
	proxyUrl := os.Getenv("PROXY_URL")
	if proxyUrl == "" {
		panic(fmt.Sprintf("Invalid PROXY_URL %s", proxyUrl))
	}

	proxyAuth := os.Getenv("PROXY_AUTH")
	if proxyAuth == "" {
		panic(fmt.Sprintf("Invalid PROXY_AUTH %s", proxyAuth))
	}

	registryUrl := os.Getenv("REGISTRY_URL")
	if registryUrl == "" {
		panic(fmt.Sprintf("Invalid REGISTRY_URL %s", registryUrl))
	}

	// new proxy
	logger := log.New(os.Stdout, "[GCR Proxy] ", log.LstdFlags)

	proxy, err := proxy.NewProxy(proxyAuth, jsonKey, proxyUrl, registryUrl)
	if err != nil {
		panic(err)
	}
	proxy.SetLogger(logger)

	// keyFile := os.Getenv("KEY_PATH")
	// crtFile := os.Getenv("CRT_PATH")

	logger.Print("Start GCR Proxy")
	//logger.Fatal(http.ListenAndServeTLS(":8000", crtFile, keyFile, proxy))
	logger.Fatal(http.ListenAndServe(":8000", proxy))
}
