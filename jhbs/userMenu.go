package jhbs

import (
	"errors"
	"fmt"
	"time"
)

/*
func LoadUserMainMenu() {

	var choice string
	for {
		choice = ""
		fmt.Println()
		fmt.Println("Choose a user function:")
		fmt.Println(strings.Repeat("=", 24))
		fmt.Println("1. Browse venue")
		fmt.Println("2. Book venue")
		fmt.Println("3. Edit booking")
		fmt.Println("4. Cancel booking")
		fmt.Print("Your choice: ")
		scanner.Scan()
		choice = scanner.Text()
		fmt.Println()
		switch choice {
		case "1":
			PrepareBrowseVenue()
		case "2":
			PrepareBookVenue()
		case "3":
			PrepareUserEditBooking()
		case "4":
			PrepareUserCancelBooking()
		default:
			return // terminate program
		}

	}

}
*/

func PrepareBrowseVenue() {

	// prepare sort criterion
	sortCriterion := func(a *Venue, b *Venue) bool { return false }

	// make temporary slice (go-routine inside this fn at venues.go)
	vSlice := MakeVenueSortSlice()

	fmt.Println("List of venues:")
	fmt.Println(vSlice)

	fmt.Printf("There are %d venues.\n", len(*vSlice))

	for {
		fmt.Println()
		fmt.Println("Either sort by")
		fmt.Println(" 1) A-Z  2) Z-A  for Name")
		fmt.Println(" 3) ASC  4) DESC for Capacity")
		fmt.Println(" 5) ASC  6) DESC for Area")
		fmt.Println(" 7) ASC  8) DESC for Hourly Rate")
		fmt.Println(" 9) ASC 10) DESC for Rating")
		fmt.Print("Enter your choice: ")
		scanner.Scan()
		choice := scanner.Text()
		switch choice {
		case "1":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.name < b.name
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "2":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.name > b.name
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "3":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.capacity < b.capacity
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "4":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.capacity > b.capacity
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "5":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.area < b.area
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "6":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.area > b.area
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "7":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.hourlyRate < b.hourlyRate
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "8":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.hourlyRate > b.hourlyRate
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "9":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.rating < b.rating
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)
		case "10":
			sortCriterion = func(a *Venue, b *Venue) bool {
				return a.rating > b.rating
			}
			vSlice.ParallelMergeSort(*vSlice, sortCriterion)
			fmt.Println(vSlice)

		default:
			fmt.Println("Back to main menu.")
			fmt.Println()
			return

		}

		choice = ""
	}
}

// BrowseVenue borrows from PrepareBrowseVenue
// sort choice represent which sort criterion to use
func BrowseVenue(choice int) *VenueSortSlice {

	// prepare sort criterion
	sortCriterion := func(a *Venue, b *Venue) bool { return false }

	// make temporary slice (spawns go-routine inside this fn at venues.go)
	vSlice := MakeVenueSortSlice()

	// 10 sort criterions
	// 1) A-Z  2) Z-A  for Name
	// 3) ASC  4) DESC for Capacity
	// 5) ASC  6) DESC for Area
	// 7) ASC  8) DESC for Hourly Rate
	// 9) ASC 10) DESC for Rating
	switch choice {
	case 1:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.name < b.name
		}
	case 2:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.name > b.name
		}
	case 3:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.capacity < b.capacity
		}
	case 4:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.capacity > b.capacity
		}
	case 5:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.area < b.area
		}
	case 6:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.area > b.area
		}
	case 7:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.hourlyRate < b.hourlyRate
		}
	case 8:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.hourlyRate > b.hourlyRate
		}
	case 9:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.rating < b.rating
		}
	case 10:
		sortCriterion = func(a *Venue, b *Venue) bool {
			return a.rating > b.rating
		}
	}
	vSlice.ParallelMergeSort(*vSlice, sortCriterion)
	return vSlice
}

