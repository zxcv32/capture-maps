package file

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	gim "github.com/ozankasikci/go-image-merge"
)

func TestMerge_NoGrids_ReturnsError(t *testing.T) {
	filename, err := Merge(nil, 1, 1)
	if err == nil {
		t.Fatal("want error merging zero grids, got nil")
	}
	if filename != "" {
		t.Fatalf("want empty filename on error, got %q", filename)
	}
}

func TestMerge_Success(t *testing.T) {
	tileDir := t.TempDir()
	tilePath := filepath.Join(tileDir, "tile.png")
	writeSolidPNG(t, tilePath, 2, 2, color.RGBA{R: 255, A: 255})

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	filename, err := Merge([]*gim.Grid{{ImageFilePath: tilePath}}, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer os.Remove(filename)

	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("merged file %q not created: %v", filename, err)
	}
}

func writeSolidPNG(t *testing.T, path string, w, h int, c color.Color) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
