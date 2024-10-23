package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		locallink Locallink
	}{
		{
			name:      "1",
			locallink: Locallink{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.locallink.Run()
			assert.Equal(t, nil, err)
		})
	}
}
