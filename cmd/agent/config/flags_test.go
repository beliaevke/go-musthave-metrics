// Package config предназначен для методов конфигурации.
package config

import (
	"reflect"
	"testing"
)

func Test_readConfig(t *testing.T) {
	tests := []struct {
		name string
		want ClientFlags
	}{
		{
			name: "1",
			want: ClientFlags{FlagRunAddr: "localhost:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readConfig(); !reflect.DeepEqual(got.FlagRunAddr, tt.want.FlagRunAddr) {
				t.Errorf("readConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
