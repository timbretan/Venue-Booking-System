package jhbshttp

import (
	"TimothyTAN_GoInAction1/jhbs"
	crypt "crypto/rand"
	"log"
	"reflect"
	"strings"
	"text/template"
	"time"

	securecookie "github.com/gorilla/securecookie"
)

// templates
var tpl *template.Template

// TODO: Should this cookie be global?
var s *securecookie.SecureCookie // s is a SecureCookie instance
// for encoding the cookie
var hashKey, blockKey []byte

// see if that field is available on that struct passed to the template
// e.g. does Data struct contain a field called "Member"?
func avail(field string, data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	return v.FieldByName(field).IsValid()
}

// makeLink replaces spaces with underscores
func makeLink(field string) string {
	return strings.Replace(field, " ", "_", -1)
}

// br replaces "\n" with <br/>
func br(field string) string {
	return strings.Replace(field, "\n", `<br/>`, -1)
}

// ulli converts string to an unordered list based on line breaks
func ulli(field string) string {
	return `<ul><li>` + strings.Replace(strings.TrimRight(field, "\n"), "\n", `</li><li>`, -1) + `</li></ul>`
}

// olli converts string to an ordered list based on line breaks
func olli(field string) string {
	return `<ol><li>` + strings.Replace(strings.TrimRight(field, "\n"), "\n", `</li><li>`, -1) + `</li></ol>`
}

// checks if a booking is cancelled
func isNotCancelled(bs jhbs.BookingStatus) bool {
	return bs.String() != "Cancelled"
}

// checks if a booking is pending
// used normally for rejecting bookings
// (You can only reject pending bookings)
func isPending(bs jhbs.BookingStatus) bool {
	return bs.String() == "Pending"
}

func PrepareJHBSHTTP() {
	// read templates, and also tell the templates what the fns refer to
	tpl = template.Must(template.New("").Funcs(template.FuncMap{
		"avail":          avail,          // does the struct field exist?
		"makelink":       makeLink,       // converts spaces to "_" for links
		"br":             br,             // converts `\n` to `<br/>`
		"ulli":           ulli,           // changes string to fit into an unordered list
		"olli":           olli,           // changes string to fit into an ordered list
		"isnotcancelled": isNotCancelled, // is the booking status not "Cancelled"?
		"ispending":      isPending,      // is the booking status "Pending"?
	}).ParseGlob("templates/*")) // Must() reads the templates

	// create hash keys for cookies
	// Hash keys should be at least 32 bytes long
	hashKey := make([]byte, 32)
	_, err := crypt.Read(hashKey) // crypt is crypto/rand
	if err != nil {
		log.Fatalln("Unable to generate hash key for securing cookies.")
	}

	// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
	// Shorter keys may weaken the encryption used.
	blockKey = make([]byte, 16)

	_, err = crypt.Read(blockKey) // crypt is crypto/rand
	if err != nil {
		log.Fatalln("Unable to generate block key for securing cookies.")
	}

	// make a secure cookie instance
	s = securecookie.New(hashKey, blockKey)

	// create an admin for this assignment
	// FUTURE: Please do not create admin this way lolz.
	// Obviously this is for the assignment and not meant for real-life production.
	adminCode, err := HashPassword("password")
	if err != nil {
		log.Fatalln("Unable to generate password for admin.")
	}
	// create an admin
	adminA := &Admin{"A000", "Admin", "Istrator", "admin", adminCode, time.Now(), time.Time{}}
	// concurrency-safe appending adminA to admins
	jhbsAdmins.mu.Lock()
	jhbsAdmins.admins = append(jhbsAdmins.admins, adminA)
	jhbsAdmins.mu.Unlock()

}
