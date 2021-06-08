package jhbs

import (
	"bufio"
	"fmt"
	"sync"
)

var scanner *bufio.Scanner

var wg sync.WaitGroup

var venuesAVL vAVL // this venueAVL is sorted based on ratings

// var bookings Bookings

// assuming this is not leap year
var daysInAMonth [12]int

// search tries for venues and bookings,
// venues for easy search and autocomplete
// (for admin) bookings for easy search autocomplete
var vTrie *Trie  // for users and admin (future vASL?)
var bTrie *BTrie // for admin

func readError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

// PrepareJHBS loads stuff for June Holidays Booking System
func PrepareJHBS() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Print("Current error: ")
			fmt.Println(err)
		}
	}()

	// make most crucial data structures first
	vTrie = MakeTrie()
	bTrie = MakeBTrie()

	daysInAMonth = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	var loadError error

	// LOAD CSV FILES
	loadError = loadVenues("csv/Venues.csv", &wg) // load into vTrie
	readError(loadError)
	loadError = loadMembers("csv/Members.csv") // load into members.members
	readError(loadError)
	loadError = loadBookings("csv/Bookings.csv", &wg) // load into bTrie
	readError(loadError)
	wg.Wait() // just in case
	fmt.Println("All CSV files loaded.")

	// process all bookings
	AutoProcessBookings()
}
