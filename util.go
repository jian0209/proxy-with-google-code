package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func debugLog(msg interface{}) {
	if debug {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [DEBUG] ", msg)
	}
}

func log(msg interface{}) {
	// get current time
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [INFO] ", msg)
}

func errorLog(msg interface{}) {
	// get current time
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " [ERROR] ", msg)
}

func readConfigFile(filename string) {
	// read the config file
	file, err := os.Open(filename)
	if err != nil {
		log(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// decode the json file
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		errorLog(err.Error())
		os.Exit(1)
	}
}

func savePassKeyToFile(secretkey string) {
	debugLog("open " + configFileName)
	file, err := os.OpenFile(configFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		errorLog(err.Error())
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
			debugLog("Changing the pass_key from " + *config.PassKey + " to " + secretkey)
			debugLog("Old line: " + line)
			debugLog("New line: " + changedLine)
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
