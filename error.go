package main

import (
	"github.com/giantswarm/microerror"
)

var maskTooBigError = microerror.New("mask too big")

// IsMaskTooBig asserts maskTooBigError.
func IsMaskTooBig(err error) bool {
	return microerror.Cause(err) == maskTooBigError
}

var nilIPError = microerror.New("nil IP")

// IsNilIP asserts nilIPError.
func IsNilIP(err error) bool {
	return microerror.Cause(err) == nilIPError
}

var ipNotContainedError = microerror.New("ip not contained")

// IsIPNotContained asserts ipNotContainedError.
func IsIPNotContainer(err error) bool {
	return microerror.Cause(err) == ipNotContainedError
}

var maskIncorrectSizeError = microerror.New("mask incorrect size")

// IsMaskIncorrectSize asserts maskIncorrectSizeError.
func IsMaskIncorrectSize(err error) bool {
	return microerror.Cause(err) == maskIncorrectSizeError
}
