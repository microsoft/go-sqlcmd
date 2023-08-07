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

func SetTelemetryClientFromInstrumentationKey(instrumentationKey string) {
	telemetryClient = appinsights.NewTelemetryClient(instrumentationKey)
}

type Telemetry struct {
	Client appinsights.TelemetryClient
}

func InitializeAppInsights() {
	instrumentationKey := "3fdeec77-2951-456f-bcb2-9f003b512a3f"
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
}

// TrackSubCommand tracks a sub-command execution event
func TrackSubCommand(command, subCommand string) {
	event := appinsights.NewEventTelemetry("sub-command")
	event.Properties["command"] = command
	event.Properties["sub-command"] = subCommand
	telemetryClient.Track(event)
}

func TrackEvent(eventName string, properties map[string]string) {
	event := appinsights.NewEventTelemetry(eventName)
	for key, value := range properties {
		event.Properties[key] = value
	}
	telemetryClient.Track(event)
}

func CloseTelemetry() {
	telemetryClient.Channel().Close()
}
func FlushTelemetry() {
	telemetryClient.Channel().Flush()
}

func LogMessage(msg string) error {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
	return nil
}
