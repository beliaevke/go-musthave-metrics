package client

import (
	"bytes"
	"fmt"
	"net/http"
)

type ClientMetrics interface {
	UpdateMetrics(mtype string, mname string, mvalue string)
}

type Localhost struct {
	host        string
	port        string
	method      string
	contentType string
}

func defaultLocalhost() Localhost {
	return Localhost{
		"http://localhost",
		":8080",
		"/update/",
		"text/plain",
	}
}

func makeURL(localhost Localhost, mtype string, mname string, mvalue string) string {
	return localhost.host + localhost.port + localhost.method + mtype + "/" + mname + "/" + fmt.Sprintf("%v", mvalue)
}

func (localhost Localhost) UpdateMetrics(mtype string, mname string, mvalue string) error {
	client := &http.Client{}
	localhost = defaultLocalhost()
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
	defer response.Body.Close()
	return nil
}
