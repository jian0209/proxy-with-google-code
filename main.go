package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/xlzd/gotp"
)

type Config struct {
	Authenticated *bool           `json:"authenticated"`
	ServerPort    *int            `json:"server_port"`
	Username      *string         `json:"username"`
	PassKey       *string         `json:"pass_key"`
	ProxyUrl      *[]ProxyUrlType `json:"proxy_url"`
}

type ProxyUrlType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

var configFileName string
var debug bool
var showQrCode bool
var generateSecret bool

var config Config
var passkey string

func init() {
	flag.BoolVar(&debug, "v", false, "verbose output")
	flag.BoolVar(&showQrCode, "q", false, "show QR code")
	flag.BoolVar(&generateSecret, "g", false, "generate secret key")
	flag.StringVar(&configFileName, "c", "config.json", "config file name")
	flag.Parse()

	ReadConfigFile(configFileName)
}

func main() {
	DebugLog("Starting the proxy server...")

	if config.PassKey == nil {
		ErrorLog("No pass_key found in the config file, eg: `\"pass_key\": \"\",`")
		os.Exit(1)
	}

	if generateSecret {
		passkey = gotp.RandomSecret(16)
		savePassKeyToFile(passkey)
		ReadConfigFile(configFileName)
	}

	if *config.PassKey == "" {
		ErrorLog("No pass_key found in the config file")
		os.Exit(1)
	}

	passkey = *config.PassKey
	if showQrCode {
		generateTOTPWithSecret(passkey)
	}

	getProxyUrl()

	var serverPort int
	if config.ServerPort == nil {
		serverPort = 8080
	} else {
		serverPort = *config.ServerPort
	}
	Log("Starting the proxy server on port " + strconv.Itoa(serverPort))
	if err := http.ListenAndServe(":"+strconv.Itoa(serverPort), nil); err != nil {
		ErrorLog(err.Error())
		os.Exit(1)
	}
}

func DebugLog(msg interface{}) {
	if debug {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [DEBUG] ", msg)
	}
}

func Log(msg interface{}) {
	// get current time
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [INFO] ", msg)
}

func ErrorLog(msg interface{}) {
	// get current time
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [ERROR] ", msg)
}

func ReadConfigFile(filename string) {
	// read the config file
	file, err := os.Open(filename)
	if err != nil {
		Log(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// decode the json file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		ErrorLog(err.Error())
		os.Exit(1)
	}
}

func savePassKeyToFile(secretkey string) {
	DebugLog("open " + configFileName)
	file, err := os.OpenFile(configFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		ErrorLog(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "pass_key") {
			var changedLine string
			if *config.PassKey == "" {
				changedLine = strings.Replace(line, "\"pass_key\": \"\"", "\"pass_key\": \""+secretkey+"\"", 1)
			} else {
				changedLine = strings.Replace(line, *config.PassKey, secretkey, 1)
			}
			DebugLog("Changing the pass_key from " + *config.PassKey + " to " + secretkey)
			DebugLog("Old line: " + line)
			DebugLog("New line: " + changedLine)
			if runtime.GOOS == "darwin" {
				exec.Command("sed", "-i", "''", "s/"+line+"/"+changedLine+"/g", configFileName).Run()
			} else if runtime.GOOS == "linux" {
				exec.Command("sed", "-i", "s/"+line+"/"+changedLine+"/g", configFileName).Run()
			}
			break
		}
	}
	if _, err := os.Stat(configFileName + "''"); err == nil {
		os.Remove(configFileName + "''")
	}
}

func generateTOTPWithSecret(secretkey string) {
	totp := gotp.NewDefaultTOTP(secretkey)
	Log("Current one-time password is:" + totp.Now())

	uri := totp.ProvisioningUri(*config.Username, "proxy_with_google_code")
	Log("Secret Key URI:" + uri)

	// Generate and display QR code in the terminal
	qrterminal.GenerateWithConfig(uri, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	})
}

func verifyOTP(secretkey string, currentkey string) bool {
	totp := gotp.NewDefaultTOTP(secretkey)

	return totp.Verify(currentkey, time.Now().Unix())
}

func getProxyUrl() {
	if config.ProxyUrl == nil {
		ErrorLog("No proxy_url found in the config file")
		os.Exit(1)
	}

	proxyUrl := *config.ProxyUrl

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Log("Url: /, Request from " + r.RemoteAddr)
		fmt.Fprintf(w, "It works!")
	})

	for _, proxy := range proxyUrl {
		DebugLog("Proxy name: " + proxy.Name + ", Proxy url: " + proxy.Url)
		func(p ProxyUrlType) {
			http.HandleFunc(fmt.Sprintf("/%s", p.Name), func(w http.ResponseWriter, r *http.Request) {
				Log("Url: " + r.RequestURI + ", Request from " + r.RemoteAddr)
				proxyUrl := func() {
					proxyReq, err := http.NewRequest(r.Method, p.Url, r.Body)
					if err != nil {
						ErrorLog(err.Error())
						http.Error(w, "Error creating proxy request: "+err.Error(), http.StatusInternalServerError)
						return
					}

					for name, values := range r.Header {
						for _, value := range values {
							proxyReq.Header.Add(name, value)
						}
					}

					resp, err := http.DefaultTransport.RoundTrip(proxyReq)
					if err != nil {
						ErrorLog(err.Error())
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

				if config.Authenticated != nil && *config.Authenticated {
					otp := r.Header.Get("x-google-code")
					if otp == "" {
						ErrorLog("No google code found in the header")
						http.Error(w, "No google code found in the header", http.StatusUnauthorized)
						return
					}

					if !verifyOTP(passkey, otp) {
						ErrorLog("Invalid google code")
						http.Error(w, "Invalid google code", http.StatusUnauthorized)
						return
					}

					Log("Valid google code, start to proxy the url: " + p.Url)
					proxyUrl()
				} else {
					proxyUrl()
				}
			})
		}(proxy)
	}
}
