package proxy_test

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kokukuma/gcr-proxy/proxy"
)

// target server
type FakeGCRHandler struct{}

func (h FakeGCRHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		header := w.Header()
		header.Add("test_key", "test_value")
		io.WriteString(w, "/")

	} else if r.URL.Path == "/v2/" {
		io.WriteString(w, "/v2/")

	} else if r.URL.Path == "/v2/token" {
		// テスト用に受け取った Authorizationをそのまま返す
		header := w.Header()
		header.Add("Authorization", r.Header.Get("Authorization"))
		io.WriteString(w, "/2/token")

	} else {
		http.Error(w, "Not Found", 404)
	}

}

var https = httptest.NewTLSServer(FakeGCRHandler{})

func getProxyServer(url string) *httptest.Server {
	host := strings.Replace(https.URL, "https://", "", -1)
	proxy := proxy.NewProxy("test-user:test-password", []byte("test-key"), host)

	// // テスト用にログ削除
	// proxy.SetLogger(log.New(ioutil.Discard, "", log.LstdFlags))

	// テスト用にInsecureSkipVerify
	proxy.SetTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	server := httptest.NewTLSServer(proxy)
	return server
}

func getClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return client
}

func TestGCRProxyRoot(t *testing.T) {
	proxyServer := getProxyServer(https.URL)
	client := getClient()

	resp, err := client.Get(proxyServer.URL)
	if err != nil {
		t.Error("target server does not serve contents", err)
	}
	defer resp.Body.Close()

	// body
	txt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("response body could not read", err)
	}
	if string(txt) != "/" {
		t.Error("proxy server return unexpected value")
	}

	// header
	value := resp.Header.Get("test_key")
	if value != "test_value" {
		t.Error("proxy server does not return header")
	}
}

func TestGCRProxyV2ResponseHeaderReplace(t *testing.T) {
	proxyServer := getProxyServer(https.URL)
	client := getClient()
	resp, err := client.Get(proxyServer.URL + "/v2/")
	if err != nil {
		t.Error("target server does not serve contents", err)
	}
	defer resp.Body.Close()

	// header
	expected := fmt.Sprintf("Bearer realm=\"%s/v2/token\",service=\"gcr.io\"", proxyServer.URL)
	actual := resp.Header.Get("Www-Authenticate")
	if actual != expected {
		t.Error(fmt.Sprintf("proxy server returned unexpected header.\nactual:%v\nexpected:%v", actual, expected))
	}
}

func TestGCRProxyV2TokenRequestHeaderReplace(t *testing.T) {
	proxyServer := getProxyServer(https.URL)
	client := getClient()

	// Add Authorization header
	req, _ := http.NewRequest("GET", proxyServer.URL+"/v2/token", nil)
	authorizationHeader := fmt.Sprintf("Bearer %s", base64.StdEncoding.EncodeToString([]byte("test-user:test-password")))
	req.Header.Set("Authorization", authorizationHeader)

	resp, err := client.Do(req)
	if err != nil {
		t.Error("target server does not serve contents", err)
	}
	defer resp.Body.Close()

	// Check header replaced
	expected := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte("_json_key:test-key")))
	actual := resp.Header.Get("Authorization")
	if actual != expected {
		t.Error(fmt.Sprintf("proxy server returned unexpected header.\nactual:%v\nexpected:%v", actual, expected))
	}
}

func TestGCRProxyInvalidPath(t *testing.T) {
	proxyServer := getProxyServer(https.URL)
	client := getClient()
	resp, err := client.Get(proxyServer.URL + "/not_found_path")
	if err != nil {
		t.Error("target server does not serve contents", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Error("target server should be return 404", resp.StatusCode)
	}
}
