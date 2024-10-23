package service

import (
	"testing"

	"crypto/hmac"
	"crypto/sha256"

	"github.com/stretchr/testify/assert"
)

func TestNewHashData(t *testing.T) {
	tests := []struct {
		name    string
		hashKey string
		want    error
	}{
		{
			name:    "1",
			hashKey: "test",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashData := NewHashData(tt.hashKey)
			assert.Equal(t, hashData.Key, tt.hashKey)
		})
	}
}

func TestMakeURL(t *testing.T) {
	tests := []struct {
		name    string
		runAddr string
		method  string
		mtype   string
		mname   string
		mvalue  string
		want    string
	}{
		{
			name:    "1",
			runAddr: "localhost",
			method:  "/update/",
			mtype:   "type",
			mname:   "name",
			mvalue:  "0",
			want:    "http://localhost/update/type/name/0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := MakeURL(tt.runAddr, tt.method, tt.mtype, tt.mname, tt.mvalue)
			assert.Equal(t, url, tt.want)
		})
	}
}

func TestMakeBatchUpdatesURL(t *testing.T) {
	tests := []struct {
		name    string
		runAddr string
		want    string
	}{
		{
			name:    "1",
			runAddr: "localhost",
			want:    "http://localhost/updates/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := MakeBatchUpdatesURL(tt.runAddr)
			assert.Equal(t, url, tt.want)
		})
	}
}

func TestGtHash(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  string
	}{
		{
			name: "1",
			data: []byte("test"),
			key:  "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := hmac.New(sha256.New, []byte(tt.key))
			h.Write(tt.data)
			hsh := getHash(tt.data, tt.key)
			assert.Equal(t, hsh, h.Sum(nil))
		})
	}
}
