package earth_test

import (
	"testing"

	"github.com/zxcv32/capture-maps/pkg/earth"
)

// Reference tile coordinates per Google's documented Mercator projection:
// https://developers.google.com/maps/documentation/javascript/coordinates
func TestCalcCentreTile(t *testing.T) {
	tests := []struct {
		name         string
		lat, lng     float64
		zoom         int
		wantX, wantY int
		wantScale    int
	}{
		{"origin at zoom 0 is the only tile", 0, 0, 0, 0, 0, 1},
		{"origin at zoom 1 is the centre of a 2x2 grid", 0, 0, 1, 1, 1, 2},
		{"west+south of origin stays in lower-left tile at zoom 1", -10, -10, 1, 0, 1, 2},
		{"north pole clamp stays within valid tile range", 85, 0, 2, 2, 0, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY, gotScale := earth.CalcCentreTile(tt.lat, tt.lng, tt.zoom)
			if gotX != tt.wantX || gotY != tt.wantY || gotScale != tt.wantScale {
				t.Fatalf("CalcCentreTile(%v, %v, %v) = (%d, %d, %d), want (%d, %d, %d)",
					tt.lat, tt.lng, tt.zoom, gotX, gotY, gotScale, tt.wantX, tt.wantY, tt.wantScale)
			}
			if gotX < 0 || gotX >= gotScale {
				t.Fatalf("tileX %d out of valid range [0,%d)", gotX, gotScale)
			}
			if gotY < 0 || gotY >= gotScale {
				t.Fatalf("tileY %d out of valid range [0,%d)", gotY, gotScale)
			}
		})
	}
}
