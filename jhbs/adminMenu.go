package jhbs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func LoadAdminMainMenu() {

	var choice string
	for {
		choice = ""
		fmt.Println()
		fmt.Println("Choose an admin function:")
		fmt.Println(strings.Repeat("=", 25))
		fmt.Println("1. List all venues")
		fmt.Println("2. Add venue")
		fmt.Println("3. Update venue")
		fmt.Println("4. Delete venue")
		fmt.Println("5. Retrieve all bookings")
		fmt.Println("6. Retrieve bookings from a venue")
		fmt.Println("7. Process bookings")
		fmt.Println("8. Reject a booking")
		fmt.Println("9. Cancel a booking")
		fmt.Println("10. Delete cancelled bookings")
		fmt.Print("Your choice: ")
		scanner.Scan()
		choice = scanner.Text()
		fmt.Println()
		switch choice {
		case "1":
			PrepareBrowseVenue() // same as for user; see userMenu.go
		case "2":
			PrepareAddVenue()
		case "3":
			PrepareUpdateVenue()
		case "4":
			PrepareDeleteVenue()
		case "5":
			PrepareRetrieveAllBookings()
		case "6":
			PrepareRetrieveVenueBookings()
		case "7":
			PrepareProcessBookings()
		case "8":
			PrepareRejectBooking()
		case "9":
			PrepareCancelBooking()
		case "10":
			PrepareDeleteCancelledBookings()
		default:
			return // terminate program
		}

	}

}

func PrepareAddVenue() {
	// tell user that there's a way to return to main menu
	fmt.Println("Back to main menu? Just press enter.")

	// ask venue details

	// venue name
	fmt.Print("Name of venue: ")
	scanner.Scan()
	name := scanner.Text()
	// assume that all venues have to be Title Case
	name = strings.Title(name)
	if name == "" {
		return
	}
	gotVenue := venuesAVL.Find(name)
	// venue already exists
	if gotVenue != nil {
		fmt.Printf("- %s already exists\n", name)
		return
	}

	// region
	fmt.Print("Region: ")
	scanner.Scan()
	region := scanner.Text()
	region = strings.Title(region)
	if region == "" {
		return
	}

	// capacity
	fmt.Print("Capacity: ")
	scanner.Scan()
	capacityStr := scanner.Text()
	if capacityStr == "" {
		return
	}
	capacity, err := strconv.Atoi(capacityStr)
	if err != nil {
		fmt.Println("- Capacity NaN. Back to main menu.")
		return
	}

	// for category, list all categories,
	// then user types category name,
	// then sequential search
	fmt.Print("Available categories: ")
	GetVenueCategories()
	fmt.Println()
	fmt.Print("Category: ")
	scanner.Scan()
	category := scanner.Text()
	category = strings.Title(category)
	categoryID, gotCategory := CheckIfVenueCategoryExists(category)
	if !gotCategory {
		fmt.Printf("- %s not a valid category. Back to main menu.\n", category)
		return
	}

	// roomtypes behave like category
	fmt.Print("Available room types: ")
	GetRoomTypes()
	fmt.Println()
	fmt.Print("Room type: ")
	scanner.Scan()
	roomType := scanner.Text()
	roomType = strings.Title(roomType)
	roomTypeID, gotRoomType := CheckIfRoomTypeExists(roomType)
	if !gotRoomType {
		fmt.Printf("- %s not a valid room type. Back to main menu.\n", roomType)
		return
	}

	// area
	fmt.Print("Area (sq m): ")
	scanner.Scan()
	areaStr := scanner.Text()
	if areaStr == "" {
		return
	}
	area, err := strconv.Atoi(areaStr)
	if err != nil {
		fmt.Println("- Area NaN. Back to main menu.")
		return
	}

	// hourly rate
	fmt.Print("Hourly Rate: SGD")
	scanner.Scan()
	hourlyRateStr := scanner.Text()
	if hourlyRateStr == "" {
		return
	}
	hourlyRate, err := strconv.Atoi(hourlyRateStr)
	if err != nil {
		fmt.Println("- Hourly rate NaN. Back to main menu.")
		return
	}

	// ratings
	fmt.Print("Number of ratings: ")
	scanner.Scan()
	ratingStr := scanner.Text()
	if ratingStr == "" {
		return
	}
	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		fmt.Println("- Area NaN. Back to main menu.")
		return
	}

	// room writeup
	fmt.Print("Room writeup: ")
	scanner.Scan()
	writeUp := scanner.Text()
	// unlike other fields, indicate no room writeup given expliclity
	if writeUp == "" {
		writeUp = "(No room writeup)"
	}

	fmt.Println()

	// create temporary venue object
	// WARNING: THIS IS NOT THE VENUE OBJECT TO BE USED
	// func NewVenue() will create the one in vASL that you will use
	v := &Venue{
		name:             name,
		region:           region,
		capacity:         capacity,
		categoryID:       categoryID, // refers to category slice
		roomTypeID:       roomTypeID, // refers to roomType slice
		area:             area,
		hourlyRate:       hourlyRate,
		rating:           rating,
		writeUp:          writeUp,
		waitlist:         &PriorityList{},
		approvedBookings: MakeBookings(),
	}

	// ask admin to confirm add venue
	fmt.Println(v)
	fmt.Print("Confirm add this venue? Y/N ")
	scanner.Scan()
	addVenue := strings.Title(scanner.Text())
	switch addVenue {
	case "Y":
		_, err := NewVenue(name, region, capacity, categoryID,
			roomTypeID, area, hourlyRate,
			rating, writeUp, &wg)
		if err != nil { // just in case
			fmt.Println(err)
			fmt.Println("- Venue not added")
		} else {
			fmt.Println("- Venue added")
		}
	default:
		fmt.Println("- No venue added")

	}

}

