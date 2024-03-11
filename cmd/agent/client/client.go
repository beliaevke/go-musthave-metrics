package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ClientMetrics interface {
	UpdateMetrics(mtype string, mname string, mvalue string)
}

type Localhost struct {
	RunAddr        string
	method         string
	contentType    string
	PollInterval   int
	ReportInterval int
}

func makeURL(localhost Localhost, mtype string, mname string, mvalue string) string {
	return "http://" + localhost.RunAddr + localhost.method + mtype + "/" + mname + "/" + fmt.Sprintf("%v", mvalue)
}

func (localhost *Localhost) Run() error {
	var err error
	parseFlags()
	localhost.RunAddr = flagRunAddr
	localhost.method = "/update/"
	localhost.contentType = "text/plain"
	localhost.ReportInterval = flagReportInterval
	localhost.PollInterval = flagPollInterval
	fmt.Printf("%s (!) Running server on %s, Report interval: %v, Poll interval: %v\n", time.Now().Format(time.DateTime), flagRunAddr, flagReportInterval, flagPollInterval)
	return err
}

func (localhost Localhost) UpdateMetrics(mtype string, mname string, mvalue string) error {
	client := &http.Client{}
	url := makeURL(localhost, mtype, mname, mvalue)
	var body []byte
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", localhost.contentType)
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, response.Body)
	response.Body.Close()
	return nil
}
