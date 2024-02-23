package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ProxyServer struct {
	Port           int
	ProxyUrl       []ProxyUrlType
	Authenticated  bool
	NumberOfFailed int
	Ctx            context.Context
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		Port:           8080,
		Authenticated:  false,
		NumberOfFailed: 0,
		Ctx:            context.Background(),
	}
}

func (p *ProxyServer) SetPort(port int) {
	p.Port = port
}

func (p *ProxyServer) GetPort() int {
	return p.Port
}

func (p *ProxyServer) SetProxyUrl(proxyUrl []ProxyUrlType) {
	p.ProxyUrl = proxyUrl
}

func (p *ProxyServer) GetProxyUrl() []ProxyUrlType {
	return p.ProxyUrl
}

func (p *ProxyServer) SetAuthenticated(authenticated bool) {
	p.Authenticated = authenticated
}

func (p *ProxyServer) GetAuthenticated() bool {
	return p.Authenticated
}

func (p *ProxyServer) SetNumberOfFailed(numberOfFailed int) {
	p.NumberOfFailed = numberOfFailed
}

func (p *ProxyServer) GetNumberOfFailed() int {
	return p.NumberOfFailed
}

func (p *ProxyServer) Start() {
	p.Handler()
	log("Starting the proxy server on port: " + strconv.Itoa(p.Port))
	if err := http.ListenAndServe("0.0.0.0:"+strconv.Itoa(p.Port), nil); err != nil {
		errorLog(err.Error())
		os.Exit(1)
	}
}

func (p *ProxyServer) Handler() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log("Url: /, Request from " + r.RemoteAddr)
		fmt.Fprintf(w, "It works!")
	})

	for _, proxy := range p.ProxyUrl {
		debugLog("Proxy name: " + proxy.Name + ", Proxy url: " + proxy.Url)

		func(p2 ProxyUrlType) {
			http.HandleFunc(fmt.Sprintf("/%s/", p2.Name), func(w http.ResponseWriter, r *http.Request) {
				log("Url: " + r.RequestURI + ", Request from " + r.RemoteAddr)

				if p.Authenticated {
					remoteIp := strings.Split(r.RemoteAddr, ":")[0]
					failedCount, _ := redisConn.Get(p.Ctx, "proxy_with_authenticate:failed:"+remoteIp).Result()
					failedCountInt, err := strconv.Atoi(failedCount)
					if err != nil {
						failedCountInt = 0
					}

					if p.NumberOfFailed != 0 && failedCountInt > p.NumberOfFailed {
						errorLog(remoteIp + " too many failed attempts, please try again later")
						http.Error(w, "Too many failed attempts, please try again later", http.StatusUnauthorized)
						return
					}

					otp := r.Header.Get("x-google-code")
					if otp == "" {
						errorLog("No google code found in the header")
						http.Error(w, "No google code found in the header", http.StatusUnauthorized)
						return
					}

					if !verifyOTP(passkey, otp) {
						errorLog("Invalid google code")
						redisConn.Set(p.Ctx, "proxy_with_authenticate:failed:"+remoteIp, failedCountInt+1, REDIS_TIMEOUT)
						http.Error(w, "Invalid google code", http.StatusUnauthorized)
						return
					}

					if redisOtp, _ := redisConn.Get(p.Ctx, "proxy_with_authenticate:otp").Result(); redisOtp == otp {
						errorLog("The google code has been used")
						http.Error(w, "The google code has been used", http.StatusUnauthorized)
						return
					}

					redisConn.Set(p.Ctx, "proxy_with_authenticate:otp", otp, REDIS_TIMEOUT)
					redisConn.Del(p.Ctx, "proxy_with_authenticate:failed:"+remoteIp)

					log("Valid google code, start to proxy the url: " + p2.Url)
					p.proxyUrl(r, p2, w)
				} else {
					p.proxyUrl(r, p2, w)
				}
			})
		}(proxy)
	}
}

func (p *ProxyServer) proxyUrl(r *http.Request, p2 ProxyUrlType, w http.ResponseWriter) {
	proxyPath := strings.Split(r.RequestURI, p2.Name)[1]
	proxyUrl := p2.Url + proxyPath
	debugLog("Proxy url: " + proxyUrl)
	proxyReq, err := http.NewRequest(r.Method, proxyUrl, r.Body)
	if err != nil {
		errorLog(err.Error())
		http.Error(w, "Error creating proxy request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		debugLog("Header: " + name + " = " + strings.Join(values, ","))
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	resp, err := http.DefaultTransport.RoundTrip(proxyReq)
	if err != nil {
		errorLog(err.Error())
		http.Error(w, "Error sending proxy: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
