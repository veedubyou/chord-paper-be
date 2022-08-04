package userstorage

import "github.com/cockroachdb/errors/domains"

var (
	UserNotFoundMark = domains.New("user_not_found")
	DefaultErrorMark = domains.New("default_error")
)
