package rtree

import (
	"errors"
	"github.com/ercross/grabjobs/internal/models"
	"math"
	"math/rand"
)

const maxEntriesPerNode = 30

// minEntriesPerNode is the minimum branching factor
// or the minimum number of children of each node
const minEntriesPerNode = maxEntriesPerNode / 2
const dimension = 2

// RTree to efficiently search spatial data.
// Ref: http://www-db.deis.unibo.it/courses/SI-LS/papers/Gut84.pdf
// Use any of NewWithEntry or NewWithEntries to obtain a new tree
type RTree struct {
	height     int
	indexCount int
	totalNodes int

	// root node may be a leaf if its the only node on the tree.
	// RTree property:: If root is not a leaf, then it must have at least 2 children
	root *node

	raw []models.Job
}

// NewWithEntry initializes a new node with an entry
func NewWithEntry(entry models.Job) *RTree {
	leaf := &node{
		mbr:      newMBRAround(entry.Location),
		children: nil,
		entries:  []*models.Job{&entry},
	}

	tree := RTree{
		height:     1,
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
	tree := NewWithEntry(entries[0])
	tree.raw = entries
	for _, entry := range entries[1:] {
		tree.Insert(entry)
	}
	return tree, nil
}

// Insert a new job into the tree.
// New index records are added at the leaves and nodes that overflow(i.e., len(node.children)>M) are split.
func (tree *RTree) Insert(entry models.Job) {
	leaf := tree.chooseLeaf(entry.Location)
	if leaf.hasEntrySpace() {
		leaf.insertNewEntry(entry)
	} else {
		// todo
	}
}

// FindJobs finds job within the distance of location specified by @of.
func (tree *RTree) FindJobs(within models.Distance, of models.Location) []models.Job {
	// todo implementation of tree is currently incomplete
	r := rand.Intn(len(tree.raw))
	return tree.raw[(r - 10):(r - 1)]
}

func (tree *RTree) chooseLeaf(location models.Location) *node {

	// if root is leaf, choose root
	if tree.root.isLeaf() {
		return tree.root
	}

	// else choose the root node of a subtree inside tree.root
	// whose mbr needs the least expansion to accommodate location
	// todo
	return nil
}

// satisfiesHeightConstraints checks that tree height satisfies the height constraint
// i.e., height of an R-tree containing N index records is at most |logm N| -1 (RTree property::)
func (tree *RTree) satisfiesHeightConstraint() bool {
	constraintFactor := math.Log(float64(tree.indexCount)) / math.Log(minEntriesPerNode)
	constraintFactor = math.Abs(constraintFactor) - 1
	return float64(tree.height) <= constraintFactor
}

// satisfiesMaximumNodesConstraint checks that this tree
// satisfies the maximum number of nodes in a tree constraint
// i.e., maximum number of nodes must satisfy ceil(N/m) + ceil(N/m^2) + 1,
// where N = indexCount or len(jobs), m = minEntriesPerNode (RTree property::)
func (tree *RTree) satisfiesMaximumNodesConstraint() bool {
	squareMin := minEntriesPerNode * minEntriesPerNode
	constraintFactor := math.Ceil(float64(tree.indexCount/minEntriesPerNode)) +
		math.Ceil(float64(tree.indexCount/squareMin)) + 1
	return float64(tree.totalNodes) <= constraintFactor
}
