package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/crypto/acme/autocert"

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

	// // https
	// keyFile := os.Getenv("KEY_PATH")
	// crtFile := os.Getenv("CRT_PATH")
	// logger.Print("Start GCR Proxy")
	// logger.Fatal(http.ListenAndServeTLS(":8000", crtFile, keyFile, proxy))

	// // http
	// logger.Print("Start GCR Proxy")
	// logger.Fatal(http.ListenAndServe(":8000", proxy))

	// autocert
	certManager, err := getcertManager()
	if err != nil {
		panic(err)
	}

	// http-01 Challenge(ドメインの所有確認)、HTTPSへのリダイレクト用のサーバー
	challengeServer := &http.Server{
		Handler: certManager.HTTPHandler(nil),
		Addr:    ":8080",
	}
	go challengeServer.ListenAndServe()

	// proxy server
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: certManager.TLSConfig(),
		Handler:   proxy,
	}

	logger.Print("Start GCR Proxy")
	logger.Fatal(server.ListenAndServeTLS("", ""))
}

func getcertManager() (*autocert.Manager, error) {
	proxyUrl := os.Getenv("PROXY_URL")
	if proxyUrl == "" {
		return nil, fmt.Errorf("Invalid PROXY_URL %s", proxyUrl)
	}
	u, err := url.Parse(proxyUrl)
	if err != nil {
		return nil, err
	}

	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,             // Let's Encryptの利用規約への同意
		HostPolicy: autocert.HostWhitelist(u.Host), // ドメイン名
		Cache:      autocert.DirCache("certs"),     // 証明書などを保存するフォルダ
	}
	return certManager, nil
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
