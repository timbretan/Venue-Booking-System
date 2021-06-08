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

	securecookie "github.com/gorilla/securecookie"
	uuid "github.com/satori/go.uuid"
)

// Member object
// unlike venues and bookings, Members are sorted by a numerical memberID
type Admin struct {
	AdminID   string // A000, A001, A002...
	FirstName string
	LastName  string

	// To be added to the assignment
	UserName  string
	Hash      string // password hash (never store passwords directly!)
	Start     time.Time
	LastLogin time.Time
}

type Admins []*Admin

type Adminship struct {
	mu     sync.RWMutex
	admins Admins
}

var jhbsAdmins Adminship

// muxAdmin.go - the servemux for admin

// adminGate provides the login form for admins
// gate here refers to a security gate
func AdminGate(w http.ResponseWriter, req *http.Request) {
	// define admin cookie
	var t *securecookie.SecureCookie
	t = securecookie.New(hashKey, blockKey)

	jhAdminCookie, err := req.Cookie("jhAdminCookie")
	// if got cookie already, that means admin already logged in
	if err == nil {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	}

	if err != nil {
		// set cookie handler
		// id := uuid.NewV4() // some use uuid for the cookie's value
		// for June Holidays booking system
		if encoded, err := t.Encode("jhAdminCookie", time.Now().UnixNano()); err == nil {
			jhAdminCookie = &http.Cookie{
				Name:    "jhAdminCookie",
				Value:   encoded,
				Expires: time.Now().Add(5 * time.Minute), // expires in 5 minutes
				// Secure:   true,
				HttpOnly: true,
			}
			// http.SetCookie(w, jhAdminCookie)
		}
	}
	var admin Admin

	//---try to locate the admin based on cookie's value---
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhAdminCookie.Value, &cCode); err == nil {
		if adminLogin, ok := adminSessions.Load(cCode.String()); ok {
			aID, err := strconv.Atoi(adminLogin.adminID[1:])
			if err != nil {
				fmt.Println("Invalid admin ID at AdminGate")
				return
			}
			admin = *(jhbsAdmins.admins)[aID]
		}
	}
	// CAUTION: if error abv, no need to re-direct back to index.gohtml!
	// this is just the gate to the admin-only area, not the admin-only area

	// parse template file (.gohtml or other extension)
	// executes template from index.gohtml
	err = tpl.ExecuteTemplate(w, "admingate.gohtml", admin)
	if err != nil {
		log.Fatalln(err)
	}
}