// FindVenue borrows from PrepareBookVenue()
// Returns Venue to book if found, otherwise a stream of string
// that shows closest matches to that search
func FindVenue(searchVenue string) (bool, []string) {

	var out []string

	// find venue in venuesAVL tree
	gotVenue := venuesAVL.Find(searchVenue)

	// got venue? good.
	if gotVenue != nil {
		word := gotVenue.venue.name
		out = append(out, word)
		return true, out
	}

	// venue does not exist? we try to help user autocomplete.
	var ch = make(chan string, 1)

	wg.Add(1)
	// concurrent autocompletes
	go func() {
		defer wg.Done()
		vTrie.AutoComplete(searchVenue, ch)
		for word := range ch {
			out = append(out, word)
		}
	}()
	wg.Wait()
	return false, out

}

// ShowVenue borrows from PrepareBookVenue()
// Returns venue info and available days + timeslots
// if venue not found, also returns an error
// accessed by MemberShownVenue() at jhbshttp.muxMember.go
func ShowVenue(searchVenue string) (*Venue, *Bookings, error) {

	// find venue in venuesAVL tree
	// yes, I know it's a repeat from FindVenue()
	// but random kids can type /jh-member/venue/ and go
	// so still must check
	gotVenue := venuesAVL.Find(searchVenue)

	// cannot find venue
	if gotVenue == nil {
		return nil, nil, errors.New("Venue not found")
	}

	theVenue := gotVenue.venue

	theVenue.approvedBookings.RLock()
	out := theVenue.approvedBookings
	theVenue.approvedBookings.RUnlock()

	return theVenue, out, nil
}

// IsValidBooking borrows from PrepareBookVenue()
// Check whether booking is valid
// Returns booking and error (if booking not valid)
// accessed by MemberShownVenue() at jhbshttp.muxMember.go
func IsValidBooking(venue string, startDay, startHour,
	endDay, endHour, mID int) (*Booking, error) {

	// parse to time.Time
	startTimeStr := fmt.Sprintf("%d/06/2021 %d:00", startDay, startHour)
	startTime, err := time.Parse("2/1/2006 15:04", startTimeStr)
	endTimeStr := fmt.Sprintf("%d/06/2021 %d:00", endDay, endHour)
	endTime, err := time.Parse("2/1/2006 15:04", endTimeStr)

	// NB: NewBooking has a fn that checks
	// whether startTime is later than endTime
	// throws err if new booking is unsuccessful
	var newBooking *Booking
	// wg.Add(1)
	newBooking, err = NewBooking(
		mID,
		venue,
		startTime,
		endTime,
		2,   // 2 means pending for BookingStatus
		&wg) // to sync concurrent tasks
	wg.Wait()
	// hope no errors, but show error just in case
	if err != nil {
		return nil, err
	}
	// booking is successful
	return newBooking, nil
}

