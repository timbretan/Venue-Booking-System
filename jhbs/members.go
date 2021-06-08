package jhbs

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// membership tiers
type MemberTier uint8

const (
	Bronze MemberTier = iota
	Silver
	Gold
	Diamond
)

// print out words of membership tier
func (t MemberTier) String() string {
	return [...]string{"Bronze", "Silver", "Gold", "Diamond"}[t]
}

// Member object
// unlike venues and bookings, Members are sorted by a numerical memberID
type Member struct {
	memberID  int // starts from the 6-digit 100000 (MIDOffset)
	firstName string
	lastName  string
	tier      MemberTier
	mobile    int  // mobile number (8-digit, follows SG)
	isActive  bool // false if member quits

	// To be added to the assignment
	userName  string
	hash      string    // password hash (never store passwords directly!)
	start     time.Time // signup time (incl date)
	lastLogin time.Time // last login time (incl date)

	// store pointers to bookings
	bookings *Bookings // booking history
}

func (m Member) String() string {
	var s string
	s += "Member "
	s += strconv.Itoa(m.memberID)
	s += ": "
	s += m.firstName
	s += " "
	s += m.lastName
	s += " ("
	s += m.userName
	s += ") ("
	s += fmt.Sprintf("%s", m.tier)
	s += " tier, "
	s += fmt.Sprintf("Mobile: %d)\n", m.mobile)
	return s
}

const MIDOffset = 100000

type Members []*Member

// slice of Members ready for concurrency
type Membership struct {
	sync.RWMutex
	members Members
}

func (members Members) String() string {
	var s string
	for _, m := range members {
		s += fmt.Sprint(m)
	}
	return s
}

var JHBase Membership // formerly members

// NewMember creates a new Member object for a new person, and appends it to the slice of members.members
// NB: It is ok for more than one member to have same name, but they cannot have same username.
// params: firstname (fn), lastname (ln), tier, mobile, username (u), hash, wg
func NewMember(fn string, ln string, tier MemberTier, mobile int, u string, hash string, start, lastLogin time.Time) (*Member, error) {

	if tier > 3 {
		return nil, errors.New("Invalid membership tier")
	}

	var err1, err2 error
	// human names shoud not have numbers

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		fn, err1 = cleanUpMemberNames(fn)
	}()
	go func() {
		defer wg.Done()
		ln, err2 = cleanUpMemberNames(ln)
	}()
	wg.Wait()

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	// for testing purposes, we are loading Members.csv which has no hash
	if hash == "" {
		// hash copied from wikipedia with some modifications
		hash = "$2a$10$N9lo8uLOibrgx2ZMRZoMyeIjZAgcfl7p92lgGxad68LJZdL17lhWy"
	}

	member := &Member{
		memberID:  0, // 0 is a placeholder for a memberID to be given anew based on position in Members slice
		firstName: fn,
		lastName:  ln,
		tier:      tier,
		mobile:    mobile,
		isActive:  true,
		userName:  u,
		hash:      hash,
		//start:     time.Now().Add(-30 * 24 * time.Hour),
		start: start,
		// lastLogin: time.Now().Add(-29 * 24 * time.Hour),
		lastLogin: lastLogin,
		bookings:  MakeBookings(),
	}

	JHBase.Append(member) // inside already has rwmutex and change memberID
	return member, nil
}

// Appends an item to the member slice
func (ms *Membership) Append(m *Member) {
	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	// addition of member to the members slice is lumped as a critical section
	ms.Lock()
	defer ms.Unlock()

	(*m).memberID = len(ms.members) + MIDOffset
	ms.members = append(ms.members, m)
}

func NotifyMember(mID int) string {
	// in case I forgot to decrement memberID for some sort criteria
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error 1")
		}
	}()

	poc := JHBase.members[mID-MIDOffset]
	str := fmt.Sprintf("- Member %d %s %s notified via its mobile %d", mID, poc.firstName, poc.lastName, poc.mobile)
	fmt.Println(str)
	return str
}

// gets member ID from a member
func (member *Member) ID() int {
	return member.memberID
}

// gets first name from a member
func (member *Member) FirstName() string {
	return member.firstName
}

// gets last name from a member
func (member *Member) LastName() string {
	return member.lastName
}

// gets membership tier from a member
func (member *Member) Tier() MemberTier {
	return member.tier
}

// gets mobile from a member
func (member *Member) Mobile() int {
	return member.mobile
}

// gets hash of a member
func (member *Member) UserName() string {
	return member.userName
}

// gets hash of a member
func (member *Member) Hash() string {
	return member.hash
}

// gets signup time (start) of a member
func (member *Member) Start() time.Time {
	return member.start
}

// gets last login time of a member
func (member *Member) LastLogin() time.Time {
	return member.lastLogin
}

// gets booking history from member
func (member *Member) Bookings() bookings {
	return member.bookings.bs
}

// sets member ID from a member
func (member *Member) SetID(id int) {
	member.memberID = id
}

// sets first name from a member
func (member *Member) SetFirstName(fn string) {
	member.firstName = fn
}

// sets last name from a member
func (member *Member) SetLastName(fn string) {
	member.lastName = fn
}

// sets membership tier from a member
func (member *Member) SetTier(tier MemberTier) {
	member.tier = tier
}

// sets mobile from a member
func (member *Member) SetMobile(mobile int) {
	member.mobile = mobile
}

// sets hash of a member
func (member *Member) SetUserName(u string) {
	member.userName = u
}

// sets hash of a member
func (member *Member) SetHash(hash string) {
	member.hash = hash
}

// sets hash of a member
func (member *Member) SetStart(start time.Time) {
	member.start = start
}

// sets hash of a member
func (member *Member) SetLastLogin(lastLogin time.Time) {
	member.lastLogin = lastLogin
}

// gets members slice within membership struct
func (ms *Membership) Members() Members {
	return ms.members
}
