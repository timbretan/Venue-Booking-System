package jhbshttp

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

// validateString validates any input named under field param
func validateString(r *http.Request, field string) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue(field)) // done at server side

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Empty " + field)
	}
	// if field has non-English chars
	if m, _ := regexp.MatchString("^[a-zA-Z ]+$", u); !m {
		return "", errors.New("Invalid " + field)
	}
	return u, nil
}

// validateTier validates membership tiers, using any input named "tier"
// requires that the input value be numbers, not membership tier names
func validateTier(r *http.Request) (int, error) {
	// number the four membership tiers
	tierNumbers := []int{0, 1, 2, 3}

	tierStr := r.FormValue("tier")
	tier, err := strconv.Atoi(tierStr)
	if err != nil {
		return -1, errors.New("Invalid membership tier")
	}

	for _, v := range tierNumbers {
		if v == tier {
			return v, nil
		}
	}
	return -1, errors.New("Invalid membership tier")
}

// validateMobile checks whether user supplied valid SG mobile number
func validateMobile(r *http.Request) (int, error) {

	mobileStr := r.FormValue("mobile")
	mobile, err := strconv.Atoi(mobileStr)

	// if user did not type in a mobile number, or mobile number is not btw. 81000000 and 98999999 (incl.), throw error
	if err != nil || mobile < 81000000 || mobile > 98999999 {
		return -1, errors.New("Mobile number is not valid SG number")
	}

	return mobile, nil
}

// validateUsername validates any input named "username"
func validateUsername(r *http.Request) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue("username")) // done at server side

	// make username lowercase
	u = strings.ToLower(u)

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Empty username")
	}
	// check if field has English chars, dash, or numbers
	if m, _ := regexp.MatchString("^[a-z0-9-]+$", u); !m {
		return "", errors.New("Invalid username")
	}
	return u, nil
}

// validatePassword validates any input named "code"
func validatePassword(r *http.Request) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue("code")) // done at server side

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Empty password")
	}
	return u, nil
}

// NB: All passwords, after validation here, go through the following fn below!
// Password validates plain password against the rules defined below.
//
// upp: at least one upper case letter.
// low: at least one lower case letter.
// num: at least one digit.
// sym: at least one special character.
// tot: at least eight characters long.
// No empty string or whitespace.

// Only used for membership signups, not for daily logins.
func IsValidPassword(pass string) bool {
	var (
		upp, low, num, sym bool
		tot                uint8
	)

	for _, char := range pass {
		switch {
		case unicode.IsUpper(char):
			upp = true
			tot++
		case unicode.IsLower(char):
			low = true
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
			return false
		}
	}

	if !upp || !low || !num || !sym || tot < 8 {
		return false
	}

	return true
}

// validateOldPassword validates any input named "oldcode"
func validateOldPassword(r *http.Request) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue("oldcode")) // done at server side

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Empty old password")
	}
	return u, nil
}

// validateConfirmNewPassword validates any input named "confirmcode"
func validateConfirmNewPassword(r *http.Request) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue("confirmcode")) // done at server side

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Please type in your new password to confirm")
	}
	return u, nil
}

// validateVenue validates any input named under field param
func validateVenue(r *http.Request, field string) (string, error) {
	// escape strings first
	u := template.HTMLEscapeString(r.FormValue(field)) // done at server side

	// if field is empty
	if len(u) == 0 {
		return "", errors.New("Empty " + field)
	}
	// if field has non-English chars
	if m, _ := regexp.MatchString("^[a-zA-Z0-9- ]+$", u); !m {
		return "", errors.New("Invalid " + field)
	}
	return u, nil
}

// for membershownvenue.gohtml,
// when user books a venue, these fields need to be validated:
// venue (see validateVenue abv), day and hour (for start and end)
// FUTURE: Make validateDay also validate day and month (e.g. 1 Aug 2021)
func validateDay(r *http.Request, field string) (int, error) {
	// escape strings first
	dayStr := template.HTMLEscapeString(r.FormValue(field)) // done at server side

	// convert string to int
	day, err := strconv.Atoi(dayStr)

	// if user did not select a day, throw error
	// FUTURE: Make this 28-31 depending on month!
	if err != nil || day < 1 || day > 30 {
		return -1, errors.New("Invalid day number")
	}

	return day, nil

}

// validateHour validates the hour
func validateHour(r *http.Request, field string) (int, error) {
	// escape strings first
	hourStr := template.HTMLEscapeString(r.FormValue(field)) // done at server side

	// convert string to int
	hour, err := strconv.Atoi(hourStr)

	// if user did not select an hour, throw error
	// hours work as 0:00 - 23:59 and 24:00 becomes 0:00 the next day
	if err != nil || hour < 0 || hour > 23 {
		return -1, errors.New("Invalid hour number")
	}

	return hour, nil

}

// validateSessionID validates the sessionID
// by looking at "sessionid" from the form
func validateSessionID(r *http.Request) (string, error) {
	// escape strings first
	sessionID := template.HTMLEscapeString(r.FormValue("sessionid")) // done at server side

	// if field is empty
	if len(sessionID) == 0 {
		return "", errors.New("Empty sessionid")
	}
	// if field has non-English chars
	if m, _ := regexp.MatchString("^[a-zA-Z0-9_=-]+$", sessionID); !m {
		return "", errors.New("Invalid sessionid")
	}

	return sessionID, nil

}

// validateSessionID validates the sessionID
// by looking at "sessionid" from the form
func validateMemberID(r *http.Request) (int, error) {
	// escape strings first
	mIDStr := template.HTMLEscapeString(r.FormValue("memberid")) // done at server side

	// if field is empty
	// convert string to int
	mID, err := strconv.Atoi(mIDStr)

	// if field has non-English chars
	if err != nil || mID < 100000 || mID > 999999 {
		return -1, errors.New("Invalid memberid")
	}

	return mID, nil

}

// validateNumber validates the number
func validateNumber(r *http.Request, field string) (int, error) {
	// escape strings first
	numStr := template.HTMLEscapeString(r.FormValue(field)) // done at server side

	// convert string to int
	num, err := strconv.Atoi(numStr)

	// if user did not select a number, throw error
	if err != nil {
		return -1, errors.New("Invalid hour number")
	}

	return num, nil

}

// showErrorOnTop displays an error msg
// e.g. if user did not sign up properly
// by leaving a field blank, tries to insert <script>s into a field
func showErrorOnTop(w http.ResponseWriter, tmpl string, data interface{}, str string) {
	tpl.ExecuteTemplate(w, tmpl, data)
	io.WriteString(w,
		`<div class="pop-up negative"><p>`+str+`</p></div>`,
	)
}

// showSuccessOnTop displays a success msg
func showSuccessOnTop(w http.ResponseWriter, tmpl string, data interface{}, str string) {
	tpl.ExecuteTemplate(w, tmpl, data)
	io.WriteString(w,
		`<div class="pop-up positive"><p>`+str+`</p></div>`,
	)
}
