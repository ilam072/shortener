package repo

import "errors"

var (
	ErrAliasAlreadyExists = errors.New("alias already exists")
	ErrAliasNotFound      = errors.New("alias not found")
)
