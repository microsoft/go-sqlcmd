package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/credman"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPersistCredentialForAds(t *testing.T) {
	ads := Ads{}
	ads.SetCrossCuttingConcerns(dependency.Options{
		EndOfLine: "",
		Output:    output.New(output.Options{}),
	})

	user := &sqlconfig.User{
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:          "testuser",
			Password:          "testpass",
			PasswordEncrypted: false,
		},
	}
	ads.persistCredentialForAds("localhost", sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{
			Port: 1433,
		},
	}, user)

	// Test if the correct target name is generated
	expectedTargetName := "Microsoft.SqlTools|itemtype:Profile|id:providerName:MSSQL|applicationName:azdata|authenticationType:SqlLogin|database:|server:localhost,1433|user:testuser"
	if ads.credential.TargetName != expectedTargetName {
		t.Errorf("Expected target name to be %s, got %s", expectedTargetName, ads.credential.TargetName)
	}

	// Test if the correct username is set
	if ads.credential.UserName != user.BasicAuth.Username {
		t.Errorf("Expected username to be %s, got %s", user.BasicAuth.Username, ads.credential.UserName)
	}

	// Test if the password is decoded correctly
	decodedPassword := secret.DecodeAsUtf16(user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted)
	assert.Equal(
		t,
		ads.credential.CredentialBlob,
		decodedPassword,
		"Expected decoded password to be %v, got %v",
		decodedPassword,
		ads.credential.CredentialBlob,
	)
}

func TestRemovePreviousCredential(t *testing.T) {
	ads := Ads{}
	ads.SetCrossCuttingConcerns(dependency.Options{
		EndOfLine: "",
		Output:    output.New(output.Options{}),
	})

	ads.credential = credman.Credential{
		TargetName: "TestTargetName",
		Persist:    credman.PersistSession,
	}

	ads.writeCredential()
	ads.removePreviousCredential()
}
