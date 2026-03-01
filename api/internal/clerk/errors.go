// Package clerk provides integration with the Clerk authentication and user management webhook events.
package clerk

import "errors"

// ErrUserNotFound is returned when attempting to act on a Clerk user that does not exist locally.
var ErrUserNotFound = errors.New("user not found")
