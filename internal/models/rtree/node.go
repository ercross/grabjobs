package rtree

import (
	"errors"
	"github.com/ercross/grabjobs/internal/models"
	"github.com/umahmood/haversine"
)

// the data to be stored inside leaf of a node.
type entry struct {
	job models.Job

	// mbr is the minimum bounding rectangle to bound this job.Location
	mbr mbr
}

func NewEntry(from models.Job) *entry {
	entry := entry{
		job: from,
		mbr: newMBRAround(from.Location),
	}
	return &entry
}

// insertIntoAny inserts e into any of first or second node
// and expands the node.mbr to which e was inserted.
// The above implies that one of first or second node will be modified.
func (e *entry) insertIntoAny(n1, n2 *node) {

	// insert into node that needs the percentage least expansion
	n1PercentExp := n1.mbr.calculatePercentageExpansion(e.mbr)
	n2PercentExp := n2.mbr.calculatePercentageExpansion(e.mbr)

	if n2PercentExp > n1PercentExp {
		n1.insertEntry(*e)
	}

	if n1PercentExp > n2PercentExp {
		n2.insertEntry(*e)
	}

	// if percentage expansion is same, insert into that with the smallest area
	if n1PercentExp == n2PercentExp {
		if n2.mbr.area() < n1.mbr.area() {
			n2.insertEntry(*e)
		} else {
			n1.insertEntry(*e)
		}
	}
}

type node struct {

	// mbr is the minimum bounding rectangle around
	// the children of this node.
	// If node.isLeaf, mbr is the minimum bounding
	// rectangle around entries of this node
	mbr      mbr
	children []*node
	parent   *node

	// entries is the spatial data to be stored in this node
	entries []*entry
}

// fetchJobs fetches jobs found @within radial distance of center location,
// provided n is a leaf
func (n *node) fetchJobs(within models.Distance, center models.Location) []models.Job {
	jobs := make([]models.Job, 0)
	if !n.isLeaf() {
		return []models.Job{}
	}

	c1 := haversine.Coord{
		Lat: center.Latitude,
		Lon: center.Longitude,
	}
	for _, entry := range n.entries {
		c2 := haversine.Coord{
			Lat: entry.job.Location.Latitude,
			Lon: entry.job.Location.Longitude,
		}
		if d, _ := haversine.Distance(c1, c2); d <= within.Value {
			jobs = append(jobs, entry.job)
		}
	}
	return jobs
}

// snInsertIntoAny inserts n into any of first or second node
// and expands the node.mbr to which n was inserted.
// The above implies that one of first or second node will be modified.
func (n *node) snInsertIntoAny(n1, n2 *node) {

	// insert into node that needs the percentage least expansion
	n1PercentExp := n1.mbr.calculatePercentageExpansion(n.mbr)
	n2PercentExp := n2.mbr.calculatePercentageExpansion(n.mbr)

	if n2PercentExp > n1PercentExp {
		n1.insertChild(n)
	}

	if n1PercentExp > n2PercentExp {
		n2.insertChild(n)
	}

	// if percentage expansion is same, insert into that with the smallest area
	if n1PercentExp == n2PercentExp {
		if n2.mbr.area() < n1.mbr.area() {
			n2.insertChild(n)
		} else {
			n1.insertChild(n)
		}
	}
}

// hasEntrySpace checks that this node has enough space to store one more entry,
func (n *node) hasEntrySpace() bool {
	entriesNewSize := len(n.entries) + 1
	return entriesNewSize <= maxEntriesPerLeaf
}

// hasNodeSpace checks that n has enough space to accommodate 1 more child nodes
func (n *node) hasNodeSpace() bool {
	childrenNewSize := len(n.children) + 2
	return childrenNewSize <= maxEntriesPerLeaf
}

// insertEntry inserts e into n, expanding n.mbr to accommodate e.mbr as well
func (n *node) insertEntry(e entry) {
	n.entries = append(n.entries, &e)
	n.mbr = n.mbr.expandToAccommodate(e.mbr)
}

func (n *node) insertMultipleChildren(children ...*node) {
	if len(children) == 0 {
		return
	}

	for _, e := range children {
		n.insertChild(e)
	}
}

func (n *node) insertMultipleEntry(entries ...*entry) {
	if len(entries) == 0 {
		return
	}

	for _, e := range entries {
		n.insertEntry(*e)
	}
}

func (n *node) insertChild(child *node) {
	child.parent = n
	n.children = append(n.children, child)
	n.mbr = n.mbr.expandToAccommodate(child.mbr)
}

// removeChild node from n.children if found and shrinks n.mbr.
// On return from removeChild, the invoking code may check that
// len(n.children) >= minEntriesPerNode. If this check fails,
// shrink tree if necessary.
func (n *node) removeChild(child *node) {

	// remove from n.children
	for i, c := range n.children {
		if c == child {
			n.children[i].parent = nil
			n.children = append(n.children[:i], n.children[i+1:]...)
			break
		}
	}

	// detach child from n
	child.parent = nil
	n.mbr = n.mbr.shrinkOnRemoval(child.mbr)
}

