//go:build windows

package windows

import "testing"

func TestCameraActiveFromTimestamps(t *testing.T) {
	tests := []struct {
		name  string
		start uint64
		stop  uint64
		want  bool
	}{
		{"idle", 100, 100, false},
		{"active", 200, 100, true},
		{"stopped", 100, 200, false},
		{"never stopped", 50, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cameraActiveFromTimestamps(tt.start, tt.stop); got != tt.want {
				t.Fatalf("cameraActiveFromTimestamps(%d, %d) = %t, want %t", tt.start, tt.stop, got, tt.want)
			}
		})
	}
}
