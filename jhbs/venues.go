package jhbs

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// venue.go allows this program to hold venue information
var venueCategories = []string{"(No Category)", "Sports", "MICE", "Arts and Entertainment", "Holiday"}
var venueRoomTypes = []string{"(No Room Type)", "Hall", "Room", "Theatre", "Sands", "Tent",
	"Street", "Square", "Racetrack", "Stadium", "Studio", "Floating Platform"}

// get all venue categories
func GetVenueCategories() {
	for i, cat := range venueCategories {
		if i == 0 { // skip no category
			continue
		}
		// list the rest
		fmt.Print(cat)
		if i != len(venueCategories)-1 {
			fmt.Print(", ")
		}
	}
}

// check if category exists
func CheckIfVenueCategoryExists(str string) (int, bool) {
	for i, cat := range venueCategories {
		if str == cat {
			return i, true
		}
	}
	return -1, false
}

// get all venue categories
func GetRoomTypes() {
	for i, t := range venueRoomTypes {
		if i == 0 { // skip no category
			continue
		}
		// list the rest
		fmt.Print(t)
		if i != len(venueRoomTypes)-1 {
			fmt.Print(", ")
		}
	}
}

// check if category exists
func CheckIfRoomTypeExists(str string) (int, bool) {
	for i, t := range venueRoomTypes {
		if str == t {
			return i, true
		}
	}
	return -1, false
}

// Venue struct
type Venue struct {
	name             string
	region           string
	capacity         int
	categoryID       int // refers to categories slice
	roomTypeID       int // refers to roomTypes slice
	area             int
	hourlyRate       int
	rating           int
	writeUp          string
	waitlist         *PriorityList // priority list of bookings
	approvedBookings *Bookings
}

func (v *Venue) String() string {
	var s string
	s += fmt.Sprintf("%s ", v.name)
	s += fmt.Sprintf("at %s for ", v.region)
	s += fmt.Sprintf("%d ppl, ", v.capacity)
	s += fmt.Sprintf("as %s ", venueCategories[v.categoryID])
	s += fmt.Sprintf("%s, ", venueRoomTypes[v.roomTypeID])
	s += fmt.Sprintf("%d sq m, ", v.area)
	s += fmt.Sprintf("SGD%d/hr rate, ", v.hourlyRate)
	s += fmt.Sprintf("rating %d\n", v.rating)
	s += fmt.Sprintf("%s\n", v.writeUp)
	return s
}

// temporary slice, only when ppl ask to sort Venues
type VenueSortSlice []*Venue

func (vs *VenueSortSlice) String() string {
	var s string
	// for i := 0; i < len(*vs); i++ {
	for _, v := range *vs {
		// like (v *Venue) String(), but without venue writeup
		s += fmt.Sprintf("%28s ", v.name)
		s += fmt.Sprintf("at %12s for ", v.region)
		s += fmt.Sprintf("%7d ppl, ", v.capacity)
		s += fmt.Sprintf("as %22s ", venueCategories[v.categoryID])
		s += fmt.Sprintf("%17s, ", venueRoomTypes[v.roomTypeID])
		s += fmt.Sprintf("%7d sq m, ", v.area)
		s += fmt.Sprintf("SGD%5d/hr rate, ", v.hourlyRate)
		s += fmt.Sprintf("rating %d\n", v.rating)
	}
	return s
}

