package jhbshttp

import (
	"TimothyTAN_GoInAction1/jhbs"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	uuid "github.com/satori/go.uuid"
)

// mux.go - the servemux

// main menu can be accessed at /
func MainMenu(w http.ResponseWriter, req *http.Request) {

	// ask for admin cookie
	_, err := req.Cookie("jhAdminCookie")
	// if got cookie already, that means admin already logged in
	if err == nil {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	}

	// then ask for member cookie
	jhMemberCookie, err := req.Cookie("jhMemberCookie")
	// if got cookie already, that means already logged in
	if err == nil {
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	}

	// for June Holidays booking system
	if encoded, err := s.Encode("jhMemberCookie", time.Now().UnixNano()); err == nil {
		jhMemberCookie = &http.Cookie{
			Name:    "jhMemberCookie",
			Value:   encoded,
			Expires: time.Now().Add(24 * time.Hour), // expires in 1 day
			// Secure:   true,
			// HttpOnly: true,
		}
		// http.SetCookie(w, jhMemberCookie)
	}
	// }
	var myUser *jhbs.Member

	//---try to locate the user based on cookie's value---
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhMemberCookie.Value, &cCode); err == nil {
		if mli, ok := memberSessions.Load(cCode.String()); ok {
			myUser = jhbs.JHBase.Members()[mli.memberID-jhbs.MIDOffset]
		}
	}

	// parse template file (.gohtml or other extension)
	// executes template from index.gohtml
	// NB: templates don't enjoy pointers to nil; use nil value directly
	if myUser == nil {
		err = tpl.ExecuteTemplate(w, "index.gohtml", nil)
	} else {
		err = tpl.ExecuteTemplate(w, "index.gohtml", *myUser)
	}
	if err != nil {
		log.Fatalln(err)
	}
}

