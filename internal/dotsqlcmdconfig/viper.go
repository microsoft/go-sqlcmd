// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package dotsqlcmdconfig

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
)

// Load loads the configuration from the file specified by the SetFileName() function.
// Any errors encountered while marshalling or saving the configuration are checked
// and handled by the injected errorHandler (via the checkErr function).
func Load() {
	if filename == "" {
		panic("Must call config.SetFileName()")
	}

	text := file.GetContents(filename)
	err := yaml.Unmarshal([]byte(text), &config)
	checkErr(err)

	trace("Config loaded from file: %v"+pal.LineBreak(), filename)
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

	var io io.WriteCloser

	b, err := yaml.Marshal(&config)
	checkErr(err)

	_, err = io.Write(b)
	checkErr(err)
}

// GetConfigFileUsed returns the path to the configuration file used by the Viper library.
func GetConfigFileUsed() string {
	return viper.ConfigFileUsed()
}