// NewVenue creates a new venue onto a Venues struct, then returns a Venue object
// NewVenue(name, region, capacity, categoryId, roomTypeID, area, hourlyRate, ratings, write-up about venue)
// In this prototype app, venues must be unique!
func NewVenue(n string, r string, c int, catID int,
	roomTypeID int, a int, hrRate int, rating int, writeUp string,
	wg *sync.WaitGroup) (*Venue, error) {

	// TODO: Validate input, in case employees are malicious or typo

	venue := &Venue{
		name:             n,
		region:           r,
		capacity:         c,
		categoryID:       catID,      // start to refers to categories slice
		roomTypeID:       roomTypeID, // refers to roomTypes slice
		area:             a,
		hourlyRate:       hrRate,
		rating:           rating,
		writeUp:          writeUp,
		waitlist:         &PriorityList{},
		approvedBookings: MakeBookings(),
	}

	// each venue has slots for Jun 2021 (2021-06)
	var err error

	// putting name of venue into vTrie involves finding in the trie
	// prints error if name of venue alrady inside vTrie
	err = vTrie.Put(n, wg)
	if err != nil {
		return nil, err
	}
	// insert into vAVL
	venuesAVL.Insert(&venuesAVL.root, venue)

	return venue, nil
}

// MakeBookingSortSlice prepares a booking sort slice for admin to sort
func MakeVenueSortSlice() *VenueSortSlice {
	// happens on its own go channel
	var vch = make(chan *Venue)
	go venuesAVL.Traverser(venuesAVL.root, vch)
	var vs VenueSortSlice
	for v := range vch {
		vs = append(vs, v)
	}
	return &vs
}

// QueryWaitlist prints out results of
// the waiting list of a venue
func QueryWaitlist(v string, wg *sync.WaitGroup) error {

	var err error
	// if venue not found
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Recovered from %s because \"%s\" is not a venue\n", err, v)
		}
	}()

	// get venue node in venuesAVL
	vNode := venuesAVL.Find(v)
	if vNode != nil {
		err = vNode.venue.waitlist.DumpNodes()
	} else {
		err = errors.New(fmt.Sprintf("%s not a venue\n", v))
	}
	return err
}

// add approved bookings
func (v *Venue) AddApprovedBooking(b *Booking) {
	v.approvedBookings.Lock()
	tb := v.approvedBookings
	tb.bs = append(tb.bs, b)
	// TODO: insertion sort of v.timesBooked
	v.sortTimesBooked()
	v.approvedBookings.Unlock()
}

// internal fn, whenever a new booking time is added
// RWMutex (Un)locking called by the function that calls this function
// e.g. v.AddTimesBooked()
func (v *Venue) sortTimesBooked() {
	items := v.approvedBookings.bs
	var n = len(items)
	for i := 1; i < n; i++ {
		j := i
		for j > 0 {
			if items[j-1].target.startTime.After(items[j].target.startTime) {
				items[j-1], items[j] = items[j], items[j-1]
			}
			j = j - 1
		}
	}
}

// Name() returns venue name
func (v *Venue) Name() string {
	return v.name
}

// Venue() returns venue info
// TODO: Please use template fn "br"!
func (v *Venue) Venue() string {
	return strings.Replace(v.String(), "\n", `<br/>`, -1)
}

// returns bookings ([]*Booking)
func (v *Venue) TimesBooked() bookings {
	v.approvedBookings.RLock()
	s := v.approvedBookings.bs
	v.approvedBookings.RUnlock()
	return s
}

/*
// Waitlist() returns waitlist of bookings
func (v *Venue) Waitlist() string {

	var s string
	p := v.waitlist

	// dump waiting list nodes, adapted from DumpNodes()
	p.mu.RLock()
	defer p.mu.RUnlock()

	currentNode := p.front
	if currentNode == nil {
		s = "Waitlist is empty"
		return s
	}
	s = fmt.Sprintf("%+v", currentNode.booking)

	for currentNode.next != nil {
		currentNode = currentNode.next
		s += fmt.Sprintf("%+v", currentNode.booking)
	}
	return s
}
*/

// Waitlist() returns waitlist of bookings, but as []*Booking
func (v *Venue) Waitlist() *Bookings {
	var ch = make(chan *Booking)
	var bs = MakeBookings()
	go v.waitlist.DumpNodes2(ch)
	for b := range ch {
		bs.bs = append(bs.bs, b)
	}
	return bs
}

// ApprovedBookings() returns a slice of approved bookings
func (v *Venue) ApprovedBookings() bookings {
	return v.approvedBookings.bs
}
