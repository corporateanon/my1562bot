package main

import (
	"fmt"
	"github.com/corporateanon/my1562bot/pkg/config"
	"github.com/corporateanon/my1562geocoder"
)

func NewGeocoder(conf *config.Config) (*my1562geocoder.Geocoder, error) {
	//TODO: add config param
	geo := my1562geocoder.NewGeocoder("./data/gobs/geocoder-data.gob")
	geo.BuildSpatialIndex(100)
	return geo, nil
}

func FormatShortAddress(res *my1562geocoder.ReverseGeocodingResult) string {
	street := res.FullAddress.Street1562.Name
	building := res.FullAddress.Address.Number
	return fmt.Sprintf("%s %d", street, building)
}
