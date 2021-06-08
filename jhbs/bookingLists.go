// Priority List resembles a priority queue (FIFO)
// but in addition, a booking inside the list can be removed
// e.g. because user or admin cancels that booking

package jhbs

import (
	"errors"
	"fmt"
	"sync"
)

type Node struct {
	priority int
	// item     string
	booking *Booking
	next    *Node
}

type PriorityList struct {
	front *Node
	back  *Node
	size  int
	mu    sync.RWMutex
}

// func (p *PriorityList) Enqueue(item string, prty int) error {
func (p *PriorityList) Enqueue(booking *Booking) error {

	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	// retrieve member
	m := JHBase.members[booking.memberID-MIDOffset]
	prty := int(m.tier) // see members.go for Member declaration

	newNode := &Node{
		priority: prty,
		booking:  booking,
		next:     nil,
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.front == nil {
		p.front = newNode
	} else {
		// check priority of front person
		if prty > p.front.priority {
			// Insert new Node before front
			newNode.next = p.front
			p.front = newNode
		} else {
			currentNode := p.front
			// if incoming user has same or lower priority, traverse further
			for currentNode.next != nil && prty <= currentNode.next.priority {
				currentNode = currentNode.next
			}
			// Either at the end of the queue
			// or at required position
			newNode.next = currentNode.next
			currentNode.next = newNode
		}
	}
	p.size++
	return nil
}

// Dequeue() pops the first node out of the queue
// reports error if empty or unsuccessful
func (p *PriorityList) Dequeue() (*Booking, error) {

	p.mu.Lock()
	defer p.mu.Unlock()

	var booking *Booking
	if p.front == nil {
		return nil, errors.New("empty queue!")
	}
	booking = p.front.booking
	if p.size == 1 {
		p.front = nil
		p.back = nil
	} else {
		p.front = p.front.next
	}
	p.size--
	return booking, nil
}

// Remove removes the booking from priority list
// but does not delete it
// usually it is thronw into a bTrie for toBeDeleted bookings
func (p *PriorityList) Remove(bID string) (*Booking, error) {

	// in case panics occur
	// but still don't want to disrupt others
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	currentNode := p.front

	// if priority list has no booking
	if currentNode == nil {
		s := fmt.Sprint("Priority list is empty.")
		return nil, errors.New(s)
	}

	// if first booking is the one to be deleted
	// treat it as de-queueing
	if currentNode.booking.bookingID == bID {
		// p.size-- inside p.Dequeue()
		return p.Dequeue() // also has p.mu.Lock()
	}

	// mutex locks after p.Dequeue(), so as to prevent locking a locked mutex
	p.mu.Lock()
	defer p.mu.Unlock()

	prevNode := currentNode
	currentNode = currentNode.next

	// if booking can be found in middle of prioritylist
	for currentNode != nil {
		tgt := currentNode.booking
		if tgt.bookingID == bID {
			fmt.Printf("%s removed from waiting list of %s.\n", bID, tgt.target.venue)
			prevNode.next = currentNode.next
			p.size--
			return tgt, nil
		}
		prevNode = currentNode
		currentNode = currentNode.next
	}

	// if found at the end
	tgt := currentNode.booking
	if tgt.bookingID == bID {
		currentNode = nil
		p.size--
		return tgt, nil
	}

	// if at the end still can't find it
	s := fmt.Sprintf("Booking %s not found", bID)
	return nil, errors.New(s)

}

// DumpNodes traverse the priority queue and prints to console
func (p *PriorityList) DumpNodes() error {

	p.mu.RLock()
	defer p.mu.RUnlock()

	currentNode := p.front
	if currentNode == nil {
		fmt.Println("Waitlist is empty.")
		return nil
	}
	fmt.Printf("%+v", currentNode.booking)

	for currentNode.next != nil {
		currentNode = currentNode.next
		fmt.Printf("%+v", currentNode.booking)
	}
	return nil
}

// DumpNodes2 traverse the priority queue and returns bookings
// if nothing, just close the channel
func (p *PriorityList) DumpNodes2(ch chan<- *Booking) {

	p.mu.RLock()
	defer p.mu.RUnlock()

	currentNode := p.front
	if currentNode == nil {
		fmt.Println("Waitlist is empty.")
		close(ch)
		return
	}
	ch <- currentNode.booking

	for currentNode.next != nil {
		currentNode = currentNode.next
		ch <- currentNode.booking
	}
	close(ch)
}

func (p *PriorityList) IsEmpty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.size == 0
}
