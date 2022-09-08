package rtree

import (
	"github.com/ercross/grabjobs/internal/models"
	"math"
)

type node struct {
	mbr      mbr
	children []*node

	// entries is the spatial data to be stored in this node
	entries []*models.Job
}

// mbr is a minimum bounding rectangle for a 2D spatial data.
// mbr is constructed as if on a cartesian coordinate.
type mbr struct {

	// highest value towards positive y-axis
	longitudeUpperBound float64

	// lowest value towards negative y-axis
	longitudeLowerBound float64

	// highest value towards positive x-axis
	latitudeUpperBound float64

	// highest value towards negative x-axis
	latitudeLowerBound float64
}

// arbitrary MBR dimension factor is a random factor
// to add to a location in order to draw an mbr over the location
const arbitraryMBRDimFactor = 0.2

// newMBR returns a new minimum bounding rectangle around this location.
// RTree property:: For each entry in a leaf node, an MBR should exist to spatially contain
// the 2D location object
func newMBRAround(entry models.Location) mbr {
	return mbr{
		longitudeUpperBound: entry.Longitude + arbitraryMBRDimFactor,
		longitudeLowerBound: entry.Longitude - arbitraryMBRDimFactor,
		latitudeUpperBound:  entry.Latitude + arbitraryMBRDimFactor,
		latitudeLowerBound:  entry.Latitude - arbitraryMBRDimFactor,
	}
}

// hasEntrySpace checks that this node has enough space to store one more entry
func (nod *node) hasEntrySpace() bool {
	entriesNewSize := len(nod.entries) + 1
	return entriesNewSize > maxEntriesPerNode
}

func (nod *node) insertNewEntry(entry models.Job) {

}

func (nod *node) split() (newOne node) {
	return *nod
}

// isLeaf checks that node is a leaf.
// RTree property:: All leaves appear on the same level
func (nod *node) isLeaf() bool {
	return len(nod.children) == 0
}

// findMBRNeedingLeastExpansion naively finds the node whose node.mbr needs the least
// expansion toAccommodate new location.
// Note that this implementation is a naive implementation and should be reimplemented
// by combining mbr area property and set unionism.
// Note that len(node.children) must be greater than zero
func (nod *node) findMBRNeedingLeastExpansion(toAccommodate models.Location) *node {
	if len(nod.children) == 0 {

		// don't deserve this but just crash the whole thing.
		// programmer error
		panic(nil)
	}

	if len(nod.children) == 1 {
		return nod.children[0]
	}

	// obtain nodes needing least latitude and longitude expansions
	var leastLatitude, leastLongitude *node

	for _, childNode := range nod.children {

		// check if toAccomodate can fit directly into this child node
		if childNode.mbr.longitudeLowerBound <= toAccommodate.Longitude &&
			toAccommodate.Longitude <= childNode.mbr.longitudeUpperBound &&
			childNode.mbr.latitudeLowerBound <= toAccommodate.Latitude &&
			toAccommodate.Latitude <= childNode.mbr.latitudeUpperBound {

			return childNode
		}

		// check how location toAccommodate will cause childNode to expand in all directions
		expansionDownwards := childNode.mbr.longitudeLowerBound - toAccommodate.Longitude
		expansionUpwards := childNode.mbr.longitudeUpperBound - toAccommodate.Longitude
		expansionRightwards := childNode.mbr.latitudeUpperBound - toAccommodate.Latitude
		expansionLeftwards := childNode.mbr.latitudeLowerBound - toAccommodate.Latitude

		// initialized to arbitrary values in case any opposite sides are equal
		leastLongitudeExpansion, leastLatitudeExpansion := expansionUpwards, expansionRightwards

		// check for least expansion on longitude
		if math.Abs(expansionDownwards) < math.Abs(expansionUpwards) {
			leastLongitudeExpansion = expansionDownwards
		} else {
			leastLongitudeExpansion = expansionUpwards
		}

		// check for least expansion on latitude
		if math.Abs(expansionLeftwards) < math.Abs(expansionRightwards) {
			leastLatitudeExpansion = expansionLeftwards
		} else {
			leastLatitudeExpansion = expansionRightwards
		}

		// compare this childNode.mbr longitudinal expansion to existing nodes needing
		// the least longitudinal expansion i.e., leastLongitude
		if leastLongitudeExpansion < math.Abs(leastLongitude.mbr.longitudeLowerBound) ||
			leastLongitudeExpansion < math.Abs(leastLongitude.mbr.longitudeUpperBound) {
			leastLongitude = childNode
		}

		// compare this childNode.mbr latitude expansion to existing nodes needing
		// the least latitude expansion i.e., leastLatitude
		if leastLatitudeExpansion < math.Abs(leastLatitude.mbr.latitudeLowerBound) ||
			leastLatitudeExpansion < math.Abs(leastLatitude.mbr.latitudeUpperBound) {
			leastLatitude = childNode
		}
	}

	// if nodes needing leastLatitude expansion and leastLongitude expansion
	// both points to the same node, return either
	if leastLatitude == leastLongitude {
		return leastLongitude
	}

	// check which of leastLatitude and leastLongitude needs the least expansion
	// on both their adjacent side
	return leastLatitude
}
