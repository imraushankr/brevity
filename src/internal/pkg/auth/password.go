package auth

import "golang.org/x/crypto/bcrypt"

func EncryptPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func IsPasswordCorrect(password, has string) error {
	return bcrypt.CompareHashAndPassword([]byte(has), []byte(password))
}