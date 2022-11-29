// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

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
	User     *string `mapstructure:"user,omitempty"`
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
	ApiVersion     string     `mapstructure:"apiVersion"`
	Endpoints      []Endpoint `mapstructure:"endpoints"`
	Contexts       []Context  `mapstructure:"contexts"`
	CurrentContext string     `mapstructure:"currentcontext"`
	Kind           string     `mapstructure:"kind"`
	Users          []User     `mapstructure:"users"`
}
