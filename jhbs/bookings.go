package jhbs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type BookingStatus int

// booking statuses
const (
	cancelled BookingStatus = iota // cannot update a cancelled booking. Immediately updates venue availability by undoing old booking
	rejected
	pending // if booking is changed, "approved" becomes "pending"
	approved
)

func (t BookingStatus) String() string {
	return [...]string{"Cancelled", "Rejected", "Pending", "Approved"}[t]
}

// Booking data
type BookingTarget struct {
	// when user makes 1st booking under the bookingID
	venue     string // verify whether the venue exists (search it in a venue Trie), before connecting to that venue's Venue struct
	startTime time.Time
	endTime   time.Time
}

// Booking is created when a user books a venue
type Booking struct {
	bookingID string
	memberID  int // members are represented by memberID
	status    BookingStatus

	// originally part of booking slice, later delegated to being pointed at as a struct
	// in order to accomodate rebooking (otherwise size of 1 Booking will balloon)
	target *BookingTarget
	// if user changes the booking, deep copy target to oldTarget,
	// then update target
	oldTarget *BookingTarget // nil if no rebooking
}

func (b *Booking) String() string {
	// in case I forgot to decrement memberID
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Oi! Did you forget to decrement memberID? (Check variable mID)")
		}
	}()

	var s string
	var t, u time.Time                 // t for start, u for end
	mID := b.memberID                  // get memberID
	m := JHBase.members[mID-MIDOffset] // retrieve member info
	t = b.target.startTime             // booking start time
	u = b.target.endTime               // booking end time
	// like (v *Venue) String(), but without venue writeup
	s += fmt.Sprintf("%s ", b.status)
	s += fmt.Sprintf("booking %s ", b.bookingID)

	s += fmt.Sprintf("for %s, ", b.target.venue)
	s += fmt.Sprintf(t.Format("02-01-2006 (Mon) 15:04"))
	s += " - "
	s += fmt.Sprintf(u.Format("02-01-2006 (Mon) 15:04"))
	s += " "
	// if the member is not deleted
	if m != nil {
		s += fmt.Sprintf("by %s ", m.firstName)
		s += fmt.Sprintf("%s ", m.lastName)
		s += fmt.Sprintf("(%d, ", mID)
		s += fmt.Sprintf("%s)\n", m.tier)
	} else {
		// if member is deleted
		s += fmt.Sprintf("by deleted Member %d\n", mID)
	}

	return s
}

// internal struct
type bookings []*Booking

// slice of pointers to bookings,
// e.g. admin asks to sort Bookings
type Bookings struct {
	sync.RWMutex
	bs bookings // bookings == []*Booking
}

func MakeBookings() *Bookings {
	return &Bookings{
		bs: make([]*Booking, 0, 16),
	}
}

// appends bookings(s) to bookings slice
// concurrency-safe thanks to RWMutex
// NB: For approved bookings of a venue,
// use v.AddApprovedBooking(b *Booking) instead
func (bs *Bookings) Append(b ...*Booking) {
	bs.Lock()
	bs.bs = append(bs.bs, b...)
	bs.Unlock()
}

// TODO: Sort booking

// delete a booking from bookings slice based on bookingID
// true if found and then deleted, false if not found
func (bs *Bookings) DeleteByBookingID(bID string) bool {
	bs.Lock()
	defer bs.Unlock()
	// sequential search
	for i, b := range bs.bs {
		// found the booking to delete
		if b.bookingID == bID {
			// delete booking
			copy(bs.bs[i:], bs.bs[(i+1):])
			bs.bs = bs.bs[:len(bs.bs)-1]
			return true
		}
	}
	// could not find booking to delete
	return false
}

// delete a booking from bookings slice based on start time of that booking
// true if found and then deleted, false if not found
func (bs *Bookings) DeleteByStartTime(start time.Time) bool {
	bs.Lock()
	defer bs.Unlock()
	// uses binary search to find booking with that start time
	// returns index of booking to be deleted,
	// and whether search index can be found
	i, isFound := bs.bs.binarySearchByStartTime(start, 0, len(bs.bs))
	if isFound {
		// delete booking if found
		copy(bs.bs[i:], bs.bs[(i+1):])
		bs.bs = bs.bs[:len(bs.bs)-1]
	}
	// could not find booking to delete
	return isFound
}

