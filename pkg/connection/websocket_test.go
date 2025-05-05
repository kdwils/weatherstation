package connection

import (
	"context"
	"testing"

	"github.com/coder/websocket"
)

func TestNewWebsocket(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		opts    *websocket.DialOptions
		wantErr bool
	}{
		{
			name:    "invalid address returns error",
			addr:    "invalid-addr",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "empty address returns error",
			addr:    "",
			opts:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := NewWebsocket(ctx, tt.addr, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWebsocket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("NewWebsocket() returned nil but expected a connection")
			}
		})
	}
}