// memberLogin is when members login
// in adminlogin.gohtml (called by admingate.gohtml),
// there is <form method="POST" action="/jh-admin-login">
// and /jh-admin-login calls adminLogin
func AdminLogin(w http.ResponseWriter, req *http.Request) {

	var admin *Admin

	/*
		// in case random ppl land up on this page
		// ask for admin cookie first
		admin, err := askForAdminCookie(w, req)
		if err != nil {

			return
		}
	*/

	// validate username (u) and code (p)
	u, err1 := validateString(req, "username")
	if err1 != nil {
		showErrorOnTop(w, "admingate.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// points admin to a member

	for _, adminFound := range jhbsAdmins.admins {
		if adminFound.UserName == u {
			admin = adminFound
			break
		}
	}
	// if user not found
	if admin == nil {
		showErrorOnTop(w, "admingate.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// validate new password (p)
	p, err2 := validatePassword(req) // requires "code" in <input>
	if err2 != nil {
		showErrorOnTop(w, "admingate.gohtml", nil,
			`Invalid username and/or password`)
		return
	}
	// check is pw is correct
	isCorrectCode := CheckPasswordHash(p, admin.Hash)
	if !isCorrectCode {
		showErrorOnTop(w, "admingate.gohtml", nil,
			`Invalid username and/or password`)
		return
	}

	// Create cookie
	jhAdminCookie, err := req.Cookie("jhAdminCookie")
	loginTime := time.Now()

	// if don't have cookie, create one
	if err != nil {
		// create an admin cookie
		var t *securecookie.SecureCookie
		t = securecookie.New(hashKey, blockKey)
		cCode := uuid.NewV4()
		if encoded, err := t.Encode("jhAdminCookie", cCode); err == nil {
			jhAdminCookie = &http.Cookie{
				Name:    "jhAdminCookie",
				Value:   encoded,
				Expires: loginTime.Add(24 * time.Hour), // expires in 1 day
				// Secure:   true,
				// HttpOnly: true,
			}
			// compose an adminLoginInfo struct,
			// then add it to adminSessions, with the cookie value mapped to the struct
			adminSessions.Store(cCode.String(), &adminLoginInfo{
				admin.AdminID,
				req.RemoteAddr, // more reliable to use remote addr
				loginTime,
			})
			http.SetCookie(w, jhAdminCookie)
		}
	}
	admin.LastLogin = loginTime
	// then redirect user to member's page after login
	io.WriteString(w, `
	<html>
		<meta http-equiv="refresh" content="2;url=/jh-admin" />
		<body style="text-align: center; display: block;">
		<div class="redirect-message">
			<h2>Welcome `+admin.FirstName+` `+admin.LastName+`!</h2>
		</div>
		<link href="css/style.css" type="text/css" rel="stylesheet">
		</body>
	</html>
	`)
}

// membersOnly is only intended for members
// it is the first page members would land upon logging in or signing up
func AdminOnly(w http.ResponseWriter, req *http.Request) {

	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// based on former jhbs.PrepareBookVenue()
	// happens if user searches for a venue
	if req.Method == http.MethodPost {
		adminGetVenueSearch(w, req, admin)
		return
	}

	tpl.ExecuteTemplate(w, "adminmain.gohtml", admin)
}

// internal function
// showVenueSearch only applies to when user searches for a venue
// happens at /jh-member and /jh-member/venue/..
// done only as POST
func adminGetVenueSearch(w http.ResponseWriter, req *http.Request, admin *Admin) {
	searchVenue, err := validateVenue(req, "venue")

	// ask user to type smth to search venue
	if err != nil {
		showErrorOnTop(w, "adminmain.gohtml", admin,
			err.Error())
		return
	}

	// find the venue
	searchVenue = strings.Title(searchVenue)
	gotVenue, out := jhbs.FindVenue(searchVenue)

	// data for use in template
	// fields must be exported (Field)!
	type Data struct {
		Admin    *Admin
		GotVenue bool
		Venues   []string
	}
	var data = &Data{admin, gotVenue, out}
	tpl.ExecuteTemplate(w, "adminmain.gohtml", data)
	return
}

// AdminProcessBooking processes booking of a venue
// Gets called by /jh-admin/process/<venue name>
// and uses adminprocessbooking.gohtml
func AdminPrepareProcessBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// get venue string from URL path
	v := req.URL.Path[len("/jh-admin/process/"):]

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
		Admin *Admin
		Venue *jhbs.Venue // get waitlist from this venue
	}
	var data = &Data{admin, theVenue}

	tpl.ExecuteTemplate(w, "adminprocessbooking.gohtml", data)

}

func AdminProcessBooking(w http.ResponseWriter, req *http.Request) {

	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// TODO: No need Bookings in data struct;
	// Use Venue.Waitlist -> Venue.ApprovedBookings i.p.v. Bookings field

	// data for use in template
	// fields must be exported (Field)!
	type Data struct {
		Admin    *Admin
		Venue    *jhbs.Venue    // get waitlist from this venue
		Bookings *jhbs.Bookings // while processing bookings, put booking in bookings, so that you can show their status after processing
	}
	var data = &Data{admin, nil, nil}

	if req.Method == http.MethodPost {

		var theVenue *jhbs.Venue
		// get newvenue field if any
		// prevents scripts from loading on newvenue field
		v, err := validateVenue(req, "venue")

		if err == nil {
			// show the venue but not its availability yet
			theVenue, _, err = jhbs.ShowVenue(v)
			// if for some reason venue not found, throw err
			if err != nil {
				showErrorOnTop(w, "adminmain.gohtml", nil,
					`Invalid search`)
				return
			}
		} else {
			// empty venue or improper venue chosen (or scripts lol)
			showErrorOnTop(w, "adminmain.gohtml", admin,
				`Invalid search`)
			return
		}

		// validate intent to cancel
		processBooking, err := validateString(req, "process")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// actual work of cancelling booking
		switch processBooking {
		case "yes":

			data.Venue = theVenue

			/*
				// retrieve all bookings in waitlist
				var theBookingsSlice []*jhbs.Booking
				var ch = make(chan *jhbs.Booking)
				var wg sync.WaitGroup
				wg.Add(1)
				go func() {
					defer wg.Done()
					// if waitlist is nil
					defer func() {
						if err := recover(); err != nil {
							fmt.Println(err)
							theBookingsSlice = nil
						}
					}()
					theVenue.WaitlistBookings().DumpNodes2(ch)
				}()
				for b := range ch {
					theBookingsSlice = append(theBookingsSlice, b)
				}
				theBookings := &jhbs.Bookings{}
				theBookings.Append(theBookingsSlice...)
				wg.Wait()
				// after ch is closed, add to Bookings of temp data struct
				data.Bookings = theBookings
			*/
			data.Bookings = theVenue.Waitlist()

			// then process bookings (bookings get updated in status)
			err = theVenue.ProcessBookings()
			if err != nil {
				fmt.Println(err)
				return
			}

			// then retrieve venue availability to get updated bookings
			theVenue, _, err = jhbs.ShowVenue(v)
			data.Venue = theVenue

		default:
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
	}

	tpl.ExecuteTemplate(w, "adminprocessbooking.gohtml", data)
}

// AdminSearchBooking is when admin wants to search for a booking
// Usually in order to edit the booking
func AdminSearchBooking(w http.ResponseWriter, req *http.Request) {

	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin   *Admin
		Booking *jhbs.Booking
	}
	var data = &Data{admin, nil}

	if req.Method == http.MethodPost {
		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", data,
				`No booking with that bookingID found`)
			return
		}
		data.Booking = theBooking
	}

	tpl.ExecuteTemplate(w, "adminsearchbooking.gohtml", data)
}