// internal fn: binary search using recursion
// for finding the booking to delete using startTime
// returns index and whether the search result is found
func (bs *bookings) binarySearchByStartTime(start time.Time, left int, right int) (int, bool) {
	mid := (left + right) / 2
	if (*bs)[mid].target.startTime == start {
		return mid, true
	}
	// if left and right are same indices, return not found (-1, false)
	if left == right {
		return -1, false
	}
	// look left
	if (*bs)[mid].target.startTime.After(start) {
		return bs.binarySearchByStartTime(start, left, mid-1)
	}
	// else look right
	return bs.binarySearchByStartTime(start, mid+1, right)

}

func (bs *Bookings) String() string {
	// in case I forgot to decrement memberID
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Oi! Did you forget to decrement memberID?")
		}
	}()
	var s string
	for _, b := range bs.bs {
		s += fmt.Sprint(b)
	}
	return s
}

// internal function to validate booking fields
func validateBookingFields(memberID int, venue string,
	startTime time.Time, endTime time.Time) error {
	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	// memberID starts from 100000 (MIDOffset)
	idCheck := memberID - MIDOffset

	if (idCheck >= len(JHBase.members)) || (idCheck < 0) {
		s := fmt.Sprintf("- MemberID %d is not a valid memberID", memberID)
		return errors.New(s)
	}

	// check if venue exists
	_, gotVenue := vTrie.Find(venue)
	if !gotVenue {
		s := fmt.Sprintf("- %s is not a venue", venue)
		return errors.New(s)
	}

	// check if start time is before end time
	if !(startTime.Before(endTime)) {
		s := "Start time is not before end time."
		return errors.New(s)
	}

	return nil
}

// when user is sure that it wants to book, create a Booking object
func NewBooking(memberID int, venue string,
	startTime time.Time, endTime time.Time, status int,
	wg *sync.WaitGroup) (*Booking, error) {

	// defer wg.Done()

	// returns if invalid memberID, wrong venue or statTime later than endTime
	err := validateBookingFields(memberID, venue,
		startTime, endTime)

	if err != nil {
		return nil, err
	}

	// generate 6-letter booking ID
	bookingID := generateID()

	bookingTarget := &BookingTarget{
		venue:     venue,
		startTime: startTime,
		endTime:   endTime,
	}

	// store bookingID in bTrie, a search BTrie for bookings
	// regenerate bookingID if got duplicate

	_, alrGotBookingID := bTrie.BFind(bookingID)
	for alrGotBookingID {
		bookingID = generateID()
		_, alrGotBookingID = bTrie.BFind(bookingID)

	}

	// then put bookingID inside a new Booking object
	// along with other details
	booking := &Booking{
		bookingID: bookingID,
		memberID:  memberID,
		status:    BookingStatus(status),
		target:    bookingTarget,
		oldTarget: nil,
	}

	// put booking into a Trie of bookings (bTrie)
	err = bTrie.BPut(booking, wg)

	// put booking in member's booking history
	JHBase.members[memberID-MIDOffset].bookings.bs = append(JHBase.members[memberID-MIDOffset].bookings.bs, booking)

	// slot booking into venue's Priority Queue
	venuesAVL.Find(venue).venue.waitlist.Enqueue(booking)

	return booking, nil

}

// when user is sure that it wants to book, create a Booking object
func EditBooking(bookingID string /*memberID int,*/, venue string,
	startTime time.Time, endTime time.Time, status int,
	wg *sync.WaitGroup) (*Booking, error) {

	// EditBooking() seems prone to errors
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	newBookingTarget := &BookingTarget{
		venue:     venue,
		startTime: startTime,
		endTime:   endTime,
	}

	// store bookingID in bTrie, a search BTrie for bookings
	// regenerate bookingID if got duplicate

	bNode, alrGotBookingID := bTrie.BFind(bookingID)

	// unlikely cannot find,
	// since for users, PrepareUserEditBooking() would have
	// check the existence of that bookingID,
	// but still I put it here just in case
	if !alrGotBookingID {
		return nil, errors.New("Can't find booking ID")
	}

	// amend booking directly inside
	// get booking from this booking node
	b := bNode.booking

	// then put bookingID inside a new Booking object
	// along with other details
	// no need change b.bookingID and b.memberID
	b.oldTarget = b.target
	b.target = newBookingTarget

	// if no change in choice of venue,
	// and user's booking status is still pending (2)
	// amend booking directly
	// no need to look through waitlist or approved list
	if venue == b.oldTarget.venue && b.status == 2 {
		return b, nil
	}

	// otherwise, make booking status 2 for pending
	// then search through waitlist and approved list
	// this applies even if venue has not changed
	// e.g. if booking got approved, but user wants to amend
	b.status = BookingStatus(2)
	// If venue choice has changed,
	// remove booking from old venue's waitlist or approved bookings
	// then put booking under new venue's waitlist
	prev := b.oldTarget.venue

	// Remove booking from old venue's waiting list and approved list
	bl1 := venuesAVL.Find(prev).venue.waitlist
	bl2 := venuesAVL.Find(prev).venue.approvedBookings

	wg.Add(2)
	go func() {
		defer wg.Done()
		defer fmt.Println("Removing from waitlist")
		bl1.Remove(bookingID)
	}()
	go func() {
		defer wg.Done()
		defer fmt.Println("Removing from approved list")
		bl2.DeleteByBookingID(bookingID)
	}()
	wg.Wait()

	// slot booking into venue's Priority Queue
	venuesAVL.Find(venue).venue.waitlist.Enqueue(b)
	// notify member that booking got edited
	NotifyMember(b.memberID)
	return b, nil

}

