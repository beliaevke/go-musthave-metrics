package main

import (
	"musthave-metrics/cmd/server/config"
	"net/http"
	"reflect"
	"testing"
)

func TestProfiler(t *testing.T) {
	tests := []struct {
		name string
		want http.Handler
	}{
		{
			name: "1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Profiler(); reflect.DeepEqual(got, tt.want) {
				t.Errorf("Profiler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateHandler(t *testing.T) {
	type args struct {
		cfg config.ServerFlags
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		{
			name: "1",
			args: args{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateHandler(tt.args.cfg); reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
