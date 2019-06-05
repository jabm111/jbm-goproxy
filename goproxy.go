package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type Config struct {
	ProxyScheme  string
	ProxyHost    string
	StaticDir    string
	StaticPrefix string
	Port         string
	Domains      string
	CertCacheDir string
}

var config *Config

func setGlobalConfig() {
	config = &Config{
		ProxyScheme:  getEnv("GO_PROXY_SCHEME", "http"),
		ProxyHost:    getEnv("GO_PROXY_HOST", "localhost:8080"),
		StaticDir:    getEnv("GO_STATIC_DIR", "static"),
		StaticPrefix: getEnv("GO_STATIC_PREFIX", "static"),
		Port:         getEnv("GO_PORT", ":8888"),
		Domains:      getEnv("GO_DOMAINS", ""),
		CertCacheDir: getEnv("GO_CERT_CACHE_DIR", getEnv("HOME", "/home/gouser")+"/letsencrypt"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

var templates = template.Must(template.New("basic").Parse(`<html><body><p>{{.}}</p></body></html>`))

func defineRoutes(mux *http.ServeMux) {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", config.ProxyHost)
			req.URL.Scheme = config.ProxyScheme
			req.URL.Host = config.ProxyHost

			fmt.Printf("goproxy: %+v\n", req.RequestURI)
		},
		ModifyResponse: func(res *http.Response) error {
			if res.StatusCode >= 500 {
				return errors.New("Error: 500")
			}

			return nil
		},
		ErrorHandler: func(res http.ResponseWriter, req *http.Request, err error) {
			templates.ExecuteTemplate(res, "basic", "An error occurred")
		},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})
}

func makeHTTPServer() *http.Server {
	staticDir := config.StaticDir
	staticServer := http.FileServer(http.Dir(staticDir))
	staticPrefix := config.StaticPrefix

	mux := http.NewServeMux()
	mux.Handle("/"+staticPrefix+"/", http.StripPrefix("/"+staticPrefix+"/", staticServer))
	defineRoutes(mux)

	server := &http.Server{
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server
}

func printConsole(s string) {
}

func main() {
	setGlobalConfig()
	fmt.Printf("goproxy: config: %+v\n", config)

	if config.Port == ":443" {
		whitelist := strings.Split(config.Domains, ",")

		var hostWhitelist autocert.HostPolicy

		if len(whitelist) >= 1 {
			hostWhitelist = autocert.HostWhitelist(whitelist...)
		} else {
			panic("Whitelist must be at least 1 domain comma separated in the GO_DOMAINS environment variable.")
		}

		certCacheDir := config.CertCacheDir

		manager := &autocert.Manager{
			Cache:      autocert.DirCache(certCacheDir),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostWhitelist,
		}
		server := makeHTTPServer()
		server.Addr = config.Port
		server.TLSConfig = manager.TLSConfig()
		server.TLSConfig.MinVersion = tls.VersionTLS12
		server.TLSConfig.CurvePreferences = []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256}
		server.TLSConfig.PreferServerCipherSuites = true
		server.TLSConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		}

		go http.ListenAndServe(":80", manager.HTTPHandler(nil))

		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		server := makeHTTPServer()
		server.Addr = config.Port

		log.Fatal(server.ListenAndServe())
	}
}