// signup is for new users
func Signup(w http.ResponseWriter, req *http.Request) {

	// -- try to locate cookie
	jhMemberCookie, err := req.Cookie("jhMemberCookie")

	// if already have cookie, redirect user to user-only area
	if err == nil {
		fmt.Println("User already has cookie; see if legitimate.")
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	}

	// or else set cookie handler
	// id := uuid.NewV4() // some use uuid for the cookie's value
	// for June Holidays booking system
	if encoded, err := s.Encode("jhMemberCookie", time.Now().UnixNano()); err == nil {
		jhMemberCookie = &http.Cookie{
			Name:    "jhMemberCookie",
			Value:   encoded,
			Expires: time.Now().Add(24 * time.Hour), // expires in 1 day
			// Secure:   true,
			// HttpOnly: true,
		}
		// http.SetCookie(w, jhMemberCookie)
	}

	var myUser *jhbs.Member

	//---try to locate the user based on cookie's value---
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhMemberCookie.Value, &cCode); err == nil {
		if mli, ok := memberSessions.Load(cCode.String()); ok {
			myUser = jhbs.JHBase.Members()[mli.memberID-jhbs.MIDOffset]
		}
	}

	// if user performs a post (via the Submit button)---
	if req.Method == http.MethodPost {

		//---form validation---
		var wg sync.WaitGroup
		var fn, ln, u, p string // firstname (fn), lastname (ln), username (u)
		var tier, mobile int
		errs := make([]error, 6)

		// WARNING: don't do go-routines here coz it may cause fatal errors
		// validate firstname (fn)
		fn, errs[0] = validateString(req, "firstname")
		// validate lastname (fn)
		ln, errs[1] = validateString(req, "lastname")
		// validate membership tier (<input> uses value=(numbers))
		tier, errs[2] = validateTier(req)
		// validate mobile
		mobile, errs[3] = validateMobile(req)
		// validate username (u)
		// requires "username" in <input>
		u, errs[4] = validateUsername(req)
		// validate password (p)
		p, errs[5] = validatePassword(req) // requires "code" in <input>

		if errs[0] != nil || errs[1] != nil { // invalid firstname and lastname
			showErrorOnTop(w, "signup.gohtml", nil,
				"Invalid name")
			return
		}

		if errs[2] != nil { // invalid tier
			showErrorOnTop(w, "signup.gohtml", nil, errs[2].Error())
			return
		}

		if errs[3] != nil { // invalid mobile
			showErrorOnTop(w, "signup.gohtml", nil, errs[3].Error())
			return
		}

		// check if username alr taken
		usernameAlreadyTaken := false
		for _, m := range jhbs.JHBase.Members() {
			if u == m.UserName() {
				usernameAlreadyTaken = true
				break
			}
		}
		if usernameAlreadyTaken {
			showErrorOnTop(w, "signup.gohtml", nil,
				`Username already taken`)
			return
		}

		// show error of username and/or password
		if errs[4] != nil || errs[5] != nil {
			showErrorOnTop(w, "signup.gohtml",
				nil, `
				Invalid username and/or password<br/>
				Allowed characters for username: a-z, A-Z, 0-9 and -<br/>
				Allowed characters for password: a-z, A-Z, 0-9 and *!@#$%^&(){}[]:;,.?/~_+-=|\<br/>
				`)
			return
		}

		// check if password is valid
		// from pwValidator.go
		isValidPassword := IsValidPassword(p)
		if !isValidPassword {
			showErrorOnTop(w, "signup.gohtml",
				nil, `
				Password not strong enough<br/>
				Password requires min. 1 uppercase letter, 1 lowercase letter, 1 digit, 1 symbol, and be at least 8 characters long.<br/>
				Allowed characters for password: a-z, A-Z, 0-9 and *!@#$%^&(){}[]:;,.?/~_+-=|\<br/>
				`)
			return
		}
		//---end form validation---

		wg.Add(2)
		var hash string
		var loginTime time.Time

		// then hash code
		go func() {
			defer wg.Done()
			hash, _ = HashPassword(p)
		}()
		go func() {
			defer wg.Done()
			loginTime = time.Now()
		}()
		wg.Wait()

		newMember, err := jhbs.NewMember(fn, ln, jhbs.MemberTier(tier), mobile, u, hash, loginTime, loginTime)
		if err != nil { // invalid mobile
			showErrorOnTop(w, "signup.gohtml", nil, err.Error())
			return
		}

		// set cookie handler
		cCode := uuid.NewV4() // cCode, the uuid for the cookie's value
		jhMemberCookie.Value, _ = s.Encode("jhMemberCookie", cCode)
		fmt.Println("The value of signup cookie:", jhMemberCookie.Value)

		// compose a memberLoginInfo struct,
		// then add it to memberSessions, with the cookie value mapped to the struct
		memberSessions.Store(cCode.String(), &memberLoginInfo{
			newMember.ID(),
			req.RemoteAddr, // more reliable to use remote address,
			loginTime,
		})
		http.SetCookie(w, jhMemberCookie)

		// "Thanks for signing up; redirecting you to members' page"
		http.Redirect(w, req, "/jh-signedup", http.StatusSeeOther)

		// TODO: Store new user information
		// (when merging with GoAdvanced assignment)
	}

	// parse template file (.gohtml or other extension)
	// executes template from index.gohtml
	err = tpl.ExecuteTemplate(w, "signup.gohtml", myUser)
	if err != nil {
		log.Fatalln(err)
	}

}

// signup is for new users
func Signedup(w http.ResponseWriter, req *http.Request) {

	// -- try to locate cookie
	_, err := req.Cookie("jhMemberCookie")

	// if don't have cookie, redirect user to main menu
	if err != nil {
		fmt.Println("No cookie called jhMemberCookie")
		io.WriteString(w, `
		<html>
			<meta http-equiv="refresh" content="2;url=/jh-member" />
			<body>
			<div class="redirect-message">
				<h2>Something has gone wrong. You will be re-directed to the home page in 2 seconds...</h2>
			</div>
			<link href="css/style.css" type="text/css" rel="stylesheet">
			</body>
		</html>
		`)
		return
	}

	io.WriteString(w, `
	<html>
		<meta http-equiv="refresh" content="2;url=/jh-member" />
		<body style="text-align: center; display: block;">
		<div class="redirect-message">
			<h2>Thank you for signing up! Bringing you to the members' page...</h2>
		</div>
		<link href="css/style.css" type="text/css" rel="stylesheet">
		</body>
	</html>
	`)
}

