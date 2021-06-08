package jhbs

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

// all these are internal functions

// generateID() generates a 6-letter string commonly used for tickets
// this one allows 24 * 23 * 23 * 23 * 23 * 23 = 154,472,232 letter combinations
// does not guarantee that strings generated will be unique
func generateID() string {
	// allowed letters
	allowedLetters := "ABCDEFGHJKLMNPQRSTUVWXYZ"
	lenAllowedLetters := len(allowedLetters)

	// start with an empty ID
	genID := ""
	// for 0th letter
	genID += string(allowedLetters[rand.Intn(lenAllowedLetters)])
	// for 1st-5th letters
	for i := 1; i < 6; i++ {
		s := allowedLetters[rand.Intn(lenAllowedLetters)]
		// prevent consecutive duplicate letters
		for s == genID[i-1] {
			s = allowedLetters[rand.Intn(lenAllowedLetters)]
		}
		genID += string(s)
	}
	return genID
}

// for cleanUpMemberNames(), in which there should be only letters and spaces (e.g. "Michael van der Aa")
var IsLetter = regexp.MustCompile(`^[a-zA-Z ]+$`).MatchString

// cleanUpTheWord() makes words lowercase and removes front and back spaces
func cleanUpTheWord(word string) string {
	word = strings.ToLower(word)
	word = strings.TrimSpace(word)
	return word
}

// cleanUpMemberNames() removes front and back spaces and numbers (human names have no numbers),
// then returns the value in Title Case
// returns nil, error if invalid name (alert user that it has typed in the wrong name)
func cleanUpMemberNames(word string) (string, error) {
	if !IsLetter(word) {
		err := fmt.Sprintf("You supplied %s, but only letters a-z or A-Z accepted", word)
		return "", errors.New(err)
	}
	word = strings.Title(strings.TrimSpace(strings.ToLower(word)))
	return word, nil
}

// cleanUpTheWordForTries() makes words lowercase and removes front and back spaces
func cleanUpTheWordForTries(word string) string {
	word = strings.TrimSpace(strings.ToLower(word))
	return word
}

func checkUserInputForMonthAndHour(entry string, mode string) (day int, hour int, err error) {

	if !(mode == "start" || mode == "end") {
		return -1, -1, errors.New("- Please supply \"start\" or \"end\"")
	}

	timeStrSlice := strings.Split(entry, " ")
	if len(timeStrSlice) != 2 {
		return -1, -1, errors.New("- Invalid input. Back to venue selection.")

	}
	day, err = strconv.Atoi(timeStrSlice[0])
	if err != nil {
		return -1, -1, errors.New("- day NaN. Back to venue selection.")
	}
	// hard-coded June to have 30 days
	// in the future, startDay > 30 becomes startDay > daysInAMonth[5]
	if day < 1 || day > 30 {
		return -1, -1, errors.New("- invalid day given. Back to venue selection.")
	}

	hour, err = strconv.Atoi(timeStrSlice[1])
	if err != nil {
		return -1, -1, errors.New("- hour NaN. Back to venue selection.")
	}

	switch mode {

	case "start":
		if hour < 0 || hour > 23 {
			return -1, -1, errors.New("- invalid start hour given. Back to main menu.")
		}

	case "end":
		if hour < 1 || hour > 24 {
			return -1, -1, errors.New("- invalid end hour given. Back to main menu.")
		}
	}
	return day, hour, nil

}

// internal function
// only admins can print that out
// prevents flooding of messages
// pAdmin stands for "printf for admin"
func pfa(s string, a ...interface{}) {
	//	if mode == "-admin" {
	fmt.Printf(s, a...)
	//	}
}

// only admins can print that out
// prevents flooding of messages
// pAdmin stands for "println for admin"
func plna(a ...interface{}) {
	//	if mode == "-admin" {
	fmt.Println(a...)
	//	}
}