// AdminPrepareEditBooking resembles MemberPrepareEditBooking
func AdminPrepareEditBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		var theVenue *jhbs.Venue
		// get newvenue field if any
		// prevents scripts from loading on newvenue field
		v, err := validateVenue(req, "newvenue")

		if err == nil {
			// show the venue and its availability
			theVenue, _, err = jhbs.ShowVenue(v)
			// if for some reason venue not found, throw err
			if err != nil {
				showErrorOnTop(w, "adminsearchbooking.gohtml", nil,
					`Invalid search`)
				return
			}
		} else if err.Error() == "Empty newvenue" {
			// nothing on newvenue field yet
			theVenue = nil
		} else {
			// improper venue chosen (or scripts lol)
			showErrorOnTop(w, "adminmain.gohtml", admin,
				`Invalid search`)
			return
		}

		// get all venues from jhbs
		allVenues := jhbs.MakeVenueSortSlice()

		type Data struct {
			Admin     *Admin        // from cookie
			Booking   *jhbs.Booking // must-have
			AllVenues *jhbs.VenueSortSlice
			NewVenue  *jhbs.Venue
		}
		data := &Data{admin, theBooking, allVenues, theVenue}
		tpl.ExecuteTemplate(w, "admineditbooking.gohtml", data)

	}
}

func AdminEditBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		bID, err := validateString(req, "bookingid")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}
		v, err := validateVenue(req, "venue")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}
		sd, err := validateDay(req, "startday")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}
		sh, err := validateHour(req, "starthour")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}
		ed, err := validateDay(req, "endday")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}
		eh, err := validateHour(req, "endhour")
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}

		// else edit the booking
		theBooking, notification, err := jhbs.AdminDoEditBooking(bID, v, sd, sh, ed, eh)
		if err != nil {
			showErrorOnTop(w, "adminsearchbooking.gohtml", admin,
				err.Error())
			return
		}

		io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="5;url=/jh-admin-search-booking" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					<h2>Booking `+bID+` has been edited to become:</h2>
					<p>`+theBooking.String()+`</p>
					<p>`+notification+`</p>
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
		`)
	}
}

// AdminPrepareCancelBooking resembles MemberPrepareCancelBooking
func AdminPrepareCancelBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached AdminPrepareCancelBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))

		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		type Data struct {
			Admin   *Admin
			Booking *jhbs.Booking
		}
		data := &Data{admin, theBooking}
		tpl.ExecuteTemplate(w, "admincancelbooking.gohtml", data)

	}
}

// AdminCancelBooking - when it actually cancels booking
func AdminCancelBooking(w http.ResponseWriter, req *http.Request) {
	_, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached AdminCancelBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))
		fmt.Println(req.FormValue("cancel"))

		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// validate intent to cancel
		cancelBooking, err := validateString(req, "cancel")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// actual work of cancelling booking
		switch cancelBooking {
		case "yes":
			err := jhbs.AdminDoCancelBooking(bID)
			var stringToWrite string
			if err == nil {
				stringToWrite = `<h2>You have cancelled booking ` + bID + `.</h2>`
			} else {
				stringToWrite = `<h2>` + err.Error() + `</h2>`
			}
			io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="2;url=/jh-admin-see-booking?memberid=`+strconv.Itoa(theBooking.MemberID())+`" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					`+stringToWrite+`
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
			`)
		default:
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
	}
}

