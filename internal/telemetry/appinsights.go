package telemetry

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
)

var telemetryClient appinsights.TelemetryClient
var isTelemetryEnabled string = "true"

type TelemetrySqlcmd struct {
	EventName  string
	Properties map[string]string
}

// Event Names
const (
	Legacy                  = "sqlcmd/legacy"
	Config                  = "sqlcmd/config"
	ConfigAddContext        = "sqlcmd/config/addcontext"
	ConfigAddEndpoint       = "sqlcmd/config/addendpoint"
	ConfigAddUser           = "sqlcmd/config/adduser"
	ConfigConnectionStrings = "sqlcmd/config/connectionstrings"
	ConfigCurrentContexts   = "sqlcmd/config/currentcontexts"
	ConfigDeleteContext     = "sqlcmd/config/deletecontext"
	ConfigDeleteEndpoint    = "sqlcmd/config/deleteendpoint"
	ConfigDeleteUser        = "sqlcmd/config/deleteuser"
	ConfigGetContexts       = "sqlcmd/config/getcontexts"
	ConfigGetEndpoints      = "sqlcmd/config/getendpoints"
	ConfigGetUsers          = "sqlcmd/config/getusers"
	ConfigUseContext        = "sqlcmd/config/usecontext"
	ConfigView              = "sqlcmd/config/view"
	Create                  = "sqlcmd/create"
	CreateAzsqlEdge         = "sqlcmd/create/azsqledge"
	CreateMssql             = "sqlcmd/create/mssql"
	Delete                  = "sqlcmd/delete"
	SqlcmdHelp              = "sqlcmd/help"
	Open                    = "sqlcmd/open"
	SqlcmdQuery             = "sqlcmd/query"
	Start                   = "sqlcmd/start"
	Stop                    = "sqlcmd/stop"
)

// Property Names
var (
	// Common Properties
	userId    string
	sessionId string
	locLang   string
	userOs    string
)