// PrepareUpdateVenue() updates the venue
// by deleting the venue node and adds a new venue node
func PrepareUpdateVenue() {
	fmt.Println("Back to main menu? Just press enter.")

	// ask venue details

	// venue name
	fmt.Print("Update which venue: ")
	scanner.Scan()
	oldVenueName := scanner.Text()
	// assume that all venues have to be Title Case
	oldVenueName = strings.Title(oldVenueName)
	if oldVenueName == "" {
		return
	}
	gotVenue := venuesAVL.Find(oldVenueName)
	// does not exist?
	if gotVenue == nil {
		fmt.Printf("- %s does not exist\n", oldVenueName)
		return
	}
	// if venue found, print it out
	oldVenue := gotVenue.venue
	fmt.Println()
	fmt.Println("- Venue found: ")
	fmt.Print(oldVenue)

	// ask for updates to venue details
	fmt.Println("Leave empty if you want to retain the values.")
	// venue name
	fmt.Print("New name of venue: ")
	scanner.Scan()
	name := scanner.Text()
	// assume that all venues have to be Title Case
	name = strings.Title(name)
	if name == "" {
		name = oldVenue.name
	}
	// just in case user does not want to change venue name
	// but accidentally keys in another existing venue
	if name != oldVenueName {
		gotVenue = venuesAVL.Find(name)
		// venue already exists
		if gotVenue != nil {
			fmt.Printf("- %s already exists\n", name)
			return
		}
	}

	// region
	fmt.Print("Region: ")
	scanner.Scan()
	region := scanner.Text()
	region = strings.Title(region)
	if region == "" {
		region = oldVenue.region
	}

	// capacity
	var capacity int
	var err error
	fmt.Print("Capacity: ")
	scanner.Scan()
	capacityStr := scanner.Text()
	// if not empty for capacity Str
	if capacityStr != "" {
		capacity, err = strconv.Atoi(capacityStr)
		if err != nil {
			fmt.Println("- Capacity NaN. Back to main menu.")
			return
		}
	} else {
		capacity = oldVenue.capacity
	}

	// for category, list all categories,
	// then user types category name,
	// then sequential search
	var categoryID int
	var gotCategory bool
	fmt.Print("Available categories: ")
	GetVenueCategories()
	fmt.Println()
	fmt.Print("Category: ")
	scanner.Scan()
	category := scanner.Text()
	if category == "" {
		categoryID = oldVenue.categoryID
	} else {
		category = strings.Title(category)
		categoryID, gotCategory = CheckIfVenueCategoryExists(category)
		if !gotCategory {
			fmt.Printf("- %s not a valid category. Back to main menu.\n", category)
			return
		}
	}

	// roomtypes behave like category
	var roomTypeID int
	var gotRoomType bool
	fmt.Print("Available room types: ")
	GetRoomTypes()
	fmt.Println()
	fmt.Print("Room type: ")
	scanner.Scan()
	roomType := scanner.Text()
	if roomType == "" {
		roomTypeID = oldVenue.roomTypeID
	} else {
		roomType = strings.Title(roomType)
		roomTypeID, gotRoomType = CheckIfRoomTypeExists(roomType)
		if !gotRoomType {
			fmt.Printf("- %s not a valid room type. Back to main menu.\n", roomType)
			return
		}
	}

	// area
	var area int
	fmt.Print("Area (sq m): ")
	scanner.Scan()
	areaStr := scanner.Text()
	if areaStr != "" {
		area, err = strconv.Atoi(areaStr)
		if err != nil {
			fmt.Println("- Area NaN. Back to main menu.")
			return
		}
	} else {
		area = oldVenue.area
	}

	// hourly rate
	var hourlyRate int
	fmt.Print("Hourly Rate: SGD")
	scanner.Scan()
	hourlyRateStr := scanner.Text()
	if hourlyRateStr != "" {
		hourlyRate, err = strconv.Atoi(hourlyRateStr)
		if err != nil {
			fmt.Println("- Hourly rate NaN. Back to main menu.")
			return
		}
	} else {
		hourlyRate = oldVenue.hourlyRate
	}

	// ratings
	var rating int
	fmt.Print("Number of ratings: ")
	scanner.Scan()
	ratingStr := scanner.Text()
	if ratingStr != "" {
		rating, err = strconv.Atoi(ratingStr)
		if err != nil {
			fmt.Println("- Area NaN. Back to main menu.")
			return
		}
	} else {
		rating = oldVenue.rating
	}

	// room writeup
	fmt.Print("Room writeup (Type '-' to mean no writeup for this venue.): ")
	scanner.Scan()
	writeUp := scanner.Text()
	// unlike other fields, indicate no room writeup given expliclity
	if writeUp == "-" {
		writeUp = "(No room writeup)"
	} else if writeUp == "" {
		writeUp = oldVenue.writeUp
	}

	fmt.Println()

	// create standby Venue object
	v := &Venue{
		name:             name,
		region:           region,
		capacity:         capacity,
		categoryID:       categoryID, // refers to category slice
		roomTypeID:       roomTypeID, // refers to roomType slice
		area:             area,
		hourlyRate:       hourlyRate,
		rating:           rating,
		writeUp:          writeUp,
		waitlist:         oldVenue.waitlist,
		approvedBookings: oldVenue.approvedBookings,
	}

	// ask admin to confirm add venue
	fmt.Println(v)
	fmt.Print("Update this venue as such? Y/N ")
	scanner.Scan()
	addVenue := strings.Title(scanner.Text())
	switch addVenue {
	case "Y":

		// remove old venue name, then add new venue name
		vTrie.Delete(oldVenueName) // This is working...
		vTrie.Put(v.name, &wg)
		// successfully add new venue for update
		venuesAVL.Remove(&venuesAVL.root, oldVenueName) // TODO: Remove is not working...
		venuesAVL.Insert(&venuesAVL.root, v)

		// update waiting list to reflect new venue
		wNode := v.waitlist.front
		for wNode != nil {
			wNode.booking.target.venue = v.name
			wNode = wNode.next
		}

		fmt.Println("- Venue updated")

		// remember to delete old venue

	default:
		fmt.Println("- No venue updated")
	}

}

