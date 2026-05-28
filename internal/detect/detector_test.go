package detect

import "testing"

func TestStatusOnAir(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"idle", Status{}, false},
		{"mic only", Status{MicActive: true}, true},
		{"camera only", Status{CameraActive: true}, true},
		{"both", Status{MicActive: true, CameraActive: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.OnAir(); got != tt.want {
				t.Fatalf("OnAir() = %t, want %t", got, tt.want)
			}
		})
	}
}