// func RejectBooking()
func RejectBooking(bID string, wg *sync.WaitGroup) error {

	// if admin rejects a rejected booking,
	// this function will definitely panic.
	// in that case, send out this message
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Unable to reject %s, maybe because it is already rejected?\n", bID)
		}
	}()

	// find whether booking with that bookingID exists
	bNode, gotBooking := bTrie.BFind(bID)
	if !gotBooking {
		s := fmt.Sprintf("Cannot reject booking %s; no such booking", bID)
		return errors.New(s)
	}

	// do not remove a cancelled booking
	if bNode.booking.status == cancelled {
		s := fmt.Sprintf("%s already cancelled its booking", bID)
		return errors.New(s)
	}

	// Get venue name
	v := bNode.booking.target.venue

	// find booking in venue's waiting list only
	// booking cannot be in approved list, so no need to search there
	bl1 := venuesAVL.Find(v).venue.waitlist

	// remove booking from waitlist
	rejectedBooking, err := bl1.Remove(bID)

	if err != nil {
		return err
	}

	// update booking status
	rejectedBooking.status = rejected
	fmt.Println(rejectedBooking)

	// inform user
	mID := rejectedBooking.memberID
	NotifyMember(mID)
	return nil
}

// CancelBooking removes a booking from the venue based on bookingID,
// pending further deletion by admin
func CancelBooking(bID string, wg *sync.WaitGroup) error {

	// if admin tries to remove a removed booking,
	// sometimes this function will panic.
	// in that case, send out this message
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Unable to remove %s, maybe because it is already removed?\n", bID)
		}
	}()

	// find whether booking with that bookingID exists
	bNode, gotBooking := bTrie.BFind(bID)
	if !gotBooking {
		s := fmt.Sprintf("Cannot delete booking %s; no such booking", bID)
		return errors.New(s)
	}

	// do not remove a cancelled booking
	if bNode.booking.status == cancelled {
		s := fmt.Sprintf("%s already cancelled its booking", bID)
		return errors.New(s)
	}

	// Get venue name
	v := bNode.booking.target.venue

	// find booking in venue's waiting list and approved list
	bl1 := venuesAVL.Find(v).venue.waitlist
	bl2 := venuesAVL.Find(v).venue.approvedBookings

	// then spawn 2 go-routines that remove
	// the pointer to that booking from both lists
	wg.Add(2)
	go func() {
		defer wg.Done()
		//_, err1 = bl1.Remove(bID)
		bl1.Remove(bID)
	}()
	go func() {
		defer wg.Done()
		bl2.DeleteByBookingID(bID)
	}()
	wg.Wait()
	// if err1 != nil && !found {
	// 	s := fmt.Sprintln(err1, "or not found in approved list")
	// 	return errors.New(s)
	// }
	// bl2.BDelete(bID)

	// update booking status
	bNode.booking.status = cancelled
	fmt.Println(bNode.booking)

	// inform user
	mID := bNode.booking.memberID
	NotifyMember(mID)
	return nil
}