func PrepareDeleteVenue() {
	fmt.Println("Back to main menu? Just press enter.")

	// venue name
	fmt.Print("Delete which venue: ")
	scanner.Scan()
	tgtVenueName := scanner.Text()
	// assume that all venues have to be Title Case
	tgtVenueName = strings.Title(tgtVenueName)
	if tgtVenueName == "" {
		return
	}
	gotVenue := venuesAVL.Find(tgtVenueName)
	// does not exist?
	if gotVenue == nil {
		fmt.Printf("- %s does not exist\n", tgtVenueName)
		return
	}
	// if venue found, print it out
	tgtVenue := gotVenue.venue
	fmt.Println()
	fmt.Println("- Venue found: ")
	fmt.Print(tgtVenue)

	// ask for updates to venue details
	fmt.Printf("Delete venue? Y/N ")
	scanner.Scan()
	deleteVenue := strings.Title(scanner.Text())
	switch deleteVenue {
	case "Y":
		// remove from vTrie and veneusAVL
		vTrie.Delete(tgtVenueName)
		venuesAVL.Remove(&venuesAVL.root, tgtVenueName)

		// cancel this venue's booking slots
		wNode := tgtVenue.waitlist.front
		for wNode != nil {
			wNode.booking.status = cancelled
			NotifyMember(wNode.booking.memberID)
			wNode = wNode.next
		}

		fmt.Println("- Venue deleted")

		// remember to delete old venue

	default:
		fmt.Println("- No venue deleted")
	}

}

