package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/xlzd/gotp"
)

var configFileName string
var debug bool
var showQrCode bool
var generateSecret bool

func init() {
	flag.BoolVar(&debug, "v", false, "verbose output")
	flag.BoolVar(&showQrCode, "q", false, "show QR code")
	flag.BoolVar(&generateSecret, "g", false, "generate secret key")
	flag.StringVar(&configFileName, "c", "config.json", "config file name")
	flag.Parse()

	readConfigFile(configFileName)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + strconv.Itoa(config.Redis.Port),
		Password: config.Redis.Auth,
		DB:       config.Redis.Db,
	})
}

func main() {
	debugLog("Starting the proxy server...")

	if config.PassKey == nil {
		errorLog("No pass_key found in the config file, eg: `\"pass_key\": \"\",`")
		os.Exit(1)
	}

	if config.ProxyUrl == nil {
		errorLog("No proxy_url found in the config file")
		os.Exit(1)
	}

	if generateSecret {
		passkey = gotp.RandomSecret(16)
		savePassKeyToFile(passkey)
		readConfigFile(configFileName)
	}

	if *config.PassKey == "" {
		errorLog("No pass_key found in the config file")
		os.Exit(1)
	}

	passkey = *config.PassKey
	if showQrCode {
		generateTOTPWithSecret(passkey)
	}

	redisConn = redisClient.Conn()
	defer redisConn.Close()

	proxyServer := NewProxyServer()

	if config.ServerPort == nil {
		proxyServer.SetPort(8080)
	} else {
		proxyServer.SetPort(*config.ServerPort)
	}

	if config.Authenticated == nil {
		proxyServer.SetAuthenticated(false)
	} else {
		proxyServer.SetAuthenticated(*config.Authenticated)
	}

	if config.NumberOfFailed == nil {
		proxyServer.SetNumberOfFailed(0)
	} else {
		proxyServer.SetNumberOfFailed(*config.NumberOfFailed)
	}

	proxyServer.SetProxyUrl(*config.ProxyUrl)
	proxyServer.Start()
}