/*
func PrepareBookVenue() {
	// in case I screw up the day offset
	// coz June is 1-30, but array for June indexes 0-29
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			fmt.Println("Check if your day[] query has offset by -1 (error by 1).")
		}
	}()

	var err error
	fmt.Println("Back to main menu? Just press enter.")
	// venue name
	for {
		fmt.Print("Find venue: ")
		scanner.Scan()
		searchVenue := scanner.Text()
		searchVenue = strings.Title(searchVenue)
		// if empty, go back to main menu
		if searchVenue == "" {
			return
		}
		gotVenue := venuesAVL.Find(searchVenue)
		// venue does not exist? we try to help user autocomplete.
		if gotVenue == nil {
			fmt.Printf("- %s does not exist but possible venues: \n", searchVenue)
			var ch = make(chan string, 1)
			wg.Add(1)
			// concurrent autocompletes
			go func() {
				defer wg.Done()
				vTrie.AutoComplete(searchVenue, ch)
				for word := range ch {
					fmt.Println(word)
				}
			}()
			wg.Wait()
			fmt.Printf("- If no possible venues, you can search for another venue\n")
			continue
		}

		// get availability of a particular date
		var day int
		fmt.Print("What days of Jun 2021? Type -1 for all days, or separate your days one space each. ")
		scanner.Scan()
		daysStr := scanner.Text()
		if daysStr == "" {
			return
		}
		if daysStr == "-1" {
			fmt.Println()
			// show available hours
			// note-to-self: use daysInAMonth[mth] in the future!
			for day := 1; day <= 30; day++ {
				fmt.Printf("Available slots for %s on 2021-06-%02d - ", searchVenue, day)
				fmt.Println(gotVenue.venue.slots.head.available[day-1])
			}
			fmt.Println()
		} else {
			daysStrSlice := strings.Split(daysStr, " ")
			days := []int{}
			for _, d := range daysStrSlice {
				day, err = strconv.Atoi(d)
				if err != nil {
					fmt.Println("- Day NaN. Back to main menu.")
					return
				}
				// in June, days are 1-30 but 0-29 for array
				if day < 1 || day > 30 {
					fmt.Println("- Day NaN. Back to main menu.")
					return
				}
				days = append(days, day)
			}

			// show available hours
			for _, day := range days {
				fmt.Printf("Available slots for %s on 2021-06-%02d - ", searchVenue, day)
				fmt.Println(gotVenue.venue.slots.head.available[day-1])
			}
			fmt.Println()
		}

		var startTimeStr, endTimeStr string
		var startDay, startHour int
		var endDay, endHour int
		var startTime, endTime time.Time
		// ask user for starttime and endtime
		fmt.Println("NB: Booking slots are hourly.")
		fmt.Println("You may book a slot that has been occupied, in case the other person cancels its booking. However, all bookings still have to be approved by admin.")
		fmt.Print("Booking start date (1-30) and hour (0-23), separated by a space (e.g. for 30 Jun, 1 PM, type 30 13): ")
		scanner.Scan()
		startTimeStr = scanner.Text()
		startDay, startHour, err = checkUserInputForMonthAndHour(startTimeStr, "start")
		// if error, ask user to book venue again
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Print("Booking end date (1-30) and hour (0-23), separated by a space (e.g. for 30 Jun, 2 PM, type 30 14): ")
		scanner.Scan()
		endTimeStr = scanner.Text()
		endDay, endHour, err = checkUserInputForMonthAndHour(endTimeStr, "end")
		// if error, ask user to book venue again
		if err != nil {
			fmt.Println(err)
			continue
		}
		// parse to time.Time
		startTimeStr = fmt.Sprintf("%d/06/2021 %d:00", startDay, startHour)
		startTime, err = time.Parse("2/1/2006 15:04", startTimeStr)
		endTimeStr = fmt.Sprintf("%d/06/2021 %d:00", endDay, endHour)
		endTime, err = time.Parse("2/1/2006 15:04", endTimeStr)

		fmt.Println()

		// IMPT Self-note: For now, user's memberID is
		// the last one in the Members' slice.
		// when live, please change this so that
		// the member ID does not get changed.
		mID := len(JHBase.members) + MIDOffset - 1

		// NB: if startTime is later than endTime,
		// this booking becomes undone;
		// user will be asked to re-make a booking.
		var newBooking *Booking
		// wg.Add(1)
		newBooking, err = NewBooking(
			mID,
			searchVenue,
			startTime,
			endTime,
			2,   // 2 means pending for BookingStatus
			&wg) // to sync concurrent tasks
		wg.Wait()
		// hope no errors, but show error just in case
		if err != nil {
			fmt.Println(err)
			continue
		}
		// booking is successful
		fmt.Printf("We have received your booking. Here are the details:\n")
		fmt.Print(newBooking)
		fmt.Println("- Please copy your bookingID for reference:")
		fmt.Println()
		fmt.Println("          ", newBooking.bookingID)
		fmt.Println()
		fmt.Println("- Owing to high server load, please wait for 3 working days for our admin to approve your booking.")
		fmt.Println("- You will be notified whether your booking is approved or rejected.")
		return
	}
}
*/
/*
// PrepareUserEditBooking() enables a user to edit its booking
// Requires bookingID and memberID for verification
// so that random ppl don't anyhow access that bookingID
func PrepareUserEditBooking() {

	// enter BookingID
	fmt.Print("Your BookingID: ")
	scanner.Scan()
	bID := scanner.Text()

	if bID == "" {
		return
	}

	// for verification purposes, enter memberID
	fmt.Print("Your memberID: ")
	scanner.Scan()
	mIDStr := scanner.Text()

	if mIDStr == "" {
		return
	}

	mID, err := strconv.Atoi(mIDStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// find booking first, then check memberID
	bNode, gotBooking := bTrie.BFind(bID)

	// if no such booking, incorrect memberID, or booking alr cancelled,
	// display same message
	// don't tell other people that other bookings with that code exists
	if (!gotBooking) || (bNode.booking.memberID != mID) || (bNode.booking.status == 0) {
		fmt.Println("- No result. Back to main menu.")
		return
	}

	oldBooking := bNode.booking
	fmt.Println(oldBooking)

	// LIKE PrepareUserBookVenue, but slightly different
	for {
		fmt.Print("Find venue: ")
		scanner.Scan()
		searchVenue := scanner.Text()
		searchVenue = strings.Title(searchVenue)
		// if empty, go back to main menu
		if searchVenue == "" {
			searchVenue = oldBooking.target.venue
		}
		// convert searchVenue string to *vAVLNode
		gotVenue := venuesAVL.Find(searchVenue)
		// venue does not exist? we try to help user autocomplete.
		if gotVenue == nil {
			fmt.Printf("- %s does not exist but possible venues: \n", searchVenue)
			var ch = make(chan string, 1)
			wg.Add(1)
			// concurrent autocompletes
			go func() {
				defer wg.Done()
				vTrie.AutoComplete(searchVenue, ch)
				for word := range ch {
					fmt.Println(word)
				}
			}()
			wg.Wait()
			fmt.Printf("- If no possible venues, you can search for another venue\n")
			continue
		}

		// user most likely would choose a new day and hour
		// and even if not, still has to check venue availability
		// so don't copy from oldBooking.target.startTime and oldBooking.target.endTime

		// get availability of a particular date
		var day int
		fmt.Print("What days of Jun 2021? Type -1 for all days, or separate your days one space each. ")
		scanner.Scan()
		daysStr := scanner.Text()
		if daysStr == "" {
			return
		}
		if daysStr == "-1" {
			fmt.Println()
			// show available hours
			// note-to-self: use daysInAMonth[mth] in the future!
			for day := 1; day <= 30; day++ {
				fmt.Printf("Available slots for %s on 2021-06-%02d - ", searchVenue, day)
				fmt.Println(gotVenue.venue.slots.head.available[day-1])
			}
			fmt.Println()
		} else {
			daysStrSlice := strings.Split(daysStr, " ")
			days := []int{}
			for _, d := range daysStrSlice {
				day, err = strconv.Atoi(d)
				if err != nil {
					fmt.Println("- Day NaN. Back to main menu.")
					return
				}
				// in June, days are 1-30 but 0-29 for array
				if day < 1 || day > 30 {
					fmt.Println("- Day NaN. Back to main menu.")
					return
				}
				days = append(days, day)
			}

			// show available hours
			for _, day := range days {
				fmt.Printf("Available slots for %s on 2021-06-%02d - ", searchVenue, day)
				fmt.Println(gotVenue.venue.slots.head.available[day-1])
			}
			fmt.Println()
		}

		// ask user for starttime and endtime
		var startTimeStr, endTimeStr string
		var startDay, startHour int
		var endDay, endHour int
		var startTime, endTime time.Time

		fmt.Println("NB: Booking slots are hourly.")
		fmt.Print("Booking start date (1-30) and hour (0-23), separated by a space (e.g. for 30 Jun, 1 PM, type 30 13): ")
		scanner.Scan()
		startTimeStr = scanner.Text()
		startDay, startHour, err = checkUserInputForMonthAndHour(startTimeStr, "start")
		// if error, ask user to book venue again
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Print("Booking end date (1-30) and hour (0-23), separated by a space (e.g. for 30 Jun, 2 PM, type 30 14): ")
		scanner.Scan()
		endTimeStr = scanner.Text()
		endDay, endHour, err = checkUserInputForMonthAndHour(endTimeStr, "end")
		// if error, ask user to book venue again
		if err != nil {
			fmt.Println(err)
			continue
		}
		// parse to time.Time
		startTimeStr = fmt.Sprintf("%d/06/2021 %d:00", startDay, startHour)
		startTime, err = time.Parse("2/1/2006 15:04", startTimeStr)
		endTimeStr = fmt.Sprintf("%d/06/2021 %d:00", endDay, endHour)
		endTime, err = time.Parse("2/1/2006 15:04", endTimeStr)

		fmt.Println()

		// NB: if startTime is later than endTime,
		// this booking becomes undone;
		// user will be asked to re-make a booking.
		var editedBooking *Booking

		editedBooking, err = EditBooking(
			bID,
			mID,
			searchVenue,
			startTime,
			endTime,
			2,   // 2 means pending for BookingStatus
			&wg) // to sync concurrent tasks

		// hope no errors, but show error just in case
		if err != nil {
			fmt.Println(err)
			continue
		}
		// booking is successful
		fmt.Printf("We have edited your booking. Here are the details:\n")
		fmt.Print(editedBooking)
		fmt.Println("- Please note your bookingID for reference:")
		fmt.Println()
		fmt.Println("          ", editedBooking.bookingID)
		fmt.Println()
		fmt.Println("- Owing to high server load, please wait for 3 working days for our admin to approve your booking.")
		fmt.Println("- You will be notified whether your edited booking is approved or rejected.")
		return
	}

}
*/

