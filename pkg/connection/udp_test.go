package connection

import (
	"context"
	"testing"
)

func TestNewUDP(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "valid connection",
			uri:     "localhost:8080",
			wantErr: false,
		},
		{
			name:    "invalid uri",
			uri:     "invalid:uri:format",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			conn, err := NewUDP(ctx, tt.uri)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				if conn != nil {
					t.Error("expected nil connection but got value")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if conn == nil {
				t.Error("expected connection but got nil")
			}

			if udp, ok := conn.(*UDP); ok {
				if udp.addr == nil {
					t.Error("UDP address should not be nil")
				}
				if udp.conn == nil {
					t.Error("UDP connection should not be nil")
				}
			} else {
				t.Error("expected UDP connection type")
			}

			if conn != nil {
				conn.Close(ctx)
			}
		})
	}
}
func TestUDPWrite(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid write",
			data:    map[string]string{"test": "data"},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "cancelled context",
			data: "test",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
		},
		{
			name:    "invalid json data",
			data:    make(chan int),
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := NewUDP(context.Background(), "localhost:8080")
			if err != nil {
				t.Fatalf("failed to create UDP connection: %v", err)
			}
			defer conn.Close(context.Background())

			err = conn.Write(tt.ctx, tt.data)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUDPRead(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := NewUDP(context.Background(), "localhost:8080")
			if err != nil {
				t.Fatalf("failed to create UDP connection: %v", err)
			}
			defer conn.Close(context.Background())

			data, err := conn.Read(tt.ctx)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && data == nil {
				t.Error("expected data but got nil")
			}
		})
	}
}
