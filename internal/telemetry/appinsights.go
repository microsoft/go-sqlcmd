package telemetry

import (
	"fmt"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

var telemetryClient appinsights.TelemetryClient

func SetTelemetryClient(client appinsights.TelemetryClient) {
	telemetryClient = client
}

type Telemetry struct {
	Client appinsights.TelemetryClient
}

func InitializeAppInsights() {
	instrumentationKey := "7d1a32e5-4dbb-4b11-ae2d-b8591bcf4cba"
	config := appinsights.NewTelemetryConfiguration(instrumentationKey)
	telemetryClient = appinsights.NewTelemetryClientFromConfig(config)
	SetTelemetryClient(telemetryClient)
	// Add a diagnostics listener for printing telemetry messages
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
		return nil
	})
}

// TrackCommand tracks a command execution event
func TrackCommand(command string) {
	event := appinsights.NewEventTelemetry("command")
	event.Properties["command"] = command
	telemetryClient.Track(event)
	FlushTelemetry()
}

// TrackSubCommand tracks a sub-command execution event
func TrackSubCommand(command, subCommand string) {
	event := appinsights.NewEventTelemetry("sub-command")
	event.Properties["command"] = command
	event.Properties["sub-command"] = subCommand
	telemetryClient.Track(event)
	FlushTelemetry()
}

func TrackEvent(eventName string) {
	event := appinsights.NewEventTelemetry(eventName)
	event.Properties["command"] = eventName
	telemetryClient.Track(event)
}

func FlushTelemetry() {
	telemetryClient.Channel().Flush()
}

// func SetTelemetryClient(client appinsights.TelemetryClient) {
// 	telemetryClient = client
// }

func LogMessage(msg string) error {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
	return nil
}