const (

	//Common Properties
	UserId    = "Sqlcmd.userId"
	SessionId = "Sqlcmd.sessionId"
	LocLang   = "Sqlcmd.locLang"
	UserOs    = "Sqlcmd.userOs"

	// Legacy Flags
	HelpSymbol                  = "Sqlcmd.Legacy.?"
	Helpflag                    = "Sqlcmd.Legacy.help.h"
	Sqlconfig                   = "Sqlcmd.sqlconfig"
	Verbosity                   = "Sqlcmd.verbosity"
	Version                     = "Sqlcmd.version"
	ApplicationIntent           = "Sqlcmd.Legacy.application-intent.K"
	AuthenticationMethod        = "Sqlcmd.Legacy.authentication-method"
	BatchTerminator             = "Sqlcmd.Legacy.batch-terminator.c"
	ChangePassword              = "Sqlcmd.Legacy.change-password.z"
	ChangePasswordExit          = "Sqlcmd.Legacy.change-password-exit.Z"
	ClientRegionalSetting       = "Sqlcmd.Legacy.client-regional-setting.R"
	ColumnSeparator             = "Sqlcmd.Legacy.column-separator.s"
	DatabaseName                = "Sqlcmd.Legacy.database-name.d"
	DedicatedAdminConnection    = "Sqlcmd.Legacy.dedicated-admin-connection.A"
	DisableCmdAndWarn           = "Sqlcmd.Legacy.disable-cmd-and-warn.X"
	DisableVariableSubstitution = "Sqlcmd.Legacy.disable-variable-substitution.x"
	DriverLoggingLevel          = "Sqlcmd.Legacy.driver-logging-level"
	EchoInput                   = "Sqlcmd.Legacy.echo-input.e"
	EnableColumnEncryption      = "Sqlcmd.Legacy.enable-column-encryption.g"
	EnableQuotedIdentifiers     = "Sqlcmd.Legacy.enable-quoted-identifiers.I"
	EncryptConnection           = "Sqlcmd.Legacy.encrypt-connection.N"
	ErrorLevel                  = "Sqlcmd.Legacy.error-level.m"
	ErrorSeverityLevel          = "Sqlcmd.Legacy.error-severity-level.V"
	ErrorsToStderr              = "Sqlcmd.Legacy.errors-to-stderr.r"
	ExitOnError                 = "Sqlcmd.Legacy.exit-on-error.b"
	FixedTypeWidth              = "Sqlcmd.Legacy.fixed-type-width.Y"
	Format                      = "Sqlcmd.Legacy.format.F"
	Headers                     = "Sqlcmd.Legacy.headers.h"
	InitialQuery                = "Sqlcmd.Legacy.initial-query.q"
	InputFile                   = "Sqlcmd.Legacy.input-file.i"
	ListServers                 = "Sqlcmd.Legacy.list-servers.L"
	LoginTimeOut                = "Sqlcmd.Legacy.login-timeOut.l"
	MultiSubnetFailover         = "Sqlcmd.Legacy.multi-subnet-failover.M"
	OutputFile                  = "Sqlcmd.Legacy.output-file.o"
	PacketSize                  = "Sqlcmd.Legacy.packet-size.a"
	Password                    = "Sqlcmd.Legacy.password.P"
	Query                       = "Sqlcmd.Legacy.query.Q"
	QueryTimeout                = "Sqlcmd.Legacy.query-timeout.t"
	RemoveControlCharacters     = "Sqlcmd.Legacy.remove-control-characters.k"
	ScreenWidth                 = "Sqlcmd.Legacy.screen-width.w"
	Server                      = "Sqlcmd.Legacy.server.S"
	TrimSpaces                  = "Sqlcmd.Legacy.trim-spaces.W"
	TrustServerCertificate      = "Sqlcmd.Legacy.trust-server-certificate.C"
	UnicodeOutputFile           = "Sqlcmd.Legacy.unicode-output-file.u"
	UseAad                      = "Sqlcmd.Legacy.use-aad.G"
	UseTrustedConnection        = "Sqlcmd.Legacy.use-trusted-connection.E"
	UserName                    = "Sqlcmd.Legacy.user-name.U"
	VariableTypeWidth           = "Sqlcmd.Legacy.variable-type-width.y"
	Variables                   = "Sqlcmd.Legacy.variables.v"
	WorkstationName             = "Sqlcmd.Legacy.workstation-name.H"

	// Modern Flags
	ConfigAddContextHelp            = "Sqlcmd.Config.AddContext.help"
	ConfigAddContextEndpoint        = "Sqlcmd.Config.AddContext.Endpoint"
	ConfigAddContextName            = "Sqlcmd.Config.AddContext.Name"
	ConfigAddContextUser            = "Sqlcmd.Config.AddContext.User"
	ConfigAddEndpointHelp           = "Sqlcmd.Config.AddEndpoint.help"
	ConfigAddEndpointAddress        = "Sqlcmd.Config.AddEndpoint.Address"
	ConfigAddEndpointName           = "Sqlcmd.Config.AddEndpoint.Name"
	ConfigAddEndpointPort           = "Sqlcmd.Config.AddEndpoint.Port"
	ConfigAddUserHelp               = "Sqlcmd.Config.AddUser.help"
	ConfigAddUserAuthType           = "Sqlcmd.Config.AddUser.AuthType"
	ConfigAddUserName               = "Sqlcmd.Config.AddUser.Name"
	ConfigAddUserPasswordEncryption = "Sqlcmd.Config.AddUser.PasswordEncryption"
	ConfigAddUserUsername           = "Sqlcmd.Config.AddUser.Username"
	ConfigAddUserPassword           = "Sqlcmd.Config.AddUser.Password"
	ConfigConnectionStringsHelp     = "Sqlcmd.Config.ConnectionStrings.help"
	ConfigConnectionStringsDatabase = "Config.ConnectionStrings.database.d"
	ConfigCurrentContextsHelp       = "Sqlcmd.Config.CurrentContexts.help"
	ConfigDeleteContextHelp         = "Sqlcmd.Config.DeleteContext.help"
	ConfigDeleteContextCascade      = "Sqlcmd.Config.DeleteContext.Cascade"
	ConfigDeleteContextname         = "Sqlcmd.Config.DeleteContext.name"
	ConfigDeleteEndpointHelp        = "Sqlcmd.Config.DeleteEndpoint.help"
	ConfigDeleteEndpointName        = "Sqlcmd.Config.DeleteEndpoint.Name"
	ConfigDeleteUserHelp            = "Sqlcmd.Config.DeleteUser.help"
	ConfigDeleteUserName            = "Sqlcmd.Config.DeleteUser.Name"
	ConfigGetContextsHelp           = "Sqlcmd.Config.GetContexts.help"
	ConfigGetContextsDetailed       = "Sqlcmd.Config.GetContexts.Detailed"
	ConfigGetContextsName           = "Sqlcmd.Config.GetContexts.Name"
	ConfigGetEndpointsHelp          = "Sqlcmd.Config.GetEndpoints.help"
	ConfigGetEndpointsDetailed      = "Sqlcmd.Config.GetEndpoints.Detailed"
	ConfigGetEndpointsName          = "Sqlcmd.Config.GetEndpoints.Name"
	ConfigGetUsersHelp              = "Sqlcmd.Config.GetUsers.help"
	ConfigGetUsersDetailed          = "Sqlcmd.Config.GetUsers.Detailed"
	ConfigGetUsersName              = "Sqlcmd.Config.GetUsers.Name"
	ConfigUseContextHelp            = "Sqlcmd.Config.UseContext.help"
	ConfigUseContextname            = "Sqlcmd.Config.UseContext.name"
	ConfigViewHelp                  = "Sqlcmd.Config.View.help"
	ConfigViewRaw                   = "Sqlcmd.Config.View.Raw"
	ConfigHelp                      = "Sqlcmd.Config.help"

	CreateHelp                          = "Sqlcmd.Create.help"
	CreateAzsqlEdgeHelp                 = "Sqlcmd.Create.AzsqlEdge.help"
	CreateAzsqlEdgeGetTagsHelp          = "Sqlcmd.Create.AzsqlEdge.GetTags.help"
	CreateAzsqlEdgeAcceptEula           = "Sqlcmd.Create.AzsqlEdge.AcceptEula"
	CreateAzsqlEdgeArchitecture         = "Sqlcmd.Create.AzsqlEdge.Architecture"
	CreateAzsqlEdgeCached               = "Sqlcmd.Create.AzsqlEdge.Cached"
	CreateAzsqlEdgeCollation            = "Sqlcmd.Create.AzsqlEdge.Collation"
	CreateAzsqlEdgeContextName          = "Sqlcmd.Create.AzsqlEdge.ContextName.c"
	CreateAzsqlEdgeErrorlogWaitLine     = "Sqlcmd.Create.AzsqlEdge.ErrorlogWaitLine"
	CreateAzsqlEdgeHostname             = "Sqlcmd.Create.AzsqlEdge.Hostname"
	CreateAzsqlEdgeName                 = "Sqlcmd.Create.AzsqlEdge.Name"
	CreateAzsqlEdgeOs                   = "Sqlcmd.Create.AzsqlEdge.Os"
	CreateAzsqlEdgePasswordEncryption   = "Sqlcmd.Create.AzsqlEdge.PasswordEncryption"
	CreateAzsqlEdgePasswordLength       = "Sqlcmd.Create.AzsqlEdge.PasswordLength"
	CreateAzsqlEdgePasswordMinNumber    = "Sqlcmd.Create.AzsqlEdge.PasswordMinNumber"
	CreateAzsqlEdgePasswordMinSpecial   = "Sqlcmd.Create.AzsqlEdge.PasswordMinSpecial"
	CreateAzsqlEdgePasswordMinUpper     = "Sqlcmd.Create.AzsqlEdge.PasswordMinUpper"
	CreateAzsqlEdgePasswordSpecialChars = "Sqlcmd.Create.AzsqlEdge.PasswordSpecialChars"
	CreateAzsqlEdgePort                 = "Sqlcmd.Create.AzsqlEdge.Port"
	CreateAzsqlEdgeRegistry             = "Sqlcmd.Create.AzsqlEdge.Registry"
	CreateAzsqlEdgeRepo                 = "Sqlcmd.Create.AzsqlEdge.Repo"
	CreateAzsqlEdgeTag                  = "Sqlcmd.Create.AzsqlEdge.Tag"
	CreateAzsqlEdgeUserDatabase         = "Sqlcmd.Create.AzsqlEdge.UserDatabase.u"
	CreateAzsqlEdgeUsing                = "Sqlcmd.Create.AzsqlEdge.Using"

	CreateMssqlHelp                 = "Sqlcmd.Create.Mssql.help"
	CreateMssqlGetTagsHelp          = "Sqlcmd.Create.Mssql.GetTags.help"
	CreateMssqlAcceptEula           = "Sqlcmd.Create.Mssql.AcceptEula"
	CreateMssqlArchitecture         = "Sqlcmd.Create.Mssql.Architecture"
	CreateMssqlCached               = "Sqlcmd.Create.Mssql.Cached"
	CreateMssqlCollation            = "Sqlcmd.Create.Mssql.Collation"
	CreateMssqlContextName          = "Sqlcmd.Create.Mssql.ContextName.c"
	CreateMssqlErrorlogWaitLine     = "Sqlcmd.Create.Mssql.ErrorlogWaitLine"
	CreateMssqlHostname             = "Sqlcmd.Create.Mssql.Hostname"
	CreateMssqlName                 = "Sqlcmd.Create.Mssql.Name"
	CreateMssqlOs                   = "Sqlcmd.Create.Mssql.Os"
	CreateMssqlPasswordEncryption   = "Sqlcmd.Create.Mssql.PasswordEncryption"
	CreateMssqlPasswordLength       = "Sqlcmd.Create.Mssql.PasswordLength"
	CreateMssqlPasswordMinNumber    = "Sqlcmd.Create.Mssql.PasswordMinNumber"
	CreateMssqlPasswordMinSpecial   = "Sqlcmd.Create.Mssql.PasswordMinSpecial"
	CreateMssqlPasswordMinUpper     = "Sqlcmd.Create.Mssql.PasswordMinUpper"
	CreateMssqlPasswordSpecialChars = "Sqlcmd.Create.Mssql.PasswordSpecialChars"
	CreateMssqlPort                 = "Sqlcmd.Create.Mssql.Port"
	CreateMssqlRegistry             = "Sqlcmd.Create.Mssql.Registry"
	CreateMssqlRepo                 = "Sqlcmd.Create.Mssql.Repo"
	CreateMssqlTag                  = "Sqlcmd.Create.Mssql.Tag"
	CreateMssqlUserDatabase         = "Sqlcmd.Create.Mssql.UserDatabase.u"
	CreateMssqlUsing                = "Sqlcmd.Create.Mssql.Using"

	DeleteHelp  = "Sqlcmd.Delete.help"
	DeleteForce = "Sqlcmd.Delete.Force"
	DeleteYes   = "Sqlcmd.Delete.Yes"

	Help = "Sqlcmd.help"

	OpenHelp    = "Sqlcmd.Open.help"
	OpenAdsHelp = "Sqlcmd.Open.Ads.help"

	QueryHelp     = "Sqlcmd.Query.help"
	QueryDatabase = "Sqlcmd.Query.database.d"
	Queryquery    = "Sqlcmd.Query.query.q"
	QueryText     = "Sqlcmd.Query.Text.t"
	StartHelp     = "Sqlcmd.Start.help"
	StopHelp      = "Sqlcmd.Stop.help"
)

