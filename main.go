package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var configFileName string
var debug bool
var showQrCode bool
var generateSecret bool
var startServer bool

func init() {
	flag.BoolVar(&debug, "v", false, "verbose output")
	flag.BoolVar(&showQrCode, "q", false, "show QR code")
	flag.BoolVar(&generateSecret, "g", false, "generate new secret key")
	flag.BoolVar(&startServer, "s", false, "start the server")
	flag.StringVar(&configFileName, "c", "config.json", "config file name")
	flag.Parse()

	readConfigFile(configFileName)

	if !startServer {
		return
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + strconv.Itoa(config.Redis.Port),
		Password: config.Redis.Auth,
		DB:       config.Redis.Db,
	})
}

func main() {
	debugLog("Starting the proxy server...")

	if generateSecret {
		generatePassKey()
	}

	if config.PassKey == nil {
		errorLog("No pass_key found in the config file, eg: `\"pass_key\": \"\",`")
		os.Exit(1)
	}

	if config.PassKey != nil {
		passkey = *config.PassKey
	}

	if passkey == "" {
		log("No pass_key found in the config file")
		log("Generating a new secret key...")
		generatePassKey()
		generateTOTPWithSecret(passkey)
	}

	if showQrCode {
		generateTOTPWithSecret(passkey)
	}

	if !startServer {
		if !showQrCode && !generateSecret {
			fmt.Println(`Usage of ./build/proxy_server:
  -c string
  		config file name (default "config.json")
  -g    generate new secret key
  -q    show QR code
  -s    start the server
  -v    verbose output`)
		}
		os.Exit(0)
	}

	if config.ProxyUrl == nil {
		errorLog("No proxy_url found in the config file")
		os.Exit(1)
	}

	redisConn = redisClient.Conn()
	defer redisConn.Close()

	proxyServer := NewProxyServer()

	if config.ServerPort != nil {
		proxyServer.SetPort(*config.ServerPort)
	}

	if config.Authenticated != nil {
		proxyServer.SetAuthenticated(*config.Authenticated)
	}

	if config.NumberOfFailed != nil {
		proxyServer.SetNumberOfFailed(*config.NumberOfFailed)
	}

	proxyServer.SetProxyUrl(*config.ProxyUrl)
	proxyServer.Start()
}
