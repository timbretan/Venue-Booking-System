package jhbshttp

import (
	"fmt"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// this password hash uses bcrypt

const cost = 14

func HashPassword(password string) (string, error) {
	start := time.Now()
	// uses 2^14 iterations (cost 14) takes avg 3.3s; 2^15 iterations (cost 15) takes avg 6.2s!
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	fmt.Println("It takes", time.Since(start), "to hash 2^"+strconv.Itoa(cost)+" iterations of the pw.")
	return string(bytes), err
}

// CompareHashAndPassword compares a bcrypt hashed password with its possible plaintext equivalent.
// Returns nil on success, or an error on failure.
// for CheckPasswordHash, nil is true, error is false
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
