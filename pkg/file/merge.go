package file

import (
	gim "github.com/ozankasikci/go-image-merge"
	"image/png"
	"os"
)

func Merge(grids []*gim.Grid, x int, y int) (string, error) {
	rgba, err := gim.New(grids, x, y).Merge()
	if err != nil {
		return "", err
	}
	filename := generateFilename()
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if err := png.Encode(file, rgba); err != nil {
		return "", err
	}
	return filename, nil
}
