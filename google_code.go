package main

import (
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/xlzd/gotp"
)

func generateTOTPWithSecret(secretkey string) {
	totp := gotp.NewDefaultTOTP(secretkey)
	log("Current one-time password is:" + totp.Now())

	uri := totp.ProvisioningUri(*config.Username, "proxy_with_google_code")
	log("Secret Key URI:" + uri)

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
