package main

import (
	"fmt"

	"github.com/my1562/geocoder"
	"github.com/my1562/telegrambot/pkg/config"
)

func NewGeocoder(conf *config.Config) (*geocoder.Geocoder, error) {
	//TODO: add config param
	geo := geocoder.NewGeocoder("./data/gobs/geocoder-data.gob")
	geo.BuildSpatialIndex(100)
	return geo, nil
}

func FormatGeocodingResult(res *geocoder.ReverseGeocodingResult) string {
	street := res.FullAddress.Street1562.Name
	building := res.FullAddress.Address.Number
	return fmt.Sprintf("%s %d", street, building)
}