// AdminPrepareRejectBooking resembles AdminPrepareCancelBooking but for rejection
func AdminPrepareRejectBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached AdminPrepareRejectBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))

		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		type Data struct {
			Admin   *Admin
			Booking *jhbs.Booking
		}
		data := &Data{admin, theBooking}
		tpl.ExecuteTemplate(w, "adminrejectbooking.gohtml", data)

	}
}

// AdminCancelBooking - when it actually cancels booking
func AdminRejectBooking(w http.ResponseWriter, req *http.Request) {
	_, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// if user randomly types this URL, send it back to /jh-member
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		fmt.Print("Reached AdminRejectBooking via POST: ")
		fmt.Println(req.FormValue("bookingid"))
		fmt.Println(req.FormValue("cancel"))

		// validate bookingID
		bID, err := validateString(req, "bookingid")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theBooking, err := jhbs.AdminFindBooking(bID)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// validate intent to reject
		rejectBooking, err := validateString(req, "reject")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// actual work of cancelling booking
		switch rejectBooking {
		case "yes":
			err := jhbs.AdminDoRejectBooking(bID)
			var stringToWrite string
			if err == nil {
				stringToWrite = `<h2>You have rejected booking ` + bID + `.</h2>`
			} else {
				stringToWrite = `<h2>` + err.Error() + `</h2>`
			}
			io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="2;url=/jh-admin-see-booking?memberid=`+strconv.Itoa(theBooking.MemberID())+`" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					`+stringToWrite+`
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
			`)
		default:
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
	}
}

// AdminViewSessions allows admin to view
func AdminViewSessions(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin             *Admin // from cookie
		AllAdminSessions  *AdminSessions
		AllMemberSessions *MemberSessions
		ReqLocation       string
	}
	data := &Data{admin, adminSessions, memberSessions, req.RemoteAddr}
	tpl.ExecuteTemplate(w, "adminviewsession.gohtml", data)

}

// AdminDeleteSession deletes a session
func AdminDeleteSession(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	if req.Method == http.MethodPost {
		group, err := validateString(req, "group")
		if err != nil {
			showErrorOnTop(w, "adminmain.gohtml", admin,
				err.Error())
			return
		}
		sessionID, err := validateSessionID(req)
		if err != nil {
			showErrorOnTop(w, "adminmain.gohtml", admin,
				err.Error())
			return
		}
		switch group {
		case "a": // do the deletion for admin sessions
			adminSessions.Delete(sessionID)

		case "b": // do the deletion for member sessions
			memberSessions.Delete(sessionID)

		default:
			showErrorOnTop(w, "adminmain.gohtml", admin,
				`Illegal operation.`)
			return
		}
		// redirect admin to adminviewsession
		io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="2;url=/jh-admin-view-session" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					<h2>Session deleted.</h2>
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
		`)
	}

}

// AdminSearchMember allows an admin to search for a member using memberID
func AdminSearchMember(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin  *Admin       // from cookie
		Member *jhbs.Member // to a member
	}
	data := &Data{admin, nil}

	if req.Method == http.MethodPost {
		// validate memberID
		mID, err := validateMemberID(req)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theMember, err := jhbs.AdminFindMember(mID)
		if err != nil {
			showErrorOnTop(w, "adminsearchmember.gohtml", data,
				`No member with that memberID found`)
			return
		}
		data.Member = theMember
	}

	tpl.ExecuteTemplate(w, "adminsearchmember.gohtml", data)

}

