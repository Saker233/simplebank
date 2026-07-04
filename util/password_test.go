package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)



func TestPassword(t *testing.T) {
	password := RandomString(6)

	hashedpassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedpassword1)

	err = CheckPassword(password, hashedpassword1)
	require.NoError(t, err)

	wrongPassword := RandomString(6)

	err = CheckPassword(wrongPassword, hashedpassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedpassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedpassword2)
	require.NotEqual(t, hashedpassword1, hashedpassword2)
}