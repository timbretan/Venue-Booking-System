package jhbs

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Trie is used for venues (as vTrie),
// because these things are easier
// to search or autocomplete with a Trie

// This Trie struct is adapted from Sebastian Ojeda at https://www.fullstackgo.io/prefix-trees-in-go
// I modified certain things to make this trie
// more robust to randomly-cased strings (e.g. cHoA ChU KanG)
// so that they are treated the same as lowercase strings.
// Also added trie traversal (root-child)

type Trie struct {
	root      *TrieNode
	wordCount int
	mu        sync.RWMutex
}

type TrieNode struct {
	prefix   rune
	parent   *TrieNode
	children map[rune]*TrieNode
	isWord   bool
	results  int // number of words this node contains
	// for venues, add in a venue struct?
}

// MakeTrie() returns an initialized trie with a root node.
func MakeTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			prefix:   0,
			parent:   nil,
			children: make(map[rune]*TrieNode),
			isWord:   false,
			results:  0,
		},
		wordCount: 0,
	}
}

// Put will add a new word to our trie, adding
// new nodes as needed.
func (t *Trie) Put(word string, wg *sync.WaitGroup) error {

	wg.Add(1)
	defer wg.Done()
	node := t.root

	// clean up the word before putting it into this trie
	word = cleanUpTheWord(word)

	if _, isWord := t.Find(word); isWord == true {
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
			node.children[c] = &TrieNode{
				prefix:   c,
				parent:   node,
				children: make(map[rune]*TrieNode),
				isWord:   false,
				results:  1,
			}
			node = node.children[c]
		}
	}

	node.isWord = true

	t.wordCount++
	return nil
}

// Find will try to return the node for the last
// character in the string if found, as well as
// a bool indicating whether or not it is a word
// in the dictionary.
func (t *Trie) Find(word string) (*TrieNode, bool) {
	node := t.root

	// clean up this word before finding it in this trie
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

// Delete removes a word from a trie.
func (t *Trie) Delete(word string) bool {
	// clean up this word before finding it in this trie
	word = cleanUpTheWord(word)

	// if cannot find the word,
	// or if there is an entry but it is not a word
	// (e.g. "Tam" is not a word but found in "Tampines"),
	// delete nothing
	node, isWord := t.Find(word)
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

// TODO: Add pointer to Booking at TrieNode
// TODO: Prepare special traverser that fetches Booking struct
func (t *Trie) TrieTraverser(node *TrieNode, str string, ch chan<- string) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.TrieTraversal(node, str, ch)
	//close the channel to avoid panic
	close(ch)
}

// TrieTraversal traverses from the root to the leaves using recursion
// Results may not be lexicographical
// func (t *Trie) TrieTraversal(node *TrieNode, str string, words []string) []string {
func (t *Trie) TrieTraversal(node *TrieNode, str string, ch chan<- string) {
	if node.isWord {
		// fmt.Println("Found word:", strings.Title(str))
		// words = append(words, strings.Title(str))
		ch <- strings.Title(str)
	}

	if len(node.children) != 0 {
		for r, n := range node.children {
			str += string(r)
			// fmt.Println("forming str:", str)
			// words = t.TrieTraversal(n, str, words)
			t.TrieTraversal(n, str, ch)
			str = str[:len(str)-1]
			// fmt.Println("after recursion str:", str)
		}
	} else {
		str = str[:len(str)-1]
		// fmt.Println("backtrack2 str:", string(str))
	}

	// return words
}

// Autocomplete helps users find desired entries
// TODO: replace word []string with ch chan string
// func (t *Trie) AutoComplete(searchString string) StringSlice {
func (t *Trie) AutoComplete(searchString string, ch chan<- string) {

	searchString = cleanUpTheWord(searchString)
	n, _ := t.Find(searchString)

	// if there are child(ren) nodes, Autocomplete via trie traversal
	if n != nil {
		// no need to sort the entries for now
		// just send the entries to the ch
		// TrieTraverser will close the channel
		go t.TrieTraverser(n, searchString, ch)
	} else {
		// if cannot find entries, close the channel
		close(ch)
	}
}
