package service

import "testing"

func TestEffectiveRowRadius(t *testing.T) {
	tests := []struct {
		name                       string
		centreTileY, scale, radius int
		want                       int
	}{
		{"plenty of room keeps radius unchanged", 500, 1024, 5, 5},
		{"near top edge clamps to available rows", 1, 1024, 5, 2},
		{"near bottom edge clamps to available rows", 1022, 1024, 5, 2},
		{"single-tile world always returns at least 1", 0, 1, 5, 1},
		{"centre tile exactly on the edge still returns 1", 0, 1024, 5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := effectiveRowRadius(tt.centreTileY, tt.scale, tt.radius)
			if got != tt.want {
				t.Fatalf("effectiveRowRadius(%d, %d, %d) = %d, want %d",
					tt.centreTileY, tt.scale, tt.radius, got, tt.want)
			}
			if got < 1 {
				t.Fatalf("effectiveRowRadius returned %d, must be >= 1", got)
			}
		})
	}
}
