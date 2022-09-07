package models

import "fmt"

// Location is a 2D representation of a place
// on a map
type Location struct {
	Longitude float64
	Latitude  float64
}

// String represents this Location as a string of Longitude,Latitude
func (l Location) String() string {
	return fmt.Sprintf("%f,%f", l.Longitude, l.Latitude)
}

// GeoJSON represents this Location in GeoJSON specification RCF7946.
// https://geojson.org/
func (l Location) GeoJSON() {
	// todo
}
