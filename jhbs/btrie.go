package jhbs

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// BTrie is used for bookings
// Unlike in a typical Trie, at the end of the bookingID,
// is a pointer to the respective Booking object

// Unlike vTrie using Trie,
// Autocomplete should not be included for booking IDs!

// This BTrie struct is adapted from Sebastian Ojeda at https://www.fullstackgo.io/prefix-trees-in-go
// with additions incl. BTrie traversal (root-child)

type BTrie struct {
	root      *BTrieNode
	wordCount int
	mu        sync.RWMutex
}

type BTrieNode struct {
	prefix   rune
	parent   *BTrieNode
	children map[rune]*BTrieNode
	isWord   bool
	results  int // number of words this node contains
	// for venues, add in a venue struct?
	booking *Booking
}

// MakeBTrie() returns an initialized BTrie with a root node.
func MakeBTrie() *BTrie {
	return &BTrie{
		root: &BTrieNode{
			prefix:   0,
			parent:   nil,
			children: make(map[rune]*BTrieNode),
			isWord:   false,
			results:  0,
			booking:  nil,
		},
		wordCount: 0,
	}
}

// cleanUpTheWord() identical to that used in Tries

// BPut will add a new word to our BTrie, adding
// new nodes as needed.
// unlike Trie.Put, BPut asks for the Booking to be added, not just the word
func (t *BTrie) BPut(booking *Booking, wg *sync.WaitGroup) error {

	wg.Add(1)
	defer wg.Done()
	node := t.root

	// clean up the word before putting it into this BTrie
	word := cleanUpTheWord(booking.bookingID)

	if _, isWord := t.BFind(word); isWord == true {
		s := fmt.Sprintf("- %v already exists", strings.Title(word))
		return errors.New(s)
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	// adjust results of children
	// if the word is already inserted beforehand, nth happens
	for _, c := range word {

		if n, ok := node.children[c]; ok {
			node = n
			n.results++
		} else {
			node.children[c] = &BTrieNode{
				prefix:   c,
				parent:   node,
				children: make(map[rune]*BTrieNode),
				isWord:   false,
				results:  1,
				booking:  nil,
			}
			node = node.children[c]
		}
	}

	node.isWord = true
	node.booking = booking

	t.wordCount++
	return nil
}

// Find will try to return the node for the last
// character in the string if found, as well as
// a bool indicating whether or not it is a word
// in the dictionary.
func (t *BTrie) BFind(word string) (*BTrieNode, bool) {

	// if BTrie not yet created
	if t.root == nil {
		return nil, false
	}

	node := t.root

	// clean up this word before finding it in this BTrie
	word = cleanUpTheWord(word)

	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, c := range word {

		if n, ok := node.children[c]; ok {
			node = n
		} else {
			return nil, false
		}
	}
	return node, node.isWord
}

// Delete removes a word from a BTrie.
func (t *BTrie) BDelete(word string) bool {
	// clean up this word before finding it in this BTrie
	word = cleanUpTheWord(word)

	// if cannot find the word,
	// or if there is an entry but it is not a word
	// (e.g. "Tam" is not a word but found in "Tampines"),
	// delete nothing
	node, isWord := t.BFind(word)
	if !isWord {
		return false
	}

	// time to delete the word
	t.mu.Lock()
	defer t.mu.Unlock()
	node.isWord = false

	for node.parent != nil {
		node.results--
		if node.results == 0 {
			delete(node.parent.children, node.prefix)
		}
		node = node.parent
	}

	t.wordCount--
	return true
}

// BTrieTraversal finds and throws all pointer to Bookings via a channel
// usually wrapped in a go-routine by caller
// e.g. var ch = make(chan string)
// go vTrie.TrieTraverser(vTrie.root, "", ch)
// for word := range ch {
// 	fmt.Println(word)
// }
func (t *BTrie) BTrieTraverser(node *BTrieNode, ch chan<- *Booking) {
	// mutex read lock
	t.mu.RLock()
	defer t.mu.RUnlock()

	t.bTrieTraversal(node, ch)
	//close the channel to avoid panic
	close(ch)
}

// internal fn
// bTrieTraversal traverses from the root to the leaves using recursion
func (t *BTrie) bTrieTraversal(node *BTrieNode, ch chan<- *Booking) {
	if node.isWord {
		ch <- node.booking // send the booking to this channel
	}

	if len(node.children) != 0 {
		for _, n := range node.children {
			t.bTrieTraversal(n, ch)
		}
	}
}

// BCancellationsDeleter deletes all bookings marked as "cancelled"
// usually wrapped in a go-routine by caller
// e.g. go BCancellationsDeleter(b.root)
// returns a counter of number of cancellations deleted
func (t *BTrie) BCancellationsDeleter(node *BTrieNode) int {
	// mutex write lock
	t.mu.Lock()
	defer t.mu.Unlock()

	var counter int
	counter = 0 // number of cancelled bookings deleted
	counter = t.bCancellationsDelete(node, counter)
	return counter
}

// internal fn
// recursively finds cancelled bookings to delete
func (t *BTrie) bCancellationsDelete(node *BTrieNode, counter int) int {
	if node.isWord {
		if node.booking.status == 0 {
			node.isWord = false
			node.booking = nil
			counter++
		}
	}

	if len(node.children) != 0 {
		for _, n := range node.children {
			counter = t.bCancellationsDelete(n, counter)
		}
	}
	return counter
}
