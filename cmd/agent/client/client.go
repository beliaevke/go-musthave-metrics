// Package client
package client

import (
	"fmt"
	"time"

	"musthave-metrics/cmd/agent/config"
)

type Locallink struct {
	RunAddr         string
	Method          string
	ContentType     string
	ContentEncoding string
	PollInterval    int
	ReportInterval  int
	HashKey         string
	RateLimit       int
	PublicKeyPath   string
	SecretToken     string
}

func (locallink *Locallink) Run() error {
	var err error
	cfg := config.ParseFlags()
	locallink.RunAddr = cfg.FlagRunAddr
	locallink.Method = "/update/"
	locallink.ContentType = "text/plain"
	locallink.ContentEncoding = "gzip"
	locallink.ReportInterval = cfg.FlagReportInterval
	locallink.PollInterval = cfg.FlagPollInterval
	locallink.HashKey = cfg.FlagHashKey
	locallink.RateLimit = cfg.FlagRateLimit
	locallink.PublicKeyPath = cfg.FlagCryptoKey
	locallink.SecretToken = "SecretToken"
	fmt.Printf("%s (!) Running server on %s, Report interval: %v, Poll interval: %v\n", time.Now().Format(time.DateTime), cfg.FlagRunAddr, cfg.FlagReportInterval, cfg.FlagPollInterval)
	return err
}
