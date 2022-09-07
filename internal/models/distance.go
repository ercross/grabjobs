package models

type DistanceUnit int

const (
	UnknownUnit DistanceUnit = iota

	Kilometer
)

type Distance struct {
	Unit  DistanceUnit
	Value float64
}
