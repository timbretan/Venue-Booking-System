package jhbshttp

import (
	"TimothyTAN_GoInAction1/jhbs"
	"errors"
	"net/http"
	"strconv"

	uuid "github.com/satori/go.uuid"
)

// askForMemberCookie asks for a cookie from client
// returns member if found
// otherwise redirects user to main page
func askForMemberCookie(w http.ResponseWriter, req *http.Request) (*(jhbs.Member), error) {

	//---try to locate cookie---
	jhMemberCookie, err := req.Cookie("jhMemberCookie")

	//---if no cookie, redirect---
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return &jhbs.Member{}, errors.New("No cookie called jhMemberCookie")
	}

	// read cookie handler
	var cCode uuid.UUID
	if err = s.Decode("jhMemberCookie", jhMemberCookie.Value, &cCode); err != nil {
		http.Redirect(w, req, "/invalid-cookie", http.StatusSeeOther)
		return &jhbs.Member{}, errors.New("Cannot decode cookie.")
	}

	//---get user's session---
	mli, ok := memberSessions.Load(cCode.String())

	//---if session not found, redirect to main menu---
	if !ok {
		http.Redirect(w, req, "/invalid-cookie", http.StatusSeeOther)
		return &jhbs.Member{}, errors.New("Map says not okay.")
	}

	//---retrieve the user info---
	myUser := jhbs.JHBase.Members()[mli.memberID-jhbs.MIDOffset]
	// fmt.Println("This user's first name is:", myUser.FirstName())

	if myUser.FirstName() == "" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return &jhbs.Member{}, errors.New("Member has no first name.")
	}
	return myUser, nil

}

// askForAdminCookie asks for an admin cookie from client
// returns admin if found
// otherwise redirects admin to main page
func askForAdminCookie(w http.ResponseWriter, req *http.Request) (*Admin, error) {

	//---try to locate cookie---
	jhAdminCookie, err := req.Cookie("jhAdminCookie")

	//---if no cookie, redirect---
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return &Admin{}, errors.New("No cookie called jhAdminCookie")
	}

	// read cookie handler
	var cCode uuid.UUID
	if err = s.Decode("jhAdminCookie", jhAdminCookie.Value, &cCode); err != nil {
		http.Redirect(w, req, "/invalid-cookie", http.StatusSeeOther)
		return &Admin{}, errors.New("Cannot decode cookie.")
	}

	//---get admin's session---
	adminLogin, ok := adminSessions.Load(cCode.String())

	//---if session not found, redirect to main menu---
	if !ok {
		http.Redirect(w, req, "/invalid-cookie", http.StatusSeeOther)
		return &Admin{}, errors.New("Map says not okay.")
	}

	//---retrieve the admin info---
	aID, err := strconv.Atoi(adminLogin.adminID[1:])
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return &Admin{}, errors.New("Invalid AdminID")
	}

	// remember to offset aID by 1
	myAdmin := jhbsAdmins.admins[aID]

	// admin has no first name? redirect.
	if myAdmin.FirstName == "" {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return &Admin{}, errors.New("Admin has no first name.")
	}
	return myAdmin, nil
}