// memberLogin is when members login
// in memberlogin.gohtml (called by index.gohtml),
// there is <form method="POST" action="/jh-member-login">
// and /jh-member-login calls memberLogin
func MemberLogin(w http.ResponseWriter, req *http.Request) {

	// validate username (u) and code (p)
	u, err1 := validateString(req, "username")
	if err1 != nil {
		showErrorOnTop(w, "index.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// points myUser to a member
	var myUser *jhbs.Member
	for _, member := range jhbs.JHBase.Members() {
		if member.UserName() == u {
			myUser = member
			break
		}
	}
	// if user not found
	if myUser == nil {
		showErrorOnTop(w, "index.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// validate password (p)
	p, err2 := validatePassword(req) // requires "code" in <input>
	if err2 != nil {
		showErrorOnTop(w, "index.gohtml", nil,
			`Invalid username and/or password`)
		return
	}
	// check is pw is correct
	isCorrectCode := CheckPasswordHash(p, myUser.Hash())
	if !isCorrectCode {
		showErrorOnTop(w, "index.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// update login time
	loginTime := time.Now()
	myUser.SetLastLogin(loginTime)
	jhbs.JHBase.Members()[myUser.ID()-jhbs.MIDOffset].SetLastLogin(loginTime)

	// Create cookie
	jhMemberCookie, err := req.Cookie("jhMemberCookie")

	// then encode it as the value for a new jhMemberCookie
	if err != nil {
		cCode := uuid.NewV4()
		encoded, err := s.Encode("jhMemberCookie", cCode)
		if err == nil {
			fmt.Println("Encoded value:", encoded)
			jhMemberCookie = &http.Cookie{
				Name:    "jhMemberCookie",
				Value:   encoded,
				Expires: loginTime.Add(24 * time.Hour), // expires in 1 day
				// Secure:   true,
				// HttpOnly: true,
			}
			// compose a memberLoginInfo struct,
			// then add it to memberSessions, with the cookie value mapped to the struct
			memberSessions.Store(cCode.String(), &memberLoginInfo{
				myUser.ID(),
				req.RemoteAddr, // more reliable to use remote address,
				loginTime,
			})
			http.SetCookie(w, jhMemberCookie)
		}
	}
	// fmt.Println("Cookie name during login:", jhMemberCookie.Name)

	// then redirect user to member's page after login
	io.WriteString(w, `
	<html>
		<meta http-equiv="refresh" content="2;url=/jh-member" />
		<body style="text-align: center; display: block;">
		<div class="redirect-message">
			<h2>Welcome back `+myUser.FirstName()+` `+myUser.LastName()+`!</h2>
		</div>
		<link href="css/style.css" type="text/css" rel="stylesheet">
		</body>
	</html>
	`)
}

// membersOnly is only intended for members
// it is the first page members would land upon logging in or signing up
func MemberOnly(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	// based on former jhbs.PrepareBookVenue()
	// happens if user searches for a venue
	if req.Method == http.MethodPost {
		getVenueSearch(w, req, myUser)
		return
	}

	tpl.ExecuteTemplate(w, "membermain.gohtml", myUser)
}

// internal function
// showVenueSearch only applies to when user searches for a venue
// happens at /jh-member and /jh-member/venue/..
// done only as POST
func getVenueSearch(w http.ResponseWriter, req *http.Request, myUser *jhbs.Member) {
	searchVenue, err := validateVenue(req, "venue")

	// ask user to type smth to search venue
	if err != nil {
		showErrorOnTop(w, "membermain.gohtml", myUser,
			err.Error())
		return
	}

	// find the venue
	searchVenue = strings.Title(searchVenue)
	gotVenue, out := jhbs.FindVenue(searchVenue)

	// data for use in template
	// fields must be exported (Field)!
	type Data struct {
		Member   *jhbs.Member
		GotVenue bool
		Venues   []string
	}
	var data = &Data{myUser, gotVenue, out}
	tpl.ExecuteTemplate(w, "membermain.gohtml", data)
	return
}

// MemberShownVenue happens after member clicks on a venue
// Here we show the user the venue info and availability
func MemberShownVenue(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	/*
		// happens if user searches for a venue
		if req.Method == http.MethodPost {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			showVenueSearch(w, req, myUser)
			return
		}
	*/

	// get venue string from URL path
	v := req.URL.Path[len("/jh-member/venue/"):]

	// find if this venue exists (prevent abusers from anyhow typing in this URL)
	v = strings.Replace(v, "_", " ", -1)

	theVenue, _, err := jhbs.ShowVenue(v)
	if err != nil {
		// redirect to prev page visited
		http.Redirect(w, req, req.Header.Get("Referer"), 302)
		return
	}

	// data for use in template
	// fields must be exported (Field)!
	type Data struct {
		Member *jhbs.Member
		Venue  *jhbs.Venue
		// VenueAvail *jhbs.TimesBooked
		Booking *jhbs.Booking
	}
	var data = &Data{myUser, theVenue, nil}

	if req.Method == http.MethodPost {

		v, err := validateVenue(req, "venue")
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		sd, err := validateDay(req, "startday")
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		sh, err := validateHour(req, "starthour")
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		ed, err := validateDay(req, "endday")
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		eh, err := validateHour(req, "endhour")
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		booking, err := jhbs.IsValidBooking(v, sd, sh, ed, eh, myUser.ID())

		// if booking unsuccessful
		if err != nil {
			showErrorOnTop(w, "membershownvenue.gohtml", data,
				err.Error())
			return
		}
		// else if booking successful
		data.Booking = booking
		tpl.ExecuteTemplate(w, "membershownvenue.gohtml", data)
	}

	tpl.ExecuteTemplate(w, "membershownvenue.gohtml", data)

}

// MemberBrowseVenue - when a member browses venues
// a result of all the venues will show up,
// then member sorts the venues slice
// calls memberbrowsevenue.gohtml
func MemberBrowseVenue(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	type Data struct {
		Member   *jhbs.Member
		Venues   *jhbs.VenueSortSlice
		SortCrit string
	}
	var data = &Data{myUser, nil, ""}

	// if GET, ask what sort criterion (1-10) member wants
	if req.Method == http.MethodGet {
		q := req.URL.Query()

		// if user comes in from e.g. /jh-member
		if q["sortcriterion"] != nil {
			// get first member of memberid
			numStr := q["sortcriterion"][0]

			// if no query string, rediret to /jh-member
			if numStr == "" {
				http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
				return
			}
			numStr = template.HTMLEscapeString(numStr)

			// if field is empty
			// convert string to int
			num, err := strconv.Atoi(numStr)

			// if field has 0-9
			if err != nil {
				http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
				return
			}

			// sort criterion in string
			// cross-check with memberbrowsevenue.gohtml
			sortCrit := []string{"from A to Z",
				"from Z to A",
				"by Capacity ASC",
				"by Capacity DESC",
				"by Area ASC",
				"by Area DESC",
				"by Hourly Rate ASC",
				"by Hourly Rate DESC",
				"by Rating ASC",
				"by Rating DESC"}

			// sorting done on the server, not on the client
			// FUTURE CONSIDERATION: Should sorting be done at client side?
			// ADV: Less data transferred btw. server and client.
			// DISADV: Client might not be able to sort fast.
			// make temporary slice (go-routine inside this fn at venues.go)
			// default: A-Z
			vSlice := jhbs.BrowseVenue(num)

			data.Venues = vSlice
			data.SortCrit = sortCrit[num-1] // rmb offset by 1
		}

	}
	tpl.ExecuteTemplate(w, "memberbrowsevenue.gohtml", data)

}

// MemberHistory - when a member views its own booking history
// calls memberhistory.gohtml
func MemberHistory(w http.ResponseWriter, req *http.Request) {
	myUser, err := askForMemberCookie(w, req)

	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	tpl.ExecuteTemplate(w, "memberhistory.gohtml", myUser)

}

// MemberPrepareEditBooking - when a member prepares to a booking
// called by memberhistory.gohtml at <form action="/jh-member-edit-booking">
func MemberPrepareEditBooking(w http.ResponseWriter, req *http.Request) {
	myUser, err := askForMemberCookie(w, req)

	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached MemberPrepareEditBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))

		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.FindBooking(bID, myUser.ID())
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}

		var theVenue *jhbs.Venue
		// var venueAvail *jhbs.TimesBooked
		// get newvenue field if any
		// prevents scripts from loading on newvenue field
		v, err := validateVenue(req, "newvenue")

		if err == nil {
			// show the venue and its availability
			theVenue, _, err = jhbs.ShowVenue(v)
			// if for some reason venue not found, throw err
			if err != nil {
				showErrorOnTop(w, "memberhistory.gohtml", nil,
					`Invalid search`)
				return
			}
		} else if err.Error() == "Empty newvenue" {
			// nothing on newvenue field yet
			theVenue = nil
		} else {
			// improper venue chosen (or scripts lol)
			showErrorOnTop(w, "membermain.gohtml", nil,
				`Invalid search`)
			return
		}

		// get all venues from jhbs
		allVenues := jhbs.MakeVenueSortSlice()

		type Data struct {
			Member    *jhbs.Member  // from cookie
			Booking   *jhbs.Booking // must-have
			AllVenues *jhbs.VenueSortSlice
			NewVenue  *jhbs.Venue
		}
		data := &Data{myUser, theBooking, allVenues, theVenue}
		tpl.ExecuteTemplate(w, "membereditbooking.gohtml", data)

	}
}

// MemberEditBooking() - when user confirms editing booking
func MemberEditBooking(w http.ResponseWriter, req *http.Request) {
	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	} else {
		bID, err := validateString(req, "bookingid")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		v, err := validateVenue(req, "venue")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		sd, err := validateDay(req, "startday")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		sh, err := validateHour(req, "starthour")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		ed, err := validateDay(req, "endday")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		eh, err := validateHour(req, "endhour")
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		// else edit the booking
		theBooking, err := jhbs.DoEditBooking(bID, myUser.ID(), v, sd, sh, ed, eh)
		if err != nil {
			showErrorOnTop(w, "memberhistory.gohtml", myUser,
				err.Error())
			return
		}
		io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="5;url=/jh-member-history" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					<h2>Your booking `+bID+` has been edited to become:</h2>
					<p>`+theBooking.String()+`</p>
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
		`)
	}
}

// MemberPrepareCancelBooking - when a member prepares to cancel a booking
// called by memberhistory.gohtml at <form action="/jh-member-cancel-booking">
func MemberPrepareCancelBooking(w http.ResponseWriter, req *http.Request) {
	myUser, err := askForMemberCookie(w, req)

	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached MemberPrepareCancelBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))

		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.FindBooking(bID, myUser.ID())
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}
		type Data struct {
			Member  *jhbs.Member
			Booking *jhbs.Booking
		}
		data := &Data{myUser, theBooking}
		tpl.ExecuteTemplate(w, "membercancelbooking.gohtml", data)

	}

}

// MemberCancelBooking - when it actually cancels booking
func MemberCancelBooking(w http.ResponseWriter, req *http.Request) {
	myUser, err := askForMemberCookie(w, req)

	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached MemberCancelBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))
		fmt.Println(req.FormValue("cancel"))

		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}
		_, err = jhbs.FindBooking(bID, myUser.ID())
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}

		// validate intent to cancel
		cancelBooking, err := validateString(req, "cancel")
		if err != nil {
			http.Redirect(w, req, "/jh-member", http.StatusSeeOther)
			return
		}

		// actual work of cancelling booking
		switch cancelBooking {
		case "yes":
			jhbs.DoCancelBooking(bID)
			io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="2;url=/jh-member" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					<h2>Your booking `+bID+` has been cancelled.</h2>
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
			`)
		default:
			http.Redirect(w, req, "/jh-member-history", http.StatusSeeOther)
			return
		}
	}
}

