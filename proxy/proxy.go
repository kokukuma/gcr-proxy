// Proxy for docker pull from GCR without authentication
package proxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Proxy struct {
	authData  string // "user:password"
	jsonKey   []byte
	proxyHost string
	transport *http.Transport
	logger    *log.Logger
}

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.IsAbs())
	fmt.Println(r.URL.Scheme)
	baseHost := r.Host

	// Before exec request (checkAuthenticate and rewrite request)
	if r.URL.Path == "/v2/token" {
		if err := proxy.checkAuthenticate(r); err != nil {
			proxy.handleError(w, 500, err)
			return
		}
		proxy.rewriteAuthorizationHeader(r)
	}
	proxy.rewriteRequestHost(r)

	// do Request
	resp, err := proxy.transport.RoundTrip(r)
	if err != nil {
		proxy.handleError(w, 500, err)
		return
	}
	defer resp.Body.Close()

	// Rewrite Response
	if r.URL.Path == "/v2/" {
		proxy.rewriteAuthenticateHeader(resp, baseHost)
	}

	//
	if err := writeResponse(w, resp); err != nil {
		proxy.handleError(w, 500, err)
		return
	}
	proxy.logger.Print(fmt.Sprintf("%s %s %s %s", resp.Request.RemoteAddr, resp.Status, resp.Request.Method, resp.Request.URL))
}

func (proxy *Proxy) handleError(w http.ResponseWriter, status int, err error) {
	proxy.logger.Print("[Error] ", err)
	http.Error(w, err.Error(), status)
}

func (proxy *Proxy) checkAuthenticate(r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	data := strings.Split(authHeader, " ")
	if len(data) == 2 {
		data, err := base64.StdEncoding.DecodeString(data[1])
		if err != nil {
			return err
		}

		if string(data) == proxy.authData {
			return nil
		}
	}
	return errors.New("Failed to authentication check")
}

func (proxy *Proxy) rewriteRequestHost(r *http.Request) {
	r.URL.Scheme = "https"
	r.URL.Host = proxy.proxyHost
	r.Host = proxy.proxyHost
	r.RequestURI = r.URL.String()
}

func (proxy *Proxy) rewriteAuthorizationHeader(r *http.Request) {
	r.Header.Del("Authorization")
	r.Header.Add("Authorization", "Basic "+basicAuth("_json_key", string(proxy.jsonKey)))
}

func (proxy *Proxy) rewriteAuthenticateHeader(resp *http.Response, host string) {
	resp.Header.Del("Www-Authenticate")
	// TODO: proxyをhttpで使ったとき,
	resp.Header.Add("Www-Authenticate", fmt.Sprintf("Bearer realm=\"https://%s/v2/token\",service=\"gcr.io\"", host))
}

func (proxy *Proxy) SetTransport(transport *http.Transport) {
	proxy.transport = transport
}

func (proxy *Proxy) SetLogger(logger *log.Logger) {
	proxy.logger = logger
}

func NewProxy(authData string, jsonKey []byte, proxyHost string) *Proxy {

	proxy := Proxy{
		authData:  authData,
		jsonKey:   jsonKey,
		proxyHost: proxyHost,
		transport: &http.Transport{
			ResponseHeaderTimeout: 5 * time.Second,
		},
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	return &proxy
}

func writeResponse(w http.ResponseWriter, resp *http.Response) error {
	// headerをコピー
	header := w.Header()
	for name, values := range resp.Header {
		for _, value := range values {
			header.Add(name, value)
		}
	}

	// status codeをコピー
	w.WriteHeader(resp.StatusCode)

	// bodyをコピー
	if _, err := io.Copy(w, resp.Body); err != nil {
		return err
	}
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
