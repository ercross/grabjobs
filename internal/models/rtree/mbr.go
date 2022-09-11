package rtree

import (
	"github.com/ercross/grabjobs/internal/models"
	"math"
)

// mbr is a minimum bounding rectangle for a 2D spatial data.
// mbr is constructed as if on a cartesian coordinate.
type mbr struct {

	// higher longitudinal value
	maxY float64

	// lower longitudinal value
	minY float64

	// higher latitude value
	maxX float64

	// lower latitude value
	minX float64
}

// arbitrary MBR dimension factor is a random factor
// to add to a location in order to draw an mbr over the location
const arbitraryMBRDimFactor = 0.2

// newMBR returns a new minimum bounding rectangle around this location.
// RTree property:: For each entry in a leaf node, an MBR should exist to spatially contain
// the 2D location object
func newMBRAround(location models.Location) mbr {
	return mbr{
		maxY: location.Longitude + arbitraryMBRDimFactor,
		minY: location.Longitude - arbitraryMBRDimFactor,
		maxX: location.Latitude + arbitraryMBRDimFactor,
		minX: location.Latitude - arbitraryMBRDimFactor,
	}
}

func (m mbr) area() float64 {
	return math.Abs(m.height() * m.width())
}

func (m mbr) height() float64 {
	return m.maxY - m.minY
}

func (m mbr) width() float64 {
	return m.maxX - m.minX
}

// overlapsWith checks that m and other overlaps.
func (m mbr) overlapsWith(other mbr) bool {
	// two rectangles do not overlap if
	// 1. either of the rectangles area is zero
	if m.area() == 0 || other.area() == 0 {
		return false
	}

	// 2. one rectangle is on the left side of the other
	if m.minX > other.maxX || other.minX > m.maxX {
		return false
	}

	// 3. one rectangle is above the other
	if m.minY > other.maxY || other.minY > m.maxY {
		return false
	}

	return true
}

// canFitWithin checks if m can fit inside other without expanding other
func (m mbr) canFitWithin(other mbr) bool {
	if m.area() < other.area() {
		return false
	}

	fitsWithinWidth := other.minX < m.minX && other.maxX > m.maxX
	fitsWithinHeight := other.minY < m.minY && other.maxY > m.maxY

	if fitsWithinWidth && fitsWithinHeight {
		return true
	}

	return false
}

// calculatePercentageExpansion calculates the percentage area expansion
// needed for m to accommodate child.
func (m mbr) calculatePercentageExpansion(child mbr) float64 {
	if child.area() > m.area() {
		return math.Inf(1)
	}

	expanded := m.expandToAccommodate(child)
	return (expanded.area() * 100) / m.area()
}

func (m mbr) expandToAccommodate(child mbr) mbr {
	// find potential expansion in all directions of this minimum bounding rectangle.
	// Initialize new expanded mbr to current mbr
	expanded := m

	// potential expansion leftwards
	if child.minX < m.minX {
		expanded.minX = child.minX
	}

	// potential expansion rightwards
	if child.maxX > m.maxX {
		expanded.maxX = child.maxX
	}

	// potential expansion downwards
	if child.minY < m.minY {
		expanded.minY = child.minY
	}

	// potential expansion upwards
	if child.maxY > m.maxY {
		expanded.maxY = child.maxY
	}
	return expanded
}

func (m mbr) shrinkOnRemoval(child mbr) mbr {
	// find potential decrease in all directions of this minimum bounding rectangle.
	// Initialize new expanded mbr to current mbr
	shrunk := m

	// potential shrink leftwards
	if child.minX == m.minX {
		shrunk.minX = shrunk.minX - child.minX
	}

	// potential expansion rightwards
	if child.maxX == m.maxX {
		shrunk.maxX = shrunk.maxX - child.maxX
	}

	// potential expansion downwards
	if child.minY == m.minY {
		shrunk.minY = shrunk.minY - child.minY
	}

	// potential expansion upwards
	if child.maxY > m.maxY {
		shrunk.maxY = shrunk.maxY - child.maxY
	}
	return shrunk
}
