// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package argo

import (
	"os"
)

type Error string

func (err Error) String() string {
	return "argo error: " + string(err)
}

type TypeMismatchError string

func (err TypeMismatchError) String() string {
	return "argo error: " + string(err)
}

var OutOfRangeError = Error("out of range size value")

func error(err os.Error) {
	panic(err)
}

func typeError(expected string, got byte) os.Error {
	return TypeMismatchError("expected " + expected + ", got " + typeNames[got])
}
