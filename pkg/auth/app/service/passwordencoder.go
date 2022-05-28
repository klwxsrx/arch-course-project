package service

type PasswordEncoder interface {
	Encode(pwd string) (string, error)
	Check(password, encoded string) bool
}
