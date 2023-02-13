package open

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/credman"
	"github.com/microsoft/go-sqlcmd/internal/secret"
)

// Type Ads is used to implement the "open ads" which launches Azure
// Data Studio and establishes a connection to the SQL Server for the current
// context
type Ads struct {
	cmdparser.Cmd

	credential credman.Credential
}

// persistCredentialForAds stores a SQL password in the Windows Credential Manager
// for the given hostname and endpoint.
func (c *Ads) persistCredentialForAds(
	hostname string,
	endpoint sqlconfig.Endpoint,
	user *sqlconfig.User,
) {
	// Create the target name that ADS will look for
	targetName := c.adsKey(
		fmt.Sprintf("%s,%d", hostname, rune(endpoint.Port)),
		"", // The default database is set on the user login
		"SqlLogin",
		user.BasicAuth.Username)

	// Store the SQL password in the Windows Credential Manager with the
	// generated target name
	c.credential = credman.Credential{
		TargetName: targetName,
		CredentialBlob: secret.DecodeAsUtf16(
			user.BasicAuth.Password, user.BasicAuth.PasswordEncrypted),
		UserName: user.BasicAuth.Username,
		Persist:  credman.PersistSession,
	}

	c.removePreviousCredential()
	c.writeCredential()
}

// adsKey returns the credential target name for the given instance, database,
// authentication type, and user.
func (c *Ads) adsKey(instance, database, authType, user string) string {
	return fmt.Sprintf(
		"Microsoft.SqlTools|"+
			"itemtype:Profile|"+
			"id:providerName:MSSQL|"+
			"applicationName:azdata|"+
			"authenticationType:%s|"+
			"database:%s|"+
			"server:%s|"+
			"user:%s",
		authType, database, instance, user)
}

// removePreviousCredential removes any previously stored credentials with
// the same target name as the current instance's credential.
func (c *Ads) removePreviousCredential() {
	credentials, err := credman.EnumerateCredentials("", true)
	c.CheckErr(err)

	for _, cred := range credentials {
		if cred.TargetName == c.credential.TargetName {
			err = credman.DeleteCredential(cred, credman.CredTypeGeneric)
			c.CheckErr(err)
			break
		}
	}
}

// writeCredential stores the current instance's credential in the Windows Credential Manager
func (c *Ads) writeCredential() {
	err := credman.WriteCredential(&c.credential, credman.CredTypeGeneric)
	c.CheckErr(err)
}