// PrepareRetrieveAllBookings lists all bookings from BTrie
// then lets the user sort the results with concurrent mergesort
func PrepareRetrieveAllBookings() {
	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	// prepare sort criterion
	sortCriterion := func(a *Booking, b *Booking) bool { return false }

	// for options 21-24 (show only bookings with a specific status)
	// count num of bookings with such condition
	partialCount := 0

	// make temporary slice (go-routine inside this fn at bookings.go)
	bSlice := MakeBookingSortSlice()

	fmt.Println("List of bookings:")
	fmt.Println(bSlice)
	fmt.Println("- Number of bookings shown: ", len(bSlice.bs))
	for {
		fmt.Println()
		fmt.Println("Either sort by")
		fmt.Println(" 1) A-Z  2) Z-A  for Booking ID")
		fmt.Println(" 3) A-Z  4) Z-A  for Venue")
		fmt.Println(" 5) A-Z  6) Z-A  for Member Last Name")
		fmt.Println(" 7) ASC  8) DESC for MemberID ")
		fmt.Println(" 9) ASC 10) DESC for Membership Tier")
		fmt.Println("11) ASC 12) DESC for Start Time") // start time becoming later or earlier
		fmt.Println("13) ASC 14) DESC for End Time")   // ditto for end time
		fmt.Println("or retrieve bookings that are")
		fmt.Println("21) Cancelled 22) Rejected 23) Pending 24) Approved")
		fmt.Print("Enter your choice: ")
		scanner.Scan()
		choice := scanner.Text()
		switch choice {
		case "1":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.bookingID < b.bookingID
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "2":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.bookingID > b.bookingID
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "3":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.target.venue < b.target.venue
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "4":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.target.venue > b.target.venue
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "5":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return JHBase.members[a.memberID-MIDOffset].lastName < JHBase.members[b.memberID-MIDOffset].lastName
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "6":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return JHBase.members[a.memberID-MIDOffset].lastName > JHBase.members[b.memberID-MIDOffset].lastName
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "7":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.memberID < b.memberID
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "8":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.memberID > b.memberID
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "9":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return JHBase.members[a.memberID-MIDOffset].tier < JHBase.members[b.memberID-MIDOffset].tier
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "10":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return JHBase.members[a.memberID-MIDOffset].tier > JHBase.members[b.memberID-MIDOffset].tier
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "11":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return (a.target.startTime).Before(b.target.startTime)
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "12":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return (a.target.startTime).After(b.target.startTime)
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "13":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return (a.target.endTime).Before(b.target.endTime)
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)
		case "14":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return (a.target.endTime).After(b.target.endTime)
			}
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			fmt.Println(bSlice)

		// only shows cancelled "0" bookings
		case "21":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.status == 0
			}
			partialCount = 0
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			for _, b := range bSlice.bs {
				if b.status != 0 {
					break
				}
				fmt.Print(b)
				partialCount++
			}
			fmt.Println("- Num of cancelled bookings: ", partialCount)

		// only shows rejected "1" bookings
		case "22":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.status == 1
			}
			partialCount = 0
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			for _, b := range bSlice.bs {
				if b.status != 1 {
					break
				}
				fmt.Print(b)
				partialCount++
			}
			fmt.Println("- Num of rejected bookings: ", partialCount)

		// only shows pending "2" bookings
		case "23":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.status == 2
			}
			partialCount = 0
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			for _, b := range bSlice.bs {
				if b.status != 2 {
					break
				}
				fmt.Print(b)
				partialCount++
			}
			fmt.Println("- Num of pending bookings: ", partialCount)

		// only shows approved "4" bookings
		case "24":
			sortCriterion = func(a *Booking, b *Booking) bool {
				return a.status == 3
			}
			partialCount = 0
			bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
			for _, b := range bSlice.bs {
				if b.status != 3 {
					break
				}
				fmt.Print(b)
				partialCount++
			}
			fmt.Println("- Num of approved bookings: ", partialCount)

		// back to main menu
		default:
			fmt.Println("Back to main menu.")
			fmt.Println()
			return

		}
		choice = ""
	}
}