// FindBooking helps to find the booking using bookingID and memberID
// used for editing and cancelling bookings
// reason for both IDs is to prevent random ppl from deleting other ppl's bookings
// NB: This is only for members. Admins should use AdminFindBooking(bID)
func FindBooking(bID string, mID int) (*Booking, error) {
	// find booking first, then check memberID
	bNode, gotBooking := bTrie.BFind(bID)

	// if no such booking, incorrect memberID, or booking alr cancelled,
	// display same message
	// don't tell other people that other bookings with that code exists
	if (!gotBooking) || (bNode.booking.memberID != mID) || (bNode.booking.status == 0) {
		return nil, errors.New("No result. Back to main menu.")
	}
	return bNode.booking, nil
}

// For members only
// DoEditBooking happens at /jh-member-edit-booking
func DoEditBooking(bID string, mID int, searchVenue string,
	startDay, startHour, endDay, endHour int) (*Booking, error) {
	var editedBooking *Booking

	// parse to time.Time
	startTimeStr := fmt.Sprintf("%d/06/2021 %d:00", startDay, startHour)
	startTime, err := time.Parse("2/1/2006 15:04", startTimeStr)
	endTimeStr := fmt.Sprintf("%d/06/2021 %d:00", endDay, endHour)
	endTime, err := time.Parse("2/1/2006 15:04", endTimeStr)

	if err != nil {
		return nil, err
	}

	// returns if invalid memberID, wrong venue or statTime later than endTime
	err = validateBookingFields(mID, searchVenue,
		startTime, endTime)

	if err != nil {
		return nil, err
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
		return nil, err
	}
	return editedBooking, nil
}

