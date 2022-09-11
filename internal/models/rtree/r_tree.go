package rtree

import (
	"errors"
	"github.com/ercross/grabjobs/internal/models"
	"github.com/umahmood/haversine"
	"math"
)

// maxEntriesPerLeaf is the maximum branching factor
// or the maximum number of entries a leaf node can contain
// before it is split. No leaf node is an exception.
// It is also the maximum number of node.children a node can
// accommodate before undergoing a split
const maxEntriesPerLeaf = 30

// minEntriesPerLeaf is the minimum branching factor
// or the minimum number of entries a leaf node must contain
// before it can be a standalone leaf node.
// Only the root node (if it's a leaf) can have lesser than minEntriesPerLeaf.
// It is also the minimum number of node.children a node must
// have to be qualified as a node
const minEntriesPerLeaf = maxEntriesPerLeaf / 5

// RTree to efficiently search spatial data.
// Ref: http://www-db.deis.unibo.it/courses/SI-LS/papers/Gut84.pdf
// Use any of NewWithEntry or NewWithEntries to obtain a new tree
type RTree struct {

	// height is the height/depth of the tree.
	// root is at height 0 (zero) and it increases further down
	height     int
	indexCount int
	totalNodes int

	// root node may be a leaf if it's the only node on the tree.
	// RTree property:: If root is not a leaf, then it must have at least 2 children.
	// RTree property:: If root is a leaf, it can contain any number of entries less than maxEntriesPerLeaf
	root *node
}

// NewWithEntry initializes a new node with an entry
func NewWithEntry(e entry) *RTree {
	leaf := &node{
		mbr:     newMBRAround(e.job.Location),
		entries: []*entry{&e},
	}

	tree := RTree{
		height:     0,
		indexCount: 1,
		totalNodes: 1,
		root:       leaf,
	}
	return &tree
}

func NewWithEntries(entries ...models.Job) (*RTree, error) {
	if len(entries) == 0 {
		return nil, errors.New("cannot initialize tree with empty entries")
	}

	tree := NewWithEntry(*NewEntry(entries[0]))

	for _, entry := range entries[1:] {
		tree.Insert(*NewEntry(entry))
	}
	return tree, nil
}

// FindJobs finds job within radial distance of center location.
func (tree *RTree) FindJobs(within models.Distance, center models.Location, d map[string][]models.Job) []models.Job {
	// *********** Current implementation *************
	// FindJobs fetches all entries that fall in ancestral/sibling relationship with center on the tree,
	// iterate through each entry to get the haversine distance.
	// Any distance that does not fall within @within, is not included in result.
	// Haversine thinks of the earth as ellipsoid using the Great Circle distance
	// ref: https://en.wikipedia.org/wiki/Haversine_formula
	//
	// ********** Alternate implementation *************
	// Draw circle around center using Vicenty's formulae
	// (ref: https://en.wikipedia.org/wiki/Vincenty%27s_formulae).
	// The resulting circle is tightly fitted inside a mbr,
	// and the mbr is used to query tree

	job := models.Job{
		Title:    "",
		Location: center,
	}
	entry := NewEntry(job)

	jobs := make([]models.Job, 0)
	if !entry.mbr.canFitWithin(tree.root.mbr) {
		return jobs
	}

	if tree.root.isLeaf() {
		return tree.root.fetchJobs(within, center)
	}
	return search(within, center, d)
}

// Insert a new job into the tree.
// New index records are added at the leaves and nodes that overflow(i.e., len(node.children)>M) are splitLeaf.
func (tree *RTree) Insert(e entry) {
	leaf := tree.chooseLeaf(e)
	if leaf.hasEntrySpace() {
		leaf.insertEntry(e)
		tree.adjustParentOf(leaf)
	} else {
		l1, l2 := leaf.splitLeaf(e)
		tree.adjustParentOnSplitOf(leaf, l1, l2)
	}
}

func (tree *RTree) grow() {
	tree.height += 1
}