// memberProfile allows member to see its profile
// this is the preliminary step before editing its own profile or pw
func MemberProfile(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)

	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	tpl.ExecuteTemplate(w, "memberprofile.gohtml", myUser)

}

// memberProfileEdit for members to edit profile
func MemberProfileEdit(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	if req.Method == http.MethodPost {

		//---form validation---
		mobileStr := req.FormValue("mobile") // match name in <input>
		mobile, _ := strconv.Atoi(mobileStr)

		// validate firstname (fn) and lastname (ln)
		// validate username (u) and code (p)
		fn, err1 := validateString(req, "firstname")
		ln, err2 := validateString(req, "lastname")
		if err1 != nil || err2 != nil {
			showErrorOnTop(w, "memberprofileedit.gohtml",
				myUser,
				"Invalid name")
			return
		}

		// validate mobile
		mobile, err4 := validateMobile(req)
		if err4 != nil {
			showErrorOnTop(w, "memberprofileedit.gohtml",
				myUser,
				err4.Error())
			return
		}

		// update particulars on myUser cookie and membership data
		myUser.SetFirstName(fn)
		myUser.SetLastName(ln)
		myUser.SetMobile(mobile)

		// update membership list too
		/*	mID := myUser.ID()
			jhbs.JHBase.Members()[mID-jhbs.MIDOffset].SetFirstName(fn)
			jhbs.JHBase.Members()[mID-jhbs.MIDOffset].SetLastName(ln)
			jhbs.JHBase.Members()[mID-jhbs.MIDOffset].SetMobile(mobile)*/

		fmt.Println("Updated particulars:", myUser)

		tpl.ExecuteTemplate(w, "memberprofileedit.gohtml", myUser)
		io.WriteString(w, `<div class="pop-up positive">Profile saved!</div>`)

	}

	tpl.ExecuteTemplate(w, "memberprofileedit.gohtml", myUser)

}

