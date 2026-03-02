// Package user encapsulates domain logic and data access for the User entity.
package user

import "errors"

// ErrUserNotFound is returned when an expected user cannot be found in the system.
var ErrUserNotFound = errors.New("user not found")
