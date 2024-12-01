package service

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"musthave-metrics/internal/crypt"
	"musthave-metrics/internal/logger"
)

type HashData struct {
	Key string
}

type hashResponseWriter struct {
	http.ResponseWriter
	HashData HashData
}

type KeyData struct {
	PrivateKeyPath string
}

func (r *hashResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	if r.HashData.Key != "" {
		hash := GetHashString(b, r.HashData.Key)
		r.ResponseWriter.Header().Set("HashSHA256", hash)
	}
	return r.ResponseWriter.Write(b)
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
		requestHash := r.Header.Get("HashSHA256")
		if requestHash != "" && hd.Key != "" {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Warnf("HashSHA256 error: " + err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			originalHash := getHash(data, hd.Key)
			decodeHash, err := base64.URLEncoding.DecodeString(requestHash)
			if err != nil {
				logger.Warnf("HashSHA256 error: " + err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !hmac.Equal(originalHash, decodeHash) {
				logger.Warnf("HashSHA256 error: hash not equal")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(data))
		}
		hw := hashResponseWriter{
			ResponseWriter: w,
			HashData:       hd,
		}
		h.ServeHTTP(&hw, r)
	}
	return http.HandlerFunc(hashVerificationFunc)
}

func NewKeyData(privateKeyPath string) *KeyData {
	return &KeyData{
		PrivateKeyPath: privateKeyPath,
	}
}

func (kd KeyData) WithEncrypt(h http.Handler) http.Handler {
	decryptFunc := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// decrypt request body
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// decrypt only non-empty data
		if len(data) > 0 {

			/*
				// test encrypt
				encryptedBody, err := crypt.Encrypt(
					"D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key.pub",
					string(data),
				)
				decryptBody, err := crypt.Decrypt(kd.PrivateKeyPath, encryptedBody)
			*/

			decryptBody, err := crypt.Decrypt(kd.PrivateKeyPath, string(data))

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// возвращаем тело запроса
			r.Body = io.NopCloser(bytes.NewReader([]byte(decryptBody)))
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(decryptFunc)
}
