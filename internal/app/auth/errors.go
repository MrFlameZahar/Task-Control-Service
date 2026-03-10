package auth

import "errors"

var (
	ErrUserExists         = errors.New("user exists")
	ErrUserNotFound         = errors.New("user doesnt exist")
	ErrInvalidCredentials = errors.New("invalid credentials")
)