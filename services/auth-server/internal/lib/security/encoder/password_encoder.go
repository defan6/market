package encoder

import "golang.org/x/crypto/bcrypt"

type PasswordEncoder struct{}

func NewPasswordEncoder() *PasswordEncoder {
	return &PasswordEncoder{}
}

func (p *PasswordEncoder) EncodePassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (p *PasswordEncoder) ComparePassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}
