package config

import "testing"

func Test_configureViper(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	configureViper("")
}
