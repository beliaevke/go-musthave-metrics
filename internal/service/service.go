package service

import (
	"fmt"
)

func MakeURL(runAddr string, method string, mtype string, mname string, mvalue string) string {
	return "http://" + runAddr + method + mtype + "/" + mname + "/" + fmt.Sprintf("%v", mvalue)
}