// splitLeaf node that fails node.hasEntrySpace test on addition of new entry e
// using the Linear-Cost Algorithm as described in
// http://www-db.deis.unibo.it/courses/SI-LS/papers/Gut84.pdf section 3.5.3.
// splitLeaf removes n.entries and distributes them into n1, n2 accordingly.
func (n *node) splitLeaf(e entry) (n1, n2 *node) {
	n.insertEntry(e)
	s1, s2 := n.linearPickSeed()
	n1, n2 = new(node), new(node)
	n1.entries = append(n1.entries, s1)
	n2.entries = append(n2.entries, s2)
	for len(n.entries) != 0 {

		// if any of first or second leaf node have less entries than minEntriesPerLeaf
		// and the n.entries is almost exhausted
		if len(n.entries) < minEntriesPerLeaf && len(n1.entries) < minEntriesPerLeaf && len(n2.entries) > minEntriesPerLeaf {
			n1.insertMultipleEntry(n.entries...)
			break
		}
		if len(n.entries) < minEntriesPerLeaf && len(n2.entries) < minEntriesPerLeaf && len(n1.entries) > minEntriesPerLeaf {
			n2.insertMultipleEntry(n.entries...)
			break
		}

		next := n.pickNext()
		next.insertIntoAny(n1, n2)
	}

	// if n is root, then n1 and n2 parent is nil
	n1.parent = n.parent
	n2.parent = n.parent
	return n1, n2
}

// splitNode splits n on addition of child to n.children, which causes this split.
// It uses the same algorithm as n.splitLeaf
func (n *node) splitNode(child *node) (n1, n2 *node) {
	n.insertChild(child)
	n1, n2 = new(node), new(node)
	s1, s2 := n.snLinearPickSeed()
	n1.children = append(n1.children, s1)
	n2.children = append(n2.children, s2)
	for len(n.children) != 0 {

		// if any of first or second node have fewer children than minEntriesPerLeaf
		// and the n.children is almost exhausted
		if len(n.children) < minEntriesPerLeaf && len(n1.children) < minEntriesPerLeaf && len(n2.children) > minEntriesPerLeaf {
			n1.insertMultipleChildren(n.children...)
			break
		}
		if len(n.children) < minEntriesPerLeaf && len(n2.children) < minEntriesPerLeaf && len(n1.children) > minEntriesPerLeaf {
			n2.insertMultipleChildren(n.children...)
			break
		}

		next := n.snPickNext()
		next.snInsertIntoAny(n1, n2)
	}
	n1.parent = n.parent
	n2.parent = n.parent
	// todo remove n from n.parent.children
	return n1, n2
}

// pickNext removes and returns next entry from n.entries
func (n *node) pickNext() *entry {
	if len(n.entries) == 1 {
		return n.entries[0]
	}

	e := n.entries[0]
	n.entries = n.entries[1:]
	return e
}

// snPickNext removes and returns next child from n.children
func (n *node) snPickNext() *node {
	if len(n.children) == 1 {
		return n.children[0]
	}

	e := n.children[0]
	n.children = n.children[1:]
	return e
}

// snLinearPickSeed is splitNodeLinearPickSeed and works like linearPickSeed,
// but unlike linearPickSeed, it works for splitting non-leaf node
func (n *node) snLinearPickSeed() (s1, s2 *node) {
	if len(n.children) == 2 {
		return n.children[0], n.children[1]
	}

	// find one node for each side that has its
	// 1. maxX closest to node.mbr.minX i.e., leftmost,
	// 2. maxY closest to node.mbr.minY i.e., downmost,
	// 3. minX closest to node.mbr.maxX i.e., rightmost,
	// 4. minY closest to node.mbr.maxY i.e., upmost
	upmost, downmost, rightmost, leftmost := n.children[0], n.children[0], n.children[0], n.children[0]
	var upmostIndex, downmostIndex, rightmostIndex, leftmostIndex int
	for i, child := range n.children {

		// 1. maxX closest to node.mbr.minX
		if (n.mbr.minX - child.mbr.maxX) < leftmost.mbr.maxX {
			leftmost = child
			leftmostIndex = i
		}

		// 2. maxY closest to node.mbr.minY i.e., downmost
		if (n.mbr.minY - child.mbr.maxY) < downmost.mbr.maxY {
			downmost = child
			downmostIndex = i
		}

		// 3. minX closest to node.mbr.maxX i.e., rightmost
		if (n.mbr.maxX - child.mbr.minX) < rightmost.mbr.minX {
			rightmost = child
			rightmostIndex = i
		}

		// 4. minY closest to node.mbr.maxY i.e., upmost
		if (n.mbr.maxY - child.mbr.minY) < upmost.mbr.minY {
			upmost = child
			upmostIndex = i
		}
	}

	// calculate separation space between each opposing entries in each dimension
	var sepAlongY, sepAlongX float64
	if rightmost != leftmost {
		sepAlongY = upmost.mbr.minY - downmost.mbr.maxY
	}
	if leftmost != rightmost {
		sepAlongX = rightmost.mbr.minX - leftmost.mbr.maxX
	}

	normalizedSepAlongY := sepAlongY / n.mbr.height()
	normalizedSepAlongX := sepAlongX / n.mbr.width()

	if normalizedSepAlongY > normalizedSepAlongX {
		n.removeChildren(upmostIndex, downmostIndex)
		return upmost, downmost
	} else {
		n.removeChildren(leftmostIndex, rightmostIndex)
		return rightmost, leftmost
	}
}

