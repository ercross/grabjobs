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

// findMBRNeedingLeastExpansion finds the node whose node.mbr needs the least
// expansion to accommodate new location.
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

		// check if toAccomodate can fit into this child node
		// todo complete below
		//if childNode.mbr.longitudeLowerBound <= toAccommodate.Longitude <= childNode.mbr.longitudeUpperBound

		expansionDownwards := childNode.mbr.longitudeLowerBound - toAccommodate.Longitude
		expansionUpwards := childNode.mbr.longitudeUpperBound - toAccommodate.Longitude
		expansionRightwards := childNode.mbr.latitudeUpperBound - toAccommodate.Latitude
		expansionLeftwards := childNode.mbr.latitudeLowerBound - toAccommodate.Latitude

		// initialized in case any opposite sides are equal
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

		// check for
		if leastLongitudeExpansion < math.Abs(leastLongitude.mbr.longitudeLowerBound) ||
			leastLongitudeExpansion < math.Abs(leastLongitude.mbr.longitudeUpperBound) {
			leastLongitude = childNode
		}

		if leastLatitudeExpansion < math.Abs(leastLatitude.mbr.latitudeLowerBound) ||
			leastLatitudeExpansion < math.Abs(leastLatitude.mbr.latitudeUpperBound) {
			leastLatitude = childNode
		}
	}

	if leastLatitude == leastLongitude {
		return leastLongitude
	}
	return nil // todo
}
