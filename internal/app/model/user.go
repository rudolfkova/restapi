package model

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                int    `json:"id"`
	Email             string `json:"email"`
	EncryptedPassword string `json:"password,omitempty"`
	Password          string `json:"-"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.By(requiredIf(u.EncryptedPassword == "")), validation.Length(6, 100)),
	)
}

func (u *User) BeforCreate() error {
	if len(u.Password) == 0 {
		return fmt.Errorf("No password for user %d", u.ID)
	}
	enc, err := encryptString(u.Password)
	if err != nil {
		return err
	}
	u.EncryptedPassword = enc
	return nil
}

func (u *User) Sanitize() {
	u.Password = ""
}

func encryptString(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