// Process bookings of one venue
func (v *Venue) ProcessBookings() error {

	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	// when the priority queue has bookings (size != 0)
	// dequeues a booking from the priority queue
	if v.waitlist == nil {
		s := fmt.Sprintf("No waitlist for %s has been created yet", v.name)
		return errors.New(s)
	}

	var b *Booking
	var err error

OUTER:
	// dequeues booking from the waitlist
	// only put it into approved booking BTrie for approvals
	// otherwise, dequeued bookings are only placed back into waitlist
	// when user amends the booking
	for err == nil {

		// dequeue
		b, err = v.waitlist.Dequeue()

		if err != nil {
			break
		}

		// check if isActive member
		if JHBase.members[b.memberID-MIDOffset].isActive == false {
			b.status = 1 // 1 means rejected
			pfa("Member %d who booked %v has quit; ", b.memberID, v.name)
			pfa("booking %s rejected\n", b.bookingID)
			continue
		}

		tgt := b.target
		start := tgt.startTime
		end := tgt.endTime

		for _, t := range v.approvedBookings.bs {
			bookedStart := t.target.startTime
			bookedEnd := t.target.endTime
			// REJECT if startTime of booking is within a booked slot
			if start.After(bookedStart) && start.Before(bookedEnd) {
				b.status = 1 // 1 means rejected
				pfa("Can't book %s %d-%02d-%02d %02d:00 - %d-%02d-%02d %02d:00; overlaps with a booked slot; ",
					v.name, start.Year(), start.Month(), start.Day(), start.Hour(),
					end.Year(), end.Month(), end.Day(), end.Hour())
				pfa("booking %s rejected\n", b.bookingID)
				// continue perBookingDequeued
				continue OUTER
			}
			// REJECT if endtime of booking is within a booked slot
			if end.After(bookedStart) && end.Before(bookedEnd) {
				b.status = 1 // 1 means rejected
				pfa("Can't book %s %d-%02d-%02d %02d:00 - %d-%02d-%02d %02d:00; overlaps with a booked slot; ",
					v.name, start.Year(), start.Month(), start.Day(), start.Hour(),
					end.Year(), end.Month(), end.Day(), end.Hour())
				pfa("booking %s rejected\n", b.bookingID)
				// continue perBookingDequeued
				continue OUTER
			}
			// REJECT if this booking goes over the booked slot
			// e.g. your booking: 1:00 PM - 4:00 PM
			// BUT booked slot: 2:00 PM - 3:00 PM
			if start.Before(bookedStart) && end.After(bookedEnd) {
				b.status = 1 // 1 means rejected
				pfa("Can't book %s %d-%02d-%02d %02d:00 - %d-%02d-%02d %02d:00; overlaps with a booked slot; ",
					v.name, start.Year(), start.Month(), start.Day(), start.Hour(),
					end.Year(), end.Month(), end.Day(), end.Hour())
				pfa("booking %s rejected\n", b.bookingID)
				// continue perBookingDequeued
				continue OUTER
			}
		}
		// ACCEPT after checking through all the timesBooked
		b.status = 3 // 3 for approved
		v.AddApprovedBooking(b)

		pfa("Successful booking for %s from %d-%02d-%02d %02d:00 to %d-%02d-%02d %02d:00; ",
			v.name, start.Year(), start.Month(), start.Day(), start.Hour(),
			end.Year(), end.Month(), end.Day(), end.Hour())
		pfa("booking %s approved\n", b.bookingID)

		// continue perBookingDequeued
		continue OUTER
	}

	pfa("End of processing bookings for venue %s\n", v.name)
	return nil
}

// AutoProcessBookings does v.ProcessBookings for every venue
func AutoProcessBookings() {

	// spawns a go=routine and a channel
	tempVSlice := MakeVenueSortSlice()
	for _, v := range *tempVSlice {
		v.ProcessBookings()
	}
}

// MakeBookingSortSlice prepares a booking sort slice for admin to sort
func MakeBookingSortSlice() *Bookings {
	// happens on its own go channel
	var bch = make(chan *Booking)
	go bTrie.BTrieTraverser(bTrie.root, bch)
	var bs *Bookings
	for b := range bch {
		bs.bs = append(bs.bs, b)
	}
	return bs
}

// BookingID returns bookingID
func (b *Booking) BookingID() string {
	return b.bookingID
}

// MemberID returns MemberID under which the booking was made
func (b *Booking) MemberID() int {
	return b.memberID
}

// BookingID returns bookingID
func (b *Booking) Status() BookingStatus {
	return b.status
}

// Venue returns venue (As a string, not as Venue obj)
func (b *Booking) Venue() string {
	return b.target.venue
}

// StartTime returns the start time
func (b *Booking) StartTime() time.Time {
	return b.target.startTime
}

// EndTime returns the end time
func (b *Booking) EndTime() time.Time {
	return b.target.endTime
}

// Booking returns the whole booking
func (b *Booking) Booking() *Booking {
	return b
}

// Bookigns returns the slice of bookings
func (bs *Bookings) Bookings() bookings {
	return bs.bs
}