// memberCodeChange whenever member changes password
func MemberCodeChange(w http.ResponseWriter, req *http.Request) {

	myUser, err := askForMemberCookie(w, req)
	if err != nil {
		fmt.Println("Cookie error:", err)
		return
	}

	if req.Method == http.MethodPost {

		// validate old password (op)
		op, err := validateOldPassword(req)
		if err != nil {
			showErrorOnTop(w, "membercodechange.gohtml", myUser,
				err.Error())
			return
		}
		// validate new password (p)
		p, err6 := validatePassword(req) // requires "code" in <input>

		// from pwValidator.go
		isValidPassword := IsValidPassword(p)
		if err6 != nil {
			showErrorOnTop(w, "membercodechange.gohtml",
				myUser, `
				Invalid password<br/>
				Allowed characters for password: a-z, A-Z, 0-9 and *!@#$%^&(){}[]:;,.?/~_+-=|\<br/>
				`)
			return
		}
		if !isValidPassword {
			showErrorOnTop(w, "membercodechange.gohtml",
				myUser, `
				Password not strong enough<br/>
				Password requires min. 1 uppercase letter, 1 lowercase letter, 1 digit, 1 symbol, and be at least 8 characters long.<br/>
				Allowed characters for password: a-z, A-Z, 0-9 and *!@#$%^&(){}[]:;,.?/~_+-=|\<br/>
				`)
			return
		}

		// validate confirm new password (np)
		cp, err := validateConfirmNewPassword(req)
		if err != nil {
			showErrorOnTop(w, "membercodechange.gohtml", myUser,
				err.Error())
			return
		}

		// check if old password input matches old password in myUser
		isOldCode := CheckPasswordHash(op, myUser.Hash())
		if !isOldCode {
			showErrorOnTop(w, "membercodechange.gohtml", myUser,
				`Old password does not match
			`)
			return
		}

		// if user uses old password as new password
		isOldCode = CheckPasswordHash(p, myUser.Hash())
		if isOldCode {
			showErrorOnTop(w, "membercodechange.gohtml", myUser,
				`Your new password should not match your old password.
			`)
			return
		}

		// if user cannot confirm new password
		if p != cp {
			showErrorOnTop(w, "membercodechange.gohtml", myUser,
				`Your new passwords do not match
			`)
			return
		}

		// then hash new password and update member's hash
		hash, _ := HashPassword(p)

		myUser.SetHash(hash) // this also gets updated in jhbs.JHBase.Members()

		// tell user password has changed successfully
		showSuccessOnTop(w, "membercodechange.gohtml",
			myUser, `
				Password successfully changed.
				`)

	}

	tpl.ExecuteTemplate(w, "membercodechange.gohtml", myUser)

}

// memberLogout deletes the cookie from client's browser
func MemberLogout(w http.ResponseWriter, req *http.Request) {
	jhMemberCookie, err := req.Cookie("jhMemberCookie")
	if err != nil {
		showSuccessOnTop(w, "index.gohtml", nil,
			`You have logged out`)
		return
	}
	// delete from sessions because logout
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhMemberCookie.Value, &cCode); err == nil {
		memberSessions.Delete(cCode.String())
	}
	// delete cookie
	jhMemberCookie.MaxAge = -1
	http.SetCookie(w, jhMemberCookie)
	// redirect user to main menu
	io.WriteString(w, `
	<html>
		<meta http-equiv="refresh" content="2;url=/" />
		<body style="text-align: center; display: block;">
		<div class="redirect-message">
			<h2>You have logged out.</h2>
		</div>
		<link href="css/style.css" type="text/css" rel="stylesheet">
		</body>
	</html>
	`)
}
