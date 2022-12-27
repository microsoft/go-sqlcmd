// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"bytes"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Load loads the configuration from the file specified by the SetFileName() function.
// Any errors encountered while marshalling or saving the configuration are checked
// and handled by the injected errorHandler (via the checkErr function).
func Load() {
	if filename == "" {
		panic("Must call config.SetFileName()")
	}

	var err error
	err = viper.ReadInConfig()
	checkErr(err)
	err = viper.BindEnv("ACCEPT_EULA")
	checkErr(err)
	viper.AutomaticEnv() // read in environment variables that match
	err = viper.Unmarshal(&config)
	checkErr(err)

	trace("Config loaded from file: %v"+pal.LineBreak(), viper.ConfigFileUsed())
}

// Save marshals the current configuration object and saves it to the configuration
// file previously specified by the SetFileName variable.
// Any errors encountered while marshalling or saving the configuration are checked
// and handled by the injected errorHandler (via the checkErr function).
func Save() {
	if filename == "" {
		panic("Must call config.SetFileName()")
	}

	if config.Version == "" {
		config.Version = "v1"
	}

	b, err := yaml.Marshal(&config)
	checkErr(err)
	err = viper.ReadConfig(bytes.NewReader(b))
	checkErr(err)
	err = viper.WriteConfig()
	checkErr(err)
}

// GetConfigFileUsed returns the path to the configuration file used by the Viper library.
func GetConfigFileUsed() string {
	return viper.ConfigFileUsed()
}

// configureViper initializes the Viper library with the given configuration file.
// This function sets the configuration file type to "yaml" and sets the environment variable prefix to "SQLCMD".
// It also sets the configuration file to use to the one provided as an argument to the function.
// This function is intended to be called at the start of the application to configure Viper before any other code uses it.
func configureViper(configFile string) {
	if configFile == "" {
		panic("Must provide configFile")
	}

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("SQLCMD")
	viper.SetConfigFile(configFile)
}
