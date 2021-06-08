package jhbs

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Load the pod map, which are stored as 2 CSV Files: SGPodPortals.csv and SGPodRoutes.csv
// solution from https://stackoverflow.com/questions/24999079/reading-csv-file-in-go
func readCsvFile(filePath string) ([][]string, error) {

	f, err := os.Open(filePath)
	defer f.Close() // close file

	if err != nil {
		// exit gracefully if unable to read file
		s := fmt.Sprintf("- Unable to read input file %s\n- %v", filePath, err)
		return nil, errors.New(s)
	}

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		// exit gracefully if unable to parse file
		s := fmt.Sprintf("- Unable to parse file as CSV for %s\n- %v", filePath, err)
		return nil, errors.New(s)
	}

	return records, nil
}

// load Venues, with error return only for unable to open file
// if a field does not have correct form of data, that row is ignored,
// but loadVenues continue with other rows
func loadVenues(filePath string, wg *sync.WaitGroup) error {

	var err error
	fmt.Println(".")
	venuesCSV, err := readCsvFile(filePath) // type [][]string
	if err != nil {
		return err
	}

	fmt.Printf("Loading venues from %s...\n", filePath)

	// prepare map of Portals
	// bool here indicates whether the Portal has been added
	if vTrie == nil {
		vTrie = MakeTrie()
	}

	// venues CSV format:
	// Venue_Num (not used)	Venue_Name	Region	Capacity	Category	Room_Type	Area	Hourly_Rate	Ratings Writeup

	var venueName, region, writeUp string
	var capacity, categoryID, roomTypeID, area, hourlyRate, rating int

	// 0th row are field names and are thus ignored
	// start from 1st row
	for i := 1; i < len(venuesCSV); i++ {

		// venue name (can have numbers, letters and symbols)
		// trim front and back space
		venueName = strings.TrimSpace(venuesCSV[i][1])

		// region name, for now same treatment as venue
		// FUTURE: throws error if got numbers and symbols
		region = strings.TrimSpace(venuesCSV[i][2])

		// capacity
		capacity, err = strconv.Atoi(venuesCSV[i][3])
		if err != nil {
			pfa("- Capacity of venue %s is NaN\n", venueName)
			continue
		}

		// category
		categoryID = -1
		// sequential searches for categories ID (not sorted!)
		for j, v := range venueCategories {
			if venuesCSV[i][4] == v {
				categoryID = j
			}
		}
		// category not found
		if categoryID == -1 {
			pfa("- Category of venue %s is not in list of categories\n", venueName)
			continue
		}

		// room type
		roomTypeID = -1
		for j, v := range venueRoomTypes {
			if venuesCSV[i][5] == v {
				roomTypeID = j
			}
		}
		// room type not found
		if roomTypeID == -1 {
			pfa("- Type of room for venue %s is not in list of room types\n", venueName)
			continue
		}

		// area
		area, err = strconv.Atoi(venuesCSV[i][6])
		if err != nil {
			pfa("- Area of venue %s is NaN\n", venueName)
			continue
		}

		// hourly rate
		hourlyRate, err = strconv.Atoi(venuesCSV[i][7])
		if err != nil {
			pfa("- Hourly rate of venue %s is NaN\n", venueName)
			continue
		}

		// rating
		rating, err = strconv.Atoi(venuesCSV[i][8])
		if err != nil {
			pfa("- Rating for venue %s is NaN\n", venueName)
			continue
		}

		// room writeup
		// region name, same treatment as venue
		writeUp = strings.TrimSpace(venuesCSV[i][9])

		// _, err := NewVenue("Haw Par Villa", "SG", 200, 0, 0, 400, 1000, "18 levels of hell")
		_, err = NewVenue(venueName, // venue_name
			region,     // region
			capacity,   // capacity
			categoryID, // categoryID
			roomTypeID, // roomTypeID
			area,       // area
			hourlyRate, // hourly rate
			rating,     // rating
			writeUp,    // writeup
			wg)         // to sync concurrent tasks
		if err != nil {
			plna(err)
			continue
		}
	}

	fmt.Printf("Done loading %s, num of venues loaded: %d.\n", filePath, vTrie.wordCount)
	fmt.Println(".")

	return nil
}

func loadMembers(filePath string) error {
	var err error
	fmt.Println(".")

	membersCSV, err := readCsvFile(filePath) // type [][]string
	if err != nil {
		return err
	}
	fmt.Printf("Loading members from %s...\n", filePath)

	// No need wg.Add?

	// members CSV format:
	// member_num (not memberID), firstName, lastName, membershipLevel (0-3)
	var tier int
	var mobile int
	// 0th row are field names and are thus ignored
	// start from 1st row
	for i := 1; i < len(membersCSV); i++ {

		// tier
		tier, err = strconv.Atoi(membersCSV[i][3])
		if err != nil {
			fmt.Println(err)
			continue
		}

		// mobile number
		mobile, err = strconv.Atoi(membersCSV[i][4])
		// reject typos
		if err != nil {
			fmt.Println(err)
			continue
		}
		// reject non-Singaporean phone numbers
		if (mobile < 81000000) || (mobile > 98999999) {
			pfa("%d not valid phone number in Singapore\n", mobile)
			continue
		}

		_, err = NewMember(
			membersCSV[i][1],                 // firstname
			membersCSV[i][2],                 // lastname
			MemberTier(tier),                 // tier
			mobile,                           // mobile
			membersCSV[i][5],                 // username
			"",                               // hash
			time.Now().Add(-30*24*time.Hour), // start
			time.Now().Add(-29*24*time.Hour), // lastLogin
		)

		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	fmt.Printf("Done loading %s, num of members loaded: %d.\n", filePath, len(JHBase.members))
	fmt.Println(".")
	return nil
}

func loadBookings(filePath string, wg *sync.WaitGroup) error {
	var err error

	fmt.Println(".")

	bookingsCSV, err := readCsvFile(filePath) // type [][]string
	if err != nil {
		return err
	}
	fmt.Printf("Loading bookings from %s...\n", filePath)

	// members CSV format:
	// booking_num (not bookingID), memberID, startTime in d/m/yyyy hh/min)

	var memberID int
	var venueName string
	var startTime, endTime time.Time

	// BODY OF BOOKINGS
	// 0th row are field names and are thus ignored
	// start from 1st row
	for i := 1; i < len(bookingsCSV); i++ {

		// memberID
		memberID, err = strconv.Atoi(bookingsCSV[i][1])
		if err != nil {
			pfa("- %s is not a memberID\n", bookingsCSV[i][1])
			continue
		}
		// validate memberId only at NewBooking()

		// venue name
		venueName = strings.TrimSpace(bookingsCSV[i][2])

		// start time
		startTime, err = time.Parse("2/1/2006 15:04", bookingsCSV[i][3])
		if err != nil {
			plna("-", err)
			continue
		}

		// end time
		endTime, err = time.Parse("2/1/2006 15:04", bookingsCSV[i][4])
		if err != nil {
			plna("-", err)
			continue
		}

		// wg.Add(1)
		_, err = NewBooking(
			memberID,
			venueName,
			startTime,
			endTime,
			2, // set all '2' for  pending
			// rand.Intn(4), // set random booking statuses (for testing purposes)
			wg) // to sync concurrent tasks
		if err != nil {
			plna(err)
			continue
		}

	}

	fmt.Printf("Done loading %s, num of bookings loaded: %d.\n", filePath, bTrie.wordCount)
	fmt.Println(".")
	return nil
}
