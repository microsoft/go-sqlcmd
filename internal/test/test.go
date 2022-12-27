package test

import (
	"testing"
)

// CatchExpectedError function is a helper function for use in unit tests. It
// expects an error value as its first argument, and a pointer to a testing.T struct
// as its second argument. If the error is not nil, the function logs the error
// and continues execution of the unit test. If the error is nil, the function
// panics with a message indicating that the code did not panic as expected.
// This function is used to verify that a particular code path produces an error
// in a unit test.
func CatchExpectedError(err any, t *testing.T) {
	if err != nil {
		t.Log("The expected error was:", err)
	} else {
		panic("The code did not panic as expected")
	}
}
