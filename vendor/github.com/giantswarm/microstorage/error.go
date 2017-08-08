package microstorage

import "github.com/giantswarm/microerror"

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts InvalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

// NotFoundError is exported because it is supposed to be used by the interface
// implementors in order to fulfil the API.
var NotFoundError = microerror.New("not found")

// IsNotFound asserts NotFoundError. The library user's code should use this
// public key matcher to verify if some storage error is of type NotFoundError.
func IsNotFound(err error) bool {
	return microerror.Cause(err) == NotFoundError
}
