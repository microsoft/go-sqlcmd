package telemetry

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

var telemetryClient appinsights.TelemetryClient
var isTelemetryEnabled string = "true"
var uniqTelemetryID string = ""

func initUniqUserId() {
	if uniqTelemetryID == "" {
		user, _ := user.Current()
		id, _ := machineid.ProtectedID(user.Username)
		uniqTelemetryID = id
	}
}

func init() {
	telemetryEnabled := strings.ToLower(os.Getenv("SQLCMD_TELEMETRY"))
	if telemetryEnabled == "false" {
		isTelemetryEnabled = "false"
	}
	if isTelemetryEnabled == "true" {
		InitializeAppInsights()
		initUniqUserId()
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
		event.Properties["userId"] = uniqTelemetryID
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
