package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	gim "github.com/ozankasikci/go-image-merge"
	log "github.com/sirupsen/logrus"
	"github.com/zxcv32/capture-maps/pkg/earth"
	"github.com/zxcv32/capture-maps/pkg/file"
)

// maxRadius bounds the worst-case tile fan-out per request to
// (2*maxRadius-1)^2 = 81 outbound Google Maps Static API calls
// (~$0.16 at Google's $2/1000 rate) instead of the previous 841.
const maxRadius = 5

// maxRequestBodyBytes caps the incoming JSON body; the real payload is a
// handful of numbers and a short string, so this is generous headroom.
const maxRequestBodyBytes = 4 << 10 // 4KiB

// printRequest struct to accept print request
type printRequest struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Zoom      int     `json:"zoom"`
	Radius    int     `json:"radius"`
	MapTypeId string  `json:"mapTypeId"`
}

// setupResponse setups all the common HTTP response headers
func setupResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Content-Type", "image/png")
}

// HandleRequest print request
func PrintHandler(apiKey string) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		setupResponse(&writer)
		if request.Method == "OPTIONS" {
			return
		}
		var requestId = rand.Int()
		log.Printf("Print request received: %d", requestId)
		body := http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
		decoder := json.NewDecoder(body)
		var task printRequest
		err := decoder.Decode(&task)
		if err != nil {
			log.Errorln(err)
			writer.WriteHeader(500)
			log.Printf("Print request not complete: %d", requestId)
			return
		}

		lat, lng, zoom, radius, mapTypeId := task.Lat, task.Lng, task.Zoom, task.Radius, task.MapTypeId
		validationError := validateRequest(lat, lng, zoom, radius, mapTypeId)
		if validationError != nil {
			http.Error(writer, validationError.Error(), 400)
			return
		}
		filename, err := captureTiles(lat, lng, zoom, radius, mapTypeId, apiKey)
		if err != nil {
			log.Errorln(err)
			http.Error(writer, "internal error", 500)
			log.Printf("Print request not complete: %d", requestId)
			return
		}
		fileBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Errorln(err)
			http.Error(writer, "internal error", 500)
			log.Printf("Print request not complete: %d", requestId)
			return
		}
		writer.WriteHeader(http.StatusOK)
		_, error := writer.Write(fileBytes)
		if error != nil {
			http.Error(writer, error.Error(), 500)
			return
		}
		file.DeleteFile(filename)
		log.Printf("Print request complete: %d", requestId)
	}
}

func validateRequest(lat float64, lng float64, zoom int, radius int, mapTypeId string) error {
	if lat < -90 || lat > 90 {
		return errors.New("invalid latitude. Possible set: [-90,90]")
	}
	if lng < -180 || lng > 180 {
		return errors.New("invalid longitude. Possible set: [-180,80]")
	}
	if zoom < 0 || zoom > 21 {
		return errors.New("invalid zoom. Possible set: [0,21]")
	}
	if radius < 1 || radius > maxRadius {
		return fmt.Errorf("invalid radius. Possible set: [1,%d]", maxRadius)
	}
	switch mapTypeId {
	case
		"hybrid",
		"satellite",
		"roadmap",
		"terrain":
	default:
		return errors.New("invalid mapTypeId. Possible values: hybrid,satellite,roadmap,terrain")
	}
	return nil
}

// effectiveRowRadius bounds how many tile rows north/south of centreTileY
// are requested, so a large radius at low zoom (or near the poles, where a
// row index can run out of [0,scale) tile rows quickly) can't produce
// duplicate or wrapped rows. The centre row itself is always valid.
func effectiveRowRadius(centreTileY int, scale int, radius int) int {
	maxRows := min(centreTileY, scale-1-centreTileY) + 1
	if maxRows < 1 {
		maxRows = 1
	}
	return min(radius, maxRows)
}

func captureTiles(lat float64, lng float64, zoom int, radius int, mapTypeId string, apiKey string) (string, error) {
	centreTileX, centreTileY, scale := earth.CalcCentreTile(lat, lng, zoom)
	rowRadius := effectiveRowRadius(centreTileY, scale, radius)
	var grids []*gim.Grid
	latitudinalTiles := min(radius*2-1, scale)
	var xIndex = centreTileX - (radius - 1)
	lastXIndexTile := latitudinalTiles + xIndex
	var j = centreTileY
	var gimGridCentre gim.Grid
	path, err := westToEast(xIndex, j, lastXIndexTile, zoom, mapTypeId, apiKey)
	if err != nil {
		return "", err
	}
	gimGridCentre.ImageFilePath = path
	j = centreTileY - rowRadius
	repeat := rowRadius - 1
	for repeat > 0 {
		repeat--
		j++ // go from north to south towards centre
		var gimGridNorth gim.Grid
		path, err := westToEast(xIndex, j, lastXIndexTile, zoom, mapTypeId, apiKey)
		if err != nil {
			file.DeleteFiles(grids)
			file.DeleteFile(gimGridCentre.ImageFilePath)
			return "", err
		}
		gimGridNorth.ImageFilePath = path
		grids = append(grids, &gimGridNorth)
	}

	grids = append(grids, &gimGridCentre)

	j = centreTileY // reset j
	repeat = rowRadius - 1
	for repeat > 0 {
		repeat--
		j++ // go south
		var gimGridSouth gim.Grid
		path, err := westToEast(xIndex, j, lastXIndexTile, zoom, mapTypeId, apiKey)
		if err != nil {
			file.DeleteFiles(grids)
			return "", err
		}
		gimGridSouth.ImageFilePath = path
		grids = append(grids, &gimGridSouth)
	}

	flashbang, err := file.Merge(grids, 1, rowRadius*2-1)
	file.DeleteFiles(grids)
	if err != nil {
		return "", err
	}
	return flashbang, nil
}

func westToEast(xIndex int, j int, lastXIndexTile int, zoom int, mapTypeId string, apiKey string) (string, error) {
	var grids []*gim.Grid
	// West to east loop
	for i := xIndex; i < lastXIndexTile; i++ {
		tileLat := earth.Tile2lat(float64(j), zoom)
		tileLng := earth.Tile2long(float64(i), zoom)
		filename, err := earth.ComputeImage(tileLat, tileLng, zoom, mapTypeId, apiKey)
		if err != nil {
			file.DeleteFiles(grids)
			return "", err
		}
		var gimGrid gim.Grid
		gimGrid.ImageFilePath = filename
		grids = append(grids, &gimGrid)
	}
	merged, err := file.Merge(grids, len(grids), 1)
	file.DeleteFiles(grids)
	if err != nil {
		return "", err
	}
	return merged, nil
}

func min(a int, b int) int {
	min := a
	if b < min {
		min = b
	}
	return min
}
