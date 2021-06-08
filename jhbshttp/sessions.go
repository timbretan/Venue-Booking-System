package jhbshttp

import (
	"TimothyTAN_GoInAction1/jhbs"
	"strconv"
	"sync"
	"time"
)

// sessions.go defines how the cookie value is linked to a struct
// this struct contains:
// for members: memberID (int), their RemoteAddr (string), last login (time.Time)
// for admins: adminID (string), their RemoteAddr (string), last login (time.Time)

type memberLoginInfo struct {
	memberID  int
	location  string    // remote IP addr
	lastLogin time.Time // last login time
}

type adminLoginInfo struct {
	adminID   string
	location  string    // remote IP addr
	lastLogin time.Time // last login time
}

type LoginInfo interface {
	Location() string
	String() string
}

func (mli *memberLoginInfo) Location() string {
	return mli.location
}

func (ali *adminLoginInfo) Location() string {
	return ali.location
}

func (mli *memberLoginInfo) String() string {
	// get member
	m := jhbs.JHBase.Members()[mli.memberID-jhbs.MIDOffset]
	// make string
	s := "MemberID: " + strconv.Itoa(mli.memberID) + "\n"
	s += "First Name: " + m.FirstName() + "\n"
	s += "Last Name: " + m.LastName() + "\n"
	s += "UserName: " + m.UserName() + "\n"
	s += "IP Addr: " + mli.location + "\n"
	s += "Last login: "
	s += mli.lastLogin.Format("02-01-2006 (Mon) 15:04:05 UTC -0700")
	return s
}

func (ali *adminLoginInfo) String() string {
	var s string
	s = "AdminID: " + ali.adminID + "\n"
	s += "IP Addr: " + ali.location + "\n"
	s += "Last login: "
	s += ali.lastLogin.Format("02-01-2006 (Mon) 15:04:05 UTC -0700")
	return s
}

//---mapSessions is used for saving sessions ---
// cookie's value is encoded in AES-256
var memberSessions = NewMemberSessions() // for members: map cookie's value to memberID, ip location and last login
var adminSessions = NewAdminSessions()   // for admins: map cookie's value to adminID, ip location and last login

// maps are not concurrency-safe
// include mutex in one of them
// could have used sync.Map but I anticipate
// a lot of reads and writes, which don't fit
// what sync.Map is intended (bursty writes with many many reads)
type MemberSessions struct {
	sync.RWMutex
	internal map[string]*memberLoginInfo // for members: map uuID (to be encoded and decoded in cookie) to memberID, ip location and last login
}

type AdminSessions struct {
	sync.RWMutex
	internal map[string]*adminLoginInfo // for admins: map uuID (to be encoded and decoded in cookie) to adminID, ip location and last login
}

type ConcurrentMap interface {
	Load(key interface{}) (interface{}, bool)
	LoadAll() *map[string]*interface{}
	Delete(key interface{})
	Store(key interface{})
}

// MemberSessions methods
func NewMemberSessions() *MemberSessions {
	return &MemberSessions{
		internal: make(map[string]*memberLoginInfo),
	}
}

func (rm *MemberSessions) Load(key string) (*memberLoginInfo, bool) {
	rm.RLock()
	value, ok := rm.internal[key]
	rm.RUnlock()
	return value, ok
}

func (rm *MemberSessions) LoadAll() *map[string]*memberLoginInfo {
	rm.RLock()
	value := &rm.internal
	rm.RUnlock()
	return value
}

func (rm *MemberSessions) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *MemberSessions) Store(key string, value *memberLoginInfo) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}

// AdminSessions methods
func NewAdminSessions() *AdminSessions {
	return &AdminSessions{
		internal: make(map[string]*adminLoginInfo),
	}
}

func (rm *AdminSessions) Load(key string) (*adminLoginInfo, bool) {
	rm.RLock()
	value, ok := rm.internal[key]
	rm.RUnlock()
	return value, ok
}

func (rm *AdminSessions) LoadAll() *map[string]*adminLoginInfo {
	rm.RLock()
	value := &rm.internal
	rm.RUnlock()
	return value
}

func (rm *AdminSessions) Delete(key string) {
	rm.Lock()
	delete(rm.internal, key)
	rm.Unlock()
}

func (rm *AdminSessions) Store(key string, value *adminLoginInfo) {
	rm.Lock()
	rm.internal[key] = value
	rm.Unlock()
}
