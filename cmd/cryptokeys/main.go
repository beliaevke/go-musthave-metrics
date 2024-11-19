package main

import "musthave-metrics/internal/crypt"

func main() {
	crypt.MakeRSACert(
		&crypt.Settings{
			PathToCertificate: "D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key.pub",
			PathToPrivateKey:  "D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key",
		},
	)
}
