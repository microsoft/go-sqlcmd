package localizer

const (
	UseContextCommand   = "sqlcmd config use-context mssql"
	PasswordEnvVar      = "SQLCMD_PASSWORD"
	PasswordEnvVar2     = "SQLCMDPASSWORD"
	EndpointFlag        = "--endpoint"
	FeedbackUrl         = "https://aka.ms/sqlcmd-feedback"
	PasswordEncryptFlag = "--password-encryption"
	AuthTypeFlag        = "--auth-type"
	ModernAuthTypeBasic = "basic"
	ModernAuthTypeOther = "other"
	UserNameFlag        = "--username"
	NameFlag            = "--name"
	GetContextCommand   = "sqlcmd config get-contexts"
	GetEndpointsCommand = "sqlcmd config get-endpoints"
	GetUsersCommand     = "sqlcmd config get-users"
	RunQueryExample     = "sqlcmd query \"SELECT @@SERVERNAME\""
	UninstallCommand    = "sqlcmd uninstall"
	AcceptEulaFlag      = "--accept-eula"
	AcceptEulaEnvVar    = "SQLCMD_ACCEPT_EULA"
	PodmanPsCommand     = "podman ps"
	DockerPsCommand     = "docker ps"
)
