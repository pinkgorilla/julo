package account

import "errors"

var (
	ErrAccountAlreadyExists = errors.New("account already exist")
	ErrAccountNotFound      = errors.New("account not found")
)
