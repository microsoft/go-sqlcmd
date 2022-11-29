package config

import "testing"

func TestEndpointExists(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	EndpointExists("")
}