// linearPickSeed picks two oppositely positioned entries with
// the greatest separation space/length between them.
// LinearPickSeed deletes s1 and s2 from n.entries
func (n *node) linearPickSeed() (s1, s2 *entry) {

	if len(n.entries) == 2 {
		return n.entries[0], n.entries[1]
	}

	// find one entry for each side that has its
	// 1. maxX closest to node.mbr.minX i.e., leftmost,
	// 2. maxY closest to node.mbr.minY i.e., downmost,
	// 3. minX closest to node.mbr.maxX i.e., rightmost,
	// 4. minY closest to node.mbr.maxY i.e., upmost
	upmost, downmost, rightmost, leftmost := n.entries[0], n.entries[0], n.entries[0], n.entries[0]
	var upmostIndex, downmostIndex, rightmostIndex, leftmostIndex int
	for i, entry := range n.entries {

		// 1. maxX closest to node.mbr.minX
		if (n.mbr.minX - entry.mbr.maxX) < leftmost.mbr.maxX {
			leftmost = entry
			leftmostIndex = i
		}

		// 2. maxY closest to node.mbr.minY i.e., downmost
		if (n.mbr.minY - entry.mbr.maxY) < downmost.mbr.maxY {
			downmost = entry
			downmostIndex = i
		}

		// 3. minX closest to node.mbr.maxX i.e., rightmost
		if (n.mbr.maxX - entry.mbr.minX) < rightmost.mbr.minX {
			rightmost = entry
			rightmostIndex = i
		}

		// 4. minY closest to node.mbr.maxY i.e., upmost
		if (n.mbr.maxY - entry.mbr.minY) < upmost.mbr.minY {
			upmost = entry
			upmostIndex = i
		}
	}

	// calculate separation space between each opposing entries in each dimension
	var sepAlongY, sepAlongX float64
	if rightmost != leftmost {
		sepAlongY = upmost.mbr.minY - downmost.mbr.maxY
	}
	if leftmost != rightmost {
		sepAlongX = rightmost.mbr.minX - leftmost.mbr.maxX
	}

	normalizedSepAlongY := sepAlongY / n.mbr.height()
	normalizedSepAlongX := sepAlongX / n.mbr.width()

	if normalizedSepAlongY > normalizedSepAlongX {
		n.removeEntries(upmostIndex, downmostIndex)
		return upmost, downmost
	} else {
		n.removeEntries(leftmostIndex, rightmostIndex)
		return rightmost, leftmost
	}
}

// removeEntries removes entries at i1 and i2 from entries
func (n *node) removeEntries(i1, i2 int) {
	n.entries = append(n.entries[:i1], n.entries[i1+1:]...)

	// shift index of i2
	if i1 > i2 {
		i2 += 1
	} else {
		i2 -= 1
	}

	n.entries = append(n.entries[:i2], n.entries[i2+1:]...)
}

// removeChildren removes children at i1 and i2 from children
func (n *node) removeChildren(i1, i2 int) {
	n.children = append(n.children[:i1], n.children[i1+1:]...)

	// shift index of i2
	if i1 > i2 {
		i2 += 1
	} else {
		i2 -= 1
	}

	n.children = append(n.children[:i2], n.children[i2+1:]...)
}

// isLeaf checks that node is a leaf.
// RTree property:: All leaves appear on the same level
func (n *node) isLeaf() bool {
	return len(n.children) == 0 && len(n.entries) != 0
}

func (n *node) isParentToLeafNode() bool {

	if n.isLeaf() {
		return false
	}

	if len(n.children[0].entries) != 0 {
		return true
	}

	return false
}

// findMBRNeedingLeastExpansion finds the child node that its mbr needs the least
// expansion toAccommodate new entry.
// If len(children) == 0, program panics as this is a programmer error
func findMBRNeedingLeastExpansion(toAccommodate entry, children []*node) *node {
	if len(children) == 0 {
		panic(errors.New("zero length children passed to findMBRNeedingLeastExpansion"))
	}
	leastExpansionPercentage := 0.00
	leastExpansion := children[0]
	for _, child := range children {
		expansionPercentage := child.mbr.calculatePercentageExpansion(toAccommodate.mbr)

		if expansionPercentage < leastExpansionPercentage {
			leastExpansionPercentage = expansionPercentage
			leastExpansion = child
			continue
		}

		// if any two nodes can have same leastExpansionPercentage,
		// pick the one with the smallest area
		if expansionPercentage == leastExpansionPercentage {
			if child.mbr.area() < leastExpansion.mbr.area() {
				leastExpansion = child
			}
		}
	}

	return leastExpansion
}
