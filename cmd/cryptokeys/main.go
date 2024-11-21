package main

import (
	"musthave-metrics/internal/crypt"
	"musthave-metrics/internal/logger"
)

func main() {
	err := crypt.MakeRSACert(
		&crypt.Settings{
			PathToCertificate: "D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key.pub",
			PathToPrivateKey:  "D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key",
		},
	)
	if err != nil {
		logger.Warnf("Error RSA Cert: " + err.Error())
		return
	}
}