// retrieves bookings in the booking list from a venue
func PrepareRetrieveVenueBookings() {
	fmt.Println("Back to main menu? Just press enter.")

	for {
		fmt.Print("Enter venue to retrieve bookings: ")
		scanner.Scan()
		venue := scanner.Text()
		venue = strings.TrimSpace(venue)
		if venue == "" {
			return
		}

		// make query as go-routine
		// this traverses waiting list from given venue
		wg.Add(1)
		go func() {
			defer wg.Done()
			QueryWaitlist(venue, &wg)
		}()
		wg.Wait()

		venue = ""
	}
}

func PrepareProcessBookings() {
	fmt.Print("Enter venue to process bookings: ")
	scanner.Scan()
	venue := scanner.Text()
	venue = strings.TrimSpace(venue)
	if venue == "" {
		return
	}

	var err error
	// make query as go-routine
	// this traverses waiting list from given venue
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = QueryWaitlist(venue, &wg)
	}()
	wg.Wait()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Would you like to process the waitlist for %s? Y/N \n", venue)
	scanner.Scan()
	wantToProcess := strings.Title(scanner.Text())
	if wantToProcess != "Y" {
		fmt.Println("- Back to main menu")
		return
	}

	// get venue node in venuesAVL
	// a venue node contains a waitlist and a list of approved bookings
	// called .waitlist and .approvedBookings respectively
	// you also need to refer to .slots (Months of Slots)
	vNode := venuesAVL.Find(venue)
	err = vNode.venue.ProcessBookings()
	if err != nil {
		fmt.Println(err)
		return
	}

}

