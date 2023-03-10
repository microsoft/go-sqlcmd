package assert

import "testing"

func NotNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) bool {
	return false
}

func NoError(t *testing.T, err error, msgAndArgs ...interface{}) bool {
	return false
}
