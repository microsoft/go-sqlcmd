// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

type GlobalOptions struct {
	TrustServerCertificate bool
	DatabaseName           string
	UseTrustedConnection   bool
	UserName               string
	Endpoint               string
	AuthenticationMethod   string
	UseAad                 bool
	PacketSize             int
	LoginTimeout           int
	WorkstationName        string
	ApplicationIntent      string
	Encrypt                string
	DriverLogLevel         int
}

var globalOptions = &GlobalOptions{}