// DoCancelBooking happens at /jh-member-cancel-booking
func DoCancelBooking(bID string) error {
	err := CancelBooking(bID, &wg)
	if err != nil {
		return err
	}
	return nil
}

/*
// PrepareUserCancelBooking() enables a user to cancel its booking
// Requires bookingID and memberID for verification
// so that random ppl don't anyhow access that bookingID
func PrepareUserCancelBooking() {

	// enter BookingID
	fmt.Print("Your BookingID: ")
	scanner.Scan()
	bID := scanner.Text()

	if bID == "" {
		return
	}

	// for verification purposes, enter memberID
	fmt.Print("Your memberID: ")
	scanner.Scan()
	mIDStr := scanner.Text()

	if mIDStr == "" {
		return
	}

	mID, err := strconv.Atoi(mIDStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// find booking first, then check memberID
	bNode, gotBooking := bTrie.BFind(bID)

	// if no such booking, incorrect memberID, or booking alr cancelled,
	// display same message
	// don't tell other people that other bookings with that code exists
	if (!gotBooking) || (bNode.booking.memberID != mID) || (bNode.booking.status == 0) {
		fmt.Println("- No result. Back to main menu.")
		return
	}

	fmt.Println(bNode.booking)

	fmt.Print("Confirm cancel booking? Y/N ")
	scanner.Scan()
	deleteBooking := strings.Title(scanner.Text())
	switch deleteBooking {
	case "Y":
		// spawns 2 go-routines to cancel booking
		// 1 on waiting list for that venue
		// another 1 on approved bookings for that venue
		// see more at CancelBooking
		err := CancelBooking(bID, &wg)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println()
	default:
		fmt.Println("- No booking deleted")
	}

}
*/
