// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package sqlconfig, defines the metadata for representing a sqlconfig file.
// It includes structs for representing an endpoint, context, user, and the overall
// sqlconfig file itself. Each struct has fields for storing the various pieces
// of information that make up an SQL configuration, such as endpoint address
// and port, context name and endpoint, and user authentication type and details.
// These structs are used to manage and manipulate the sqlconfig.
package sqlconfig

type EndpointDetails struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type ContainerDetails struct {
	Id    string `mapstructure:"id"`
	Image string `mapstructure:"image"`
}

type AssetDetails struct {
	*ContainerDetails `mapstructure:"container,omitempty" yaml:"container,omitempty"`
}

type Endpoint struct {
	*AssetDetails   `mapstructure:"asset,omitempty" yaml:"asset,omitempty"`
	EndpointDetails `mapstructure:"endpoint" yaml:"endpoint"`
	Name            string `mapstructure:"name"`
}

type ContextDetails struct {
	Endpoint string  `mapstructure:"endpoint"`
	User     *string `mapstructure:"user,omitempty" yaml:"user,omitempty"`
}

type Context struct {
	ContextDetails `mapstructure:"context" yaml:"context"`
	Name           string `mapstructure:"name"`
}

type BasicAuthDetails struct {
	Username          string `mapstructure:"username"`
	PasswordEncrypted bool   `mapstructure:"password-encrypted" yaml:"password-encrypted"`
	Password          string `mapstructure:"password"`
}

type User struct {
	Name               string            `mapstructure:"name"`
	AuthenticationType string            `mapstructure:"authentication-type" yaml:"authentication-type"`
	BasicAuth          *BasicAuthDetails `mapstructure:"basic-auth,omitempty" yaml:"basic-auth,omitempty"`
}

type Sqlconfig struct {
	Version        string     `mapstructure:"version"`
	Endpoints      []Endpoint `mapstructure:"endpoints"`
	Contexts       []Context  `mapstructure:"contexts"`
	CurrentContext string     `mapstructure:"currentcontext"`
	Users          []User     `mapstructure:"users"`
}