func PrepareRejectBooking() {
	fmt.Println("Back to main menu? Just press enter.")
	for {

		fmt.Print("Enter BookingID to remove: ")
		scanner.Scan()
		bID := scanner.Text()

		if bID == "" {
			return
		}

		// REFACTOR TARGET START (also in bookings.go)
		// find whether booking with that bookingID exists
		bNode, gotBooking := bTrie.BFind(bID)
		if !gotBooking {
			fmt.Printf("Cannot reject booking %s; no such booking", bID)
			continue
		}

		// do not reject a cancelled or rejected booking
		if bNode.booking.status == cancelled || bNode.booking.status == rejected {
			fmt.Printf("%s already rejected or cancelled its booking\n", bID)
			continue
		}
		// REFACTOR TARGET END (also in bookings.go)

		fmt.Println(bNode.booking)

		fmt.Print("Reject booking? Y/N ")
		scanner.Scan()
		wantToReject := strings.Title(scanner.Text())
		if wantToReject != "Y" {
			fmt.Println("- Back to main menu")
			return
		}

		err := RejectBooking(bID, &wg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func PrepareCancelBooking() {
	fmt.Println("Back to main menu? Just press enter.")
	for {
		fmt.Print("Enter BookingID to cancel: ")
		scanner.Scan()
		bID := scanner.Text()

		if bID == "" {
			return // back to main menu
		}

		// find whether booking with that bookingID exists
		bNode, gotBooking := bTrie.BFind(bID)
		if !gotBooking {
			fmt.Printf("Cannot reject booking %s; no such booking", bID)
			continue
		}

		// do not remove a cancelled booking
		if bNode.booking.status == cancelled {
			fmt.Printf("%s already cancelled its booking", bID)
			continue
		}

		fmt.Println(bNode.booking)

		fmt.Print("Cancel booking? Y/N ")
		scanner.Scan()
		wantToCancel := strings.Title(scanner.Text())
		if wantToCancel != "Y" {
			fmt.Println("- Back to main menu")
			return
		}

		err := CancelBooking(bID, &wg)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}
}

// PrepareDeleteCancelledBookings will delete
// all bookings marked as "cancelled"
// although not used in this prototype application,
// in reality this applies to bookings made 3 years ago
// that are marked cancelled
func PrepareDeleteCancelledBookings() {

	// in case things go wrong
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	// retrieve all cancelled bookings
	// prepare sort criterion
	sortCriterion := func(a *Booking, b *Booking) bool { return false }

	// for options 21-24 (show only bookings with a specific status)
	// count num of bookings with such condition
	partialCount := 0

	// make temporary slice (go-routine inside this fn at bookings.go)
	bSlice := MakeBookingSortSlice()

	fmt.Println("List of cancelled bookings:")

	// from PrepareRetrieveAllBookings option 21
	sortCriterion = func(a *Booking, b *Booking) bool {
		return a.status == 0
	}
	partialCount = 0
	bSlice.bs.ParallelMergeSort(bSlice.bs, sortCriterion)
	for _, b := range bSlice.bs {
		if b.status != 0 {
			break
		}
		fmt.Print(b)
		partialCount++
	}
	fmt.Println("- Num of cancelled bookings: ", partialCount)
	// end from PrepareRetrieveAllBookings option 21

	if partialCount == 0 {
		fmt.Println("No cancelled bookings to delete.")
		return
	}

	fmt.Print("Confirm delete ALL cancelled bookings? Y/N ")
	scanner.Scan()
	confirmDelete := strings.Title(scanner.Text())

	// if admin does not want to delete cancelled bookings
	if confirmDelete != "Y" {
		return
	}

	// report number of cancelled bookings deleted
	var counter int
	counter = 0 // number of cancelled bookings deleted

	// delete all cancelled bookings in a go-routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		counter = bTrie.BCancellationsDeleter(bTrie.root)
	}()
	wg.Wait()

	fmt.Printf("- Deleted %d cancelled bookings\n", counter)
}

/*-------------------*/
func AdminProcessBooking(v *Venue) error {
	err := v.ProcessBookings()
	if err != nil {
		return err
	}
	return nil
}

// AdminFindBooking helps to find the booking using only bookingID
// used for editing, rejecting and cancelling bookings
// NB: This is only for admins. Members should use FindBooking(bID, mID)
func AdminFindBooking(bID string) (*Booking, error) {
	// find booking first, then check memberID
	bNode, gotBooking := bTrie.BFind(bID)

	// if no such booking, return error
	if !gotBooking {
		return nil, errors.New("No result. Back to main menu.")
	}
	return bNode.booking, nil
}

// For admins only
// DoAdminEditBooking happens at /jh-admin-edit-booking
// returns the booking, notification=to-member string, error
func AdminDoEditBooking(bID string, searchVenue string,
	startDay, startHour, endDay, endHour int) (*Booking, string, error) {
	var editedBooking *Booking

	// parse to time.Time
	startTimeStr := fmt.Sprintf("%d/06/2021 %d:00", startDay, startHour)
	startTime, err := time.Parse("2/1/2006 15:04", startTimeStr)
	endTimeStr := fmt.Sprintf("%d/06/2021 %d:00", endDay, endHour)
	endTime, err := time.Parse("2/1/2006 15:04", endTimeStr)

	if err != nil {
		return nil, "", err
	}

	// check if venue exists
	_, gotVenue := vTrie.Find(searchVenue)
	if !gotVenue {
		s := fmt.Sprintf("- %s is not a venue", searchVenue)
		return nil, "", errors.New(s)
	}

	// check if start time is before end time
	if !(startTime.Before(endTime)) {
		s := "Start time is not before end time."
		return nil, "", errors.New(s)
	}

	// edit booking also checks for whether startTime is earlier than endTime
	editedBooking, err = EditBooking(
		bID,
		searchVenue,
		startTime,
		endTime,
		2,   // 2 means pending for BookingStatus
		&wg) // to sync concurrent tasks

	// hope no errors, but show error just in case
	if err != nil {
		return nil, "", err
	}

	str := NotifyMember(editedBooking.memberID)

	return editedBooking, str, nil
}

// AdminDoCancelBooking calls DoCancelBooking
// thus identical to members cancelling booking
// but can return error
func AdminDoCancelBooking(bID string) error {
	return DoCancelBooking(bID)
}

// AdminDoRejectBooking enables admin to reject a booking
func AdminDoRejectBooking(bID string) error {

	bNode, gotBooking := bTrie.BFind(bID)
	if !gotBooking {
		s := fmt.Sprintf("Cannot reject booking %s; no such booking", bID)
		return errors.New(s)
	}

	// do not reject a cancelled or rejected booking
	if bNode.booking.status == cancelled || bNode.booking.status == rejected {
		s := fmt.Sprintf("%s already rejected or cancelled its booking\n", bID)
		return errors.New(s)
	}

	err := RejectBooking(bID, &wg)
	if err != nil {
		return err
	}

	return nil
}

// AdminFindMember helps to find the booking using bookingID and memberID
// used for editing and cancelling bookings
// NB: This is only for admins. Members should use FindBooking(bID, mID)
func AdminFindMember(mID int) (*Member, error) {

	// calculate index in JHBase.members
	index := mID - MIDOffset
	// if out of bounds, return nil, error
	if index < 0 || index >= len(JHBase.members) {
		return nil, errors.New("Calling memberID out of range")
	}
	// then find member
	member := JHBase.members[mID-MIDOffset]

	// if member has been deleted, return error
	if member == nil {
		return nil, errors.New("Member already deleted")
	}

	return member, nil
}

// AdminFindMember helps to find the booking using bookingID and memberID
// used for editing and cancelling bookings
// NB: This is only for admins. Members should use FindBooking(bID, mID)
func AdminDoDeleteMember(mID int) error {

	// calculate index in JHBase.members
	index := mID - MIDOffset
	// if out of bounds, return nil, error
	if index < 0 || index >= len(JHBase.members) {
		return errors.New("Calling memberID out of range")
	}
	// then find member
	member := JHBase.members[mID-MIDOffset]

	// if member has been deleted, return error
	if member == nil {
		return errors.New("Member already deleted")
	}

	for _, b := range member.bookings.bs {
		// cancel booking
		// note: this also removes booking from waiting list
		// or approved list from the venue
		err := CancelBooking(b.bookingID, &wg)
		if err != nil {
			fmt.Println(err)
		}
	}

	JHBase.members[mID-MIDOffset] = nil
	return nil

}