func AdminSeeMemberBooking(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin  *Admin       // from cookie
		Member *jhbs.Member // to a member
	}
	data := &Data{admin, nil}

	// if GET
	if req.Method == http.MethodGet {

		q := req.URL.Query()
		// get first member of memberid
		mIDStr := q["memberid"][0]

		// if no query string, don't check memberID
		if mIDStr != "" {
			mIDStr = template.HTMLEscapeString(mIDStr)

			// if field is empty
			// convert string to int
			mID, err := strconv.Atoi(mIDStr)

			// if field has non-English chars
			if err != nil || mID < 100000 || mID > 999999 {
				http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
				return
			}

			theMember, err := jhbs.AdminFindMember(mID)
			// no member with such ID found? Go back to prev search result.
			if err != nil {
				http.Redirect(w, req, req.Header.Get("Referer"), http.StatusSeeOther)
				return
			}

			data.Member = theMember
		}
	}

	// if POST
	if req.Method == http.MethodPost {
		// validate memberID
		mID, err := validateMemberID(req)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theMember, err := jhbs.AdminFindMember(mID)
		if err != nil {
			showErrorOnTop(w, "adminsearchmember.gohtml", data,
				`No member with that memberID found`)
			return
		}

		data.Member = theMember
	}

	tpl.ExecuteTemplate(w, "adminseememberbooking.gohtml", data)

}

// AdminDeleteMember allows an admin to delete a member
func AdminPrepareDeleteMember(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin  *Admin       // from cookie
		Member *jhbs.Member // to a member
	}
	data := &Data{admin, nil}

	// if user randomly types this URL, send it back to /jh-admin
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		mID, err := validateMemberID(req)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		theMember, err := jhbs.AdminFindMember(mID)
		if err != nil {
			showErrorOnTop(w, "adminsearchmember.gohtml", data,
				`No member with that memberID found`)
			return
		}

		data.Member = theMember
		tpl.ExecuteTemplate(w, "admindeletemember.gohtml", data)
	}
}

// AdminDeleteMember allows an admin to delete a member
func AdminDeleteMember(w http.ResponseWriter, req *http.Request) {
	admin, err := askForAdminCookie(w, req)
	if err != nil {
		fmt.Println(err)
		return
	}

	type Data struct {
		Admin  *Admin       // from cookie
		Member *jhbs.Member // to a member
	}
	data := &Data{admin, nil}

	// if user randomly types this URL, send it back to /jh-admin
	if req.Method != http.MethodPost {
		http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
		return
	} else {
		mID, err := validateMemberID(req)
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}
		_, err = jhbs.AdminFindMember(mID)
		if err != nil {
			showErrorOnTop(w, "adminsearchmember.gohtml", data,
				`No member with that memberID found`)
			return
		}

		// validate intent to cancel
		deleteMember, err := validateString(req, "delete")
		if err != nil {
			http.Redirect(w, req, "/jh-admin", http.StatusSeeOther)
			return
		}

		// actual work of cancelling booking
		switch deleteMember {
		case "yes":
			jhbs.AdminDoDeleteMember(mID) // TODO
			io.WriteString(w, `
			<html>
				<meta http-equiv="refresh" content="2;url=/jh-admin-search-member" />
				<body style="text-align: center; display: block;">
				<div class="redirect-message">
					<h2>Member `+strconv.Itoa(mID)+` has been deleted.</h2>
				</div>
				<link href="css/style.css" type="text/css" rel="stylesheet">
				</body>
			</html>
			`)
		default:
			http.Redirect(w, req, "/jh-admin-search-member", http.StatusSeeOther)
			return
		}
	}
}

// memberLogout deletes the cookie from client's browser
func AdminLogout(w http.ResponseWriter, req *http.Request) {
	jhAdminCookie, err := req.Cookie("jhAdminCookie")
	if err != nil {
		showSuccessOnTop(w, "index.gohtml", nil,
			`You have logged out`)
		return
	}
	// delete from sessions because logout
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhAdminCookie.Value, &cCode); err == nil {
		adminSessions.Delete(cCode.String())
	}

	jhAdminCookie.MaxAge = -1 // delete cookie
	http.SetCookie(w, jhAdminCookie)
	// redirect user to main menu
	io.WriteString(w, `
	<html>
		<meta http-equiv="refresh" content="2;url=/" />
		<body style="text-align: center; display: block;">
		<div class="redirect-message">
			<h2>You have logged out as admin.</h2>
		</div>
		<link href="css/style.css" type="text/css" rel="stylesheet">
		</body>
	</html>
	`)
}
