package config

import (
	"testing"

	. "github.com/microsoft/go-sqlcmd/cmd/sqlconfig"
)

func TestAddUser(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	AddUser(User{
		Name:               "",
		AuthenticationType: "basic",
		BasicAuth:          nil,
	})
}

func TestAddUser2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	AddUser(User{
		Name:               "",
		AuthenticationType: "basic",
		BasicAuth: &BasicAuthDetails{
			Username:          "",
			PasswordEncrypted: false,
			Password:          "",
		},
	})
}