func initCommonProps() {
	if userId == "" {
		userId = getOrCreateUserId()
	}
	if sessionId == "" {
		sessionId = uuid.NewString()
	}
	locLang = localizer.LocaleName
	userOs = runtime.GOOS
}

func init() {
	telemetryEnabled := strings.ToLower(os.Getenv("SQLCMD_TELEMETRY"))
	if telemetryEnabled == "false" {
		isTelemetryEnabled = "false"
	}
	if isTelemetryEnabled == "true" {
		InitializeAppInsights()
		initCommonProps()
	}
}

func InitializeTelemetryLogging() {
	loggingEnabled := strings.ToLower(os.Getenv("SQLCMD_TELEMETRY_LOGGING"))
	if isTelemetryEnabled == "true" && loggingEnabled == "true" {
		// Add a diagnostics listener for printing telemetry messages
		appinsights.NewDiagnosticsMessageListener(func(msg string) error {
			fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
			return nil
		})
	}
}

func InitializeAppInsights() {
	instrumentationKey := "f305b208-557d-4fba-bf06-25345c4dfdbc"
	config := appinsights.NewTelemetryConfiguration(instrumentationKey)
	telemetryClient = appinsights.NewTelemetryClientFromConfig(config)
	InitializeTelemetryLogging()
}

func TrackEvent(eventName string, properties map[string]string) {
	if isTelemetryEnabled == "true" {
		event := appinsights.NewEventTelemetry(eventName)
		for key, value := range properties {
			event.Properties[key] = value
		}
		telemetryClient.Track(event)
	}
}

func CloseTelemetry() {
	select {
	case <-telemetryClient.Channel().Close(10 * time.Second):
		// Ten second timeout for retries.

		// If we got here, then all telemetry was submitted
		// successfully, and we can proceed to exiting.
	case <-time.After(30 * time.Second):
		// Thirty second absolute timeout.  This covers any
		// previous telemetry submission that may not have
		// completed before Close was called.

		// There are a number of reasons we could have
		// reached here.  We gave it a go, but telemetry
		// submission failed somewhere.  Perhaps old events
		// were still retrying, or perhaps we're throttled.
		// Either way, we don't want to wait around for it
		// to complete, so let's just exit.
	}
}

func CloseTelemetryAndExit(exitcode int) {
	if isTelemetryEnabled == "true" {
		CloseTelemetry()
	}
	os.Exit(exitcode)
}
