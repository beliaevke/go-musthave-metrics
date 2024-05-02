package service

import (
	"fmt"
	"io"
	"net/http"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

type HashData struct {
	Key string
}

func NewHashData(hashKey string) *HashData {
	return &HashData{
		Key: hashKey,
	}
}

func MakeURL(runAddr string, method string, mtype string, mname string, mvalue string) string {
	return "http://" + runAddr + method + mtype + "/" + mname + "/" + fmt.Sprintf("%v", mvalue)
}

func MakeBatchUpdatesURL(runAddr string) string {
	return "http://" + runAddr + "/updates/"
}

func GetHashString(data []byte, key string) string {
	return base64.URLEncoding.EncodeToString(getHash(data, key))
}

func getHash(data []byte, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return h.Sum(nil)
}

func (hd HashData) WithHashVerification(h http.Handler) http.Handler {
	hashVerificationFunc := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		requestHash := r.Header.Get("HashSHA256")
		if requestHash != "" && hd.Key != "" {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			originalHash := getHash(data, hd.Key)
			decodeHash, err := base64.URLEncoding.DecodeString(requestHash)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !hmac.Equal(originalHash, decodeHash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(hashVerificationFunc)
}
