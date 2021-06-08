// merge sort
// adapted from Rosetta's code
// which involves optimisation of using only
// length of a sub-array during "merge" process

package jhbs

import (
	"sync"
)

// until generics come out, I have to do this for different structs
// see mergeSortVenues.go for base definitions for merge sort

// for implementing merging step in mergesort
func (vss *VenueSortSlice) merge(a VenueSortSlice, mid int, c func(a *Venue, b *Venue) bool) {

	var s = (make(VenueSortSlice, len(a)/2+1)) // temp slice for merge step
	copy(s, a[:mid])
	l, r := 0, mid
	for i := 0; ; i++ {

		// if s[l].age <= a[r].age { // use age
		// if s[l].lastName <= a[r].lastName { // use last name
		// if (*s[l]).target.venue > (*a[r]).target.venue {
		compared := c((s[l]), (a[r]))
		if compared {
			a[i] = s[l]
			l++
			if l == mid {
				break
			}
		} else {
			a[i] = a[r]
			r++
			if r == len(a) {
				copy(a[i+1:], s[l:mid])
				break
			}
		}

	}
}

// sequential version of mergesort
// used when size of SliceOfPersons is small
// otherwise creating co-routines for small slices slows down mergesort a lot
func (vss *VenueSortSlice) mergeSort(a VenueSortSlice, c func(a *Venue, b *Venue) bool) {
	if len(a) > 1 {
		mid := len(a) / 2
		vss.mergeSort(a[:mid], c)
		vss.mergeSort(a[mid:], c)
		vss.merge(a, mid, c)
	}
}

// this Mergesort makes use of concurrency
func (vss *VenueSortSlice) ParallelMergeSort(a VenueSortSlice, c func(a *Venue, b *Venue) bool) {

	if len(a) < 2 {
		return
	}
	if len(a) <= maxForSequentialMergeSort {
		vss.mergeSort(a, c) // Sequential
	} else { // Concurrent
		mid := len(a) / 2

		// concurrency; implement each recursive merge sort in a go-routine
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			vss.ParallelMergeSort(a[:mid], c)
		}()

		go func() {
			defer wg.Done()
			vss.ParallelMergeSort(a[mid:], c)
		}()

		wg.Wait()
		// if a[mid-1].age <= a[mid].age { // use age
		// if a[mid-1].lastName <= a[mid].lastName { // use last name
		// if (*a[mid-1]).target.venue > (*a[mid]).target.venue {
		compared := c((a[mid-1]), (a[mid]))
		if compared {
			return
		}
		vss.merge(a, mid, c)
	}
}

// reverses a SliceOfPersons; used for changing order from ASC to DESC
func (vss *VenueSortSlice) Reverse() {
	// swap start and end, in converging indices to the mid
	for i := 0; i < len(*vss)/2; i++ {
		(*vss)[i], (*vss)[len(*vss)-1-i] = (*vss)[len(*vss)-1-i], (*vss)[i]
	}
}
