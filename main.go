package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	jhbs "TimothyTAN_GoInAction1/jhbs"
	jhbshttp "TimothyTAN_GoInAction1/jhbshttp"
)

func init() {
	// set random seed
	rand.Seed(time.Now().Unix())

	var wg sync.WaitGroup
	wg.Add(2)
	// load files for JHBS
	go func() {
		defer wg.Done()
		jhbs.PrepareJHBS()
	}()
	// load files for JHBSHTTP
	go func() {
		defer wg.Done()
		jhbshttp.PrepareJHBSHTTP()
	}()
	wg.Wait()

}

func main() {
	// DEBUG: print bookings of the 1st member
	// fmt.Print(jhbs.JHBase.Members()[2].Bookings())

	http.HandleFunc("/", jhbshttp.MainMenu) // main menu for users
	http.HandleFunc("/css/", jhbshttp.ServeResource)

	// signups
	http.HandleFunc("/jh-signup", jhbshttp.Signup)     // signup page for new members
	http.HandleFunc("/jh-signedup", jhbshttp.Signedup) // signed-up page for new members, redirecting to members-only area (/jh-member)

	// member login
	http.HandleFunc("/jh-member-login", jhbshttp.MemberLogin)

	// members-only area
	// here they can search for a venue
	http.HandleFunc("/jh-member", jhbshttp.MemberOnly)              // members-only area
	http.HandleFunc("/jh-member/venue/", jhbshttp.MemberShownVenue) // when member is shown venue

	// under "BROWSE": browse for a venue
	// caution: No idea why "jh-member-browse" does not work
	// but "jh-member-brow" and "jh-member-browse-venue" works
	http.HandleFunc("/jh-member-browse-venue", jhbshttp.MemberBrowseVenue) // when member browses all the venues

	// under "HISTORY": view booking history, edit and cancel booking
	http.HandleFunc("/jh-member-history", jhbshttp.MemberHistory)                     // member views own booking history
	http.HandleFunc("/jh-member-edit-booking", jhbshttp.MemberPrepareEditBooking)     // when member edits booking
	http.HandleFunc("/jh-member-edit-booking-2", jhbshttp.MemberEditBooking)          // when member edits booking
	http.HandleFunc("/jh-member-cancel-booking", jhbshttp.MemberPrepareCancelBooking) // when member cancels booking
	http.HandleFunc("/jh-member-cancel-booking-2", jhbshttp.MemberCancelBooking)      // when member cancels booking

	// under "PROFILE": view profile, edit profile and code (password)
	http.HandleFunc("/jh-member-profile", jhbshttp.MemberProfile)          // member's profile, used for editing own member info
	http.HandleFunc("/jh-member-profile-edit", jhbshttp.MemberProfileEdit) // member's profile, used for editing own member info
	http.HandleFunc("/jh-member-code-change", jhbshttp.MemberCodeChange)   // when member want to change password

	// under "LOGOUT"
	http.HandleFunc("/jh-member-logout", jhbshttp.MemberLogout) // when member logout

	// admin login uses an unintuitive login URL
	// much like how some WordPress sites have custom URLs for wp-admin login
	http.HandleFunc("/jh-standby", jhbshttp.AdminGate)      // webpage for admin to log in
	http.HandleFunc("/jh-admin-login", jhbshttp.AdminLogin) // when admin logins

	// admin-only area
	// here it can search for venue and process bookings to it
	http.HandleFunc("/jh-admin", jhbshttp.AdminOnly)                           // admin-only area
	http.HandleFunc("/jh-admin/process/", jhbshttp.AdminPrepareProcessBooking) // when admin processes bookings to a venue
	http.HandleFunc("/jh-admin/process-2", jhbshttp.AdminProcessBooking)       // when admin processes bookings to a venue

	// under "BOOKINGS": search, edit, reject or cancel booking
	http.HandleFunc("/jh-admin-search-booking", jhbshttp.AdminSearchBooking)        // when admin searches for a booking
	http.HandleFunc("/jh-admin-edit-booking", jhbshttp.AdminPrepareEditBooking)     // when admin edits booking
	http.HandleFunc("/jh-admin-edit-booking-2", jhbshttp.AdminEditBooking)          // when admin edits booking
	http.HandleFunc("/jh-admin-reject-booking", jhbshttp.AdminPrepareRejectBooking) // when admin rejects booking
	http.HandleFunc("/jh-admin-reject-booking-2", jhbshttp.AdminRejectBooking)      // when admin rejects booking
	http.HandleFunc("/jh-admin-cancel-booking", jhbshttp.AdminPrepareCancelBooking) // when admin cancels booking
	http.HandleFunc("/jh-admin-cancel-booking-2", jhbshttp.AdminCancelBooking)      // when admin cancels booking

	// under "SESSIONS": view and delete sessions
	http.HandleFunc("/jh-admin-view-session", jhbshttp.AdminViewSessions)    // when admin views login sessions of admins and members
	http.HandleFunc("/jh-admin-delete-session", jhbshttp.AdminDeleteSession) // when admin deletes a login session of either admin or member

	// under "MEMBERS": search member, see its bookings and delete member
	// while seeing its bookings, you can also edit, reject or cancel booking
	http.HandleFunc("/jh-admin-search-member", jhbshttp.AdminSearchMember)        // when admin searches for a member
	http.HandleFunc("/jh-admin-see-booking", jhbshttp.AdminSeeMemberBooking)      // when admin sees a member's bookings
	http.HandleFunc("/jh-admin-delete-member", jhbshttp.AdminPrepareDeleteMember) // when admin wants to delete member
	http.HandleFunc("/jh-admin-delete-member-2", jhbshttp.AdminDeleteMember)      // when admin deletes member

	// under "LOGOUT"
	http.HandleFunc("/jh-admin-logout", jhbshttp.AdminLogout)

	// in case of invalid cookies
	// i.e. client has a cookie,
	// but session does not recognise this cookie's value
	// handlers defined at muxOther.go
	http.HandleFunc("/invalid-cookie", jhbshttp.InvalidCookie)

	// DEBUGGER
	// go func() {
	// 	for {
	// 		select {
	// 		case <-time.Tick(1 * time.Second):
	// 			if len(members.members) >= 1 {
	// 				fmt.Printf("Here are the member details: %v\n", members.members[0])
	// 			}
	// 		}
	// 	}
	// }()

	// starts server on port 5221
	http.ListenAndServe(":5221", nil)

}