// adjustParentOnSplitOf adjusts tree when there is a split on potential
// addition of node to node.parent. l1 and l2 contains initial entries/children of node,
// distributed between them.
// The adjustment is propagated up the tree,
// updating the n1 and n2 parent node.mbr to expand and accommodate n1.mbr, n2.mbr
func (tree *RTree) adjustParentOnSplitOf(node, n1, n2 *node) {

	// If node is root, set n1, n2 as its children and grow the tree.
	// Then, len(tree.root.children) will be lesser than minEntriesPerNode.
	// This is allowed for root node i.e., root can have children lesser than minEntriesPerNode
	if node == tree.root {
		n1.parent = node
		n2.parent = node
		node.insertChild(n1)
		node.insertChild(n2)
		node.entries = nil
		tree.root = node
		tree.grow()
		return
	}

	// remove node from parent node.children, then add n1 and n2 as new children
	parent := node.parent

	// make space for n1 below. It's not necessary to
	// shrink tree here as suggested by node.removeChild
	// because an insertion follows immediately
	parent.removeChild(node)
	parent.insertChild(n1)
	// check that parent can take one more node, else split parent
	if parent.hasNodeSpace() {
		parent.insertChild(n2)
	} else {
		n1, n2 := parent.splitNode(n2)
		tree.adjustParentOnSplitOf(parent, n1, n2)
	}
}

// adjustParentOf adjusts the tree if no split was done after inserting a new entry
func (tree *RTree) adjustParentOf(node *node) {
	if node == tree.root {
		return
	}

	// update the dimension of all ancestor nodes' mbr
	node.parent.mbr = node.parent.mbr.expandToAccommodate(node.mbr)
	tree.adjustParentOf(node.parent)
}

// chooseLeaf chooses a leaf on the tree suitable to accommodate the new entry e.
// Note that the resulting leaf can be root
func (tree *RTree) chooseLeaf(e entry) *node {

	// if root is leaf, choose root
	if tree.root.isLeaf() {
		return tree.root
	}

	// if tree is not leaf, then walkDown the tree
	// to find the best node to insert e
	node := tree.walkDown(e, tree.root)

	// node.children here are all leaf node. check tree.walkDown
	leastExpansion := findMBRNeedingLeastExpansion(e, node.children)
	leastExpansion.parent = node
	return leastExpansion
}

// walkDown recursively traverses the tree downwards starting from currentPosition,
// in search of the leaf node parent that can accommodate entry toFit.
// Note that currentPosition must be a node on this tree.
// The node returned to caller always satisfies node.isParentToLeafNode
func (tree *RTree) walkDown(toFit entry, currentPosition *node) *node {
	if currentPosition.isParentToLeafNode() {
		return currentPosition
	}

	overlaps := make([]*node, 0)
	for _, child := range currentPosition.children {
		if toFit.mbr.canFitWithin(child.mbr) {
			return child
		}

		if child.mbr.overlapsWith(toFit.mbr) {
			overlaps = append(overlaps, child)
			continue
		}
	}

	// if no children.mbr of currentPosition overlaps with toFit,
	// return child needing the least expansion
	if len(overlaps) == 0 {
		return findMBRNeedingLeastExpansion(toFit, currentPosition.children)
	}

	leastExpansion := findMBRNeedingLeastExpansion(toFit, overlaps)
	return tree.walkDown(toFit, leastExpansion)
}

// satisfiesHeightConstraints checks that tree height satisfies the height constraint
// i.e., height of an R-tree containing N index records is at most |logm N| -1 (RTree property::)
func (tree *RTree) satisfiesHeightConstraint() bool {
	constraintFactor := math.Log(float64(tree.indexCount)) / math.Log(minEntriesPerLeaf)
	constraintFactor = math.Abs(constraintFactor) - 1
	return float64(tree.height) <= constraintFactor
}

// satisfiesMaximumNodesConstraint checks that this tree
// satisfies the maximum number of nodes in a tree constraint
// i.e., maximum number of nodes must satisfy ceil(N/m) + ceil(N/m^2) + 1,
// where N = indexCount or len(jobs), m = minEntriesPerLeaf (RTree property::)
func (tree *RTree) satisfiesMaximumNodesConstraint() bool {
	squareMin := minEntriesPerLeaf * minEntriesPerLeaf
	constraintFactor := math.Ceil(float64(tree.indexCount/minEntriesPerLeaf)) +
		math.Ceil(float64(tree.indexCount/squareMin)) + 1
	return float64(tree.totalNodes) <= constraintFactor
}

func search(within models.Distance, center models.Location, d map[string][]models.Job) []models.Job {
	jobs := make([]models.Job, 0)
	c1 := haversine.Coord{
		Lat: center.Latitude,
		Lon: center.Longitude,
	}
	for _, v := range d {
		for _, j := range v {
			c2 := haversine.Coord{
				Lat: j.Location.Latitude,
				Lon: j.Location.Longitude,
			}
			if _, km := haversine.Distance(c1, c2); km <= within.Value {
				jobs = append(jobs, j)
			}
		}
	}
	return jobs
}
