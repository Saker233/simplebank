package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error) {
	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hasedPassword), nil
}

func CheckPassword(password, hashedpassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(password))
}