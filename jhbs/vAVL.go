package jhbs

import (
	"fmt"
	"sync"
)

// AVL tree adapted from Julienne Walker's presentation at
// http://eternallyconfuzzled.com/tuts/datastructures/jsw_tut_avl.aspx.
// with some modifications and additions.

// vAVLNode is a node in an AVL tree.
type vAVLNode struct {
	venue   *Venue       // venues will be sorted inside vAVL based on their ratings
	Balance int          // balance factor
	Link    [2]*vAVLNode // children, indexed by "direction", 0 or 1.
}

// TODO: implement mu.RWLock and mu.RWUnlock
type vAVL struct {
	mu   sync.RWMutex
	root *vAVLNode
	size int // size of vAVL
}

// A little readability function for returning the opposite of a direction,
// where a direction is 0 or 1.  Go inlines this.
// Where JW writes !dir, this code has opp(dir).
func opp(dir int) int {
	return 1 - dir
}

// single rotation
func (v *vAVL) single(root *vAVLNode, dir int) *vAVLNode {
	save := root.Link[opp(dir)]
	root.Link[opp(dir)] = save.Link[dir]
	save.Link[dir] = root
	return save
}

// double rotation
func (v *vAVL) double(root *vAVLNode, dir int) *vAVLNode {
	save := root.Link[opp(dir)].Link[dir]

	root.Link[opp(dir)].Link[dir] = save.Link[opp(dir)]
	save.Link[opp(dir)] = root.Link[opp(dir)]
	root.Link[opp(dir)] = save

	save = root.Link[opp(dir)]
	root.Link[opp(dir)] = save.Link[dir]
	save.Link[dir] = root
	return save
}

// adjust valance factors after double rotation
func (v *vAVL) adjustBalance(root *vAVLNode, dir, bal int) {
	n := root.Link[dir]
	nn := n.Link[opp(dir)]
	switch nn.Balance {
	case 0:
		root.Balance = 0
		n.Balance = 0
	case bal:
		root.Balance = -bal
		n.Balance = 0
	default:
		root.Balance = 0
		n.Balance = bal
	}
	nn.Balance = 0
}

func (v *vAVL) insertBalance(root *vAVLNode, dir int) *vAVLNode {
	n := root.Link[dir]
	bal := 2*dir - 1
	if n.Balance == bal {
		root.Balance = 0
		n.Balance = 0
		return v.single(root, opp(dir))
	}
	v.adjustBalance(root, dir, bal)
	return v.double(root, opp(dir))
}

func (v *vAVL) insertR(root *vAVLNode, tgt *Venue) (*vAVLNode, bool) {
	if root == nil {
		return &vAVLNode{venue: tgt}, false
	}
	dir := 0
	// DON'T USE RATINGS TO COMPARE! HIGH CHANCE THAT THE RATINGS ARE THE SAME
	// AND THE AVL WILL REMOVE OR INSERT THE WRONG NODE / WAY!
	// if root.venue.rating < tgt.rating {
	if root.venue.name < tgt.name {
		dir = 1
	}
	var done bool
	root.Link[dir], done = v.insertR(root.Link[dir], tgt)
	if done {
		return root, true
	}
	root.Balance += 2*dir - 1
	switch root.Balance {
	case 0:
		return root, true
	case 1, -1:
		return root, false
	}
	return v.insertBalance(root, dir), true
}

// Insert a node into the AVL tree.
// Venue is inserted even if other venue with the same name already exists.
// NB: Always whether the venue exists via vTrie, before inserting into vAVL!
func (v *vAVL) Insert(tree **vAVLNode, tgt *Venue) {
	v.mu.Lock()
	*tree, _ = v.insertR(*tree, tgt)
	v.size++
	v.mu.Unlock()
}

func (v *vAVL) removeBalance(root *vAVLNode, dir int) (*vAVLNode, bool) {
	n := root.Link[opp(dir)]
	bal := 2*dir - 1
	switch n.Balance {
	case -bal:
		root.Balance = 0
		n.Balance = 0
		return v.single(root, dir), false
	case bal:
		v.adjustBalance(root, opp(dir), -bal)
		return v.double(root, dir), false
	}
	root.Balance = -bal
	n.Balance = bal
	return v.single(root, dir), true
}

func (v *vAVL) removeR(root *vAVLNode, tgt *Venue) (*vAVLNode, bool) {
	if root == nil {
		return nil, false
	}
	// if root.venue.rating == tgt.rating {
	if root.venue.name == tgt.name {
		switch {
		case root.Link[0] == nil:
			return root.Link[1], false
		case root.Link[1] == nil:
			return root.Link[0], false
		}
		heir := root.Link[0]
		for heir.Link[1] != nil {
			heir = heir.Link[1]
		}
		root.venue = heir.venue
		tgt = heir.venue
	}
	dir := 0
	// if root.venue.rating < tgt.rating {
	if root.venue.name < tgt.name {
		dir = 1
	}
	var done bool
	root.Link[dir], done = v.removeR(root.Link[dir], tgt)
	if done {
		return root, true
	}
	root.Balance += 1 - 2*dir
	switch root.Balance {
	case 1, -1:
		return root, true
	case 0:
		return root, false
	}
	return v.removeBalance(root, dir)
}

// Remove a single Venue from an AVL tree.
// requires searching for that tree first
// WARNING: seems to delete the wrong node
// when there are 2 nodes that have the same ratings
func (v *vAVL) Remove(tree **vAVLNode, item string) {
	tgt := v.findNode(*tree, item)
	if tgt != nil {
		v.mu.Lock()
		*tree, _ = v.removeR(*tree, tgt.venue)
		v.size--
		v.mu.Unlock()
	} else {
		fmt.Printf("%s not found in vAVL\n", item)
	}
}

// Find searches for whether that venue exists
// returns nil if not found
// useful for Remove
func (v *vAVL) Find(item string) *vAVLNode {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.findNode(v.root, item)
}

func (v *vAVL) findNode(root *vAVLNode, item string) *vAVLNode {
	if root == nil {
		return nil
	} else {
		// compare venue name strings
		if root.venue.name == item {
			return root
		} else {
			node := v.findNode(root.Link[0], item)
			if node != nil {
				return node
			} else {
				return v.findNode(root.Link[1], item)
			}
		}
	}
}

// Traverse AVL with in-order traversal: Venues go to channel first based on A-Z
func (v *vAVL) Traverser(root *vAVLNode, ch chan<- *Venue) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	v.traverse(root, ch)
	close(ch)
}

// Traverse AVL with in-order traversal: Venues go to channel first based on A-Z
func (v *vAVL) traverse(root *vAVLNode, ch chan<- *Venue) {
	if root != nil {
		// to left of node
		v.traverse(root.Link[0], ch)
		// at the root, send pointer of venue to ch
		ch <- root.venue
		// to right of node
		v.traverse(root.Link[1], ch)
	}

}
