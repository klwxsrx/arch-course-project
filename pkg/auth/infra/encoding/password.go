package encoding

import (
	"github.com/klwxsrx/arch-course-project/pkg/auth/app/service"
	"golang.org/x/crypto/bcrypt"
)

type passwordEncoder struct {
}

func (e *passwordEncoder) Encode(pwd string) (string, error) {
	result, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(result), err
}

func (e *passwordEncoder) Check(password, encoded string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password))
	return err == nil
}

func NewPasswordEncoder() service.PasswordEncoder {
	return &passwordEncoder{}
}
