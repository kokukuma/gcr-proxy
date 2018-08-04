package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kokukuma/gcr-proxy/proxy"
)

func main() {
	// load json key
	// jsonKeyPath := os.Getenv("SERVICE_ACCOUNT_PATH")
	// jsonKey, err := ioutil.ReadFile(jsonKeyPath)
	// if err != nil {
	// 	panic(err)
	// }

	// proxyHost
	proxyHost := os.Getenv("PROXY_HOST")
	if proxyHost == "" {
		panic(fmt.Sprintf("Invalid PROXY_HOST %s", proxyHost))
	}

	proxyAuth := os.Getenv("PROXY_AUTH")
	if proxyAuth == "" {
		panic(fmt.Sprintf("Invalid PROXY_AUTH %s", proxyAuth))
	}

	// new proxy
	logger := log.New(os.Stdout, "[GCR Proxy] ", log.LstdFlags)

	//proxy := proxy.NewProxy(proxyAuth, jsonKey, proxyHost)
	proxy := proxy.NewProxy(proxyAuth, []byte{}, proxyHost)
	proxy.SetLogger(logger)

	// keyFile := os.Getenv("KEY_PATH")
	// crtFile := os.Getenv("CRT_PATH")

	logger.Print("Start GCR Proxy")
	// logger.Fatal(http.ListenAndServeTLS(":8000", crtFile, keyFile, proxy))
	logger.Fatal(http.ListenAndServe(":8000", proxy))
}
