// Package that run password encryptions
package bcrypt

import "golang.org/x/crypto/bcrypt"

// Method to encrypt a password
var Hash = func(password string) (string, error) {
	// Generate the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}
