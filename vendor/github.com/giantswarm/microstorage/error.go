package microstorage

import (
	"github.com/juju/errgo"
)

var InvalidConfigError = errgo.New("invalid config")

// IsInvalidConfig asserts InvalidConfigError.
func IsInvalidConfig(err error) bool {
	return errgo.Cause(err) == InvalidConfigError
}

var NotFoundError = errgo.New("not found")

// IsNotFound asserts NotFoundError. The library user's code should use this
// public key matcher to verify if some storage error is of type NotFoundError.
func IsNotFound(err error) bool {
	return errgo.Cause(err) == NotFoundError
}
