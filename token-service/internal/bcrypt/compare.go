// Package that run password encryptions
package bcrypt

import "golang.org/x/crypto/bcrypt"

// Method to compare passworrd and hashed password
var Compare = func(password, hashedPassword string) bool {
	// Generate the password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	return err == nil
}
