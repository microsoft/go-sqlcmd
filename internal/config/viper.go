// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"bytes"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func configureViper(configFile string) {
	if configFile == "" {
		panic("Must provide configFile")
	}

	createEmptyFileIfNotExistsCallback(configFile)

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("SQLCMD")
	viper.SetConfigFile(configFile)
}

func load() {
	var err error

	err = viper.ReadInConfig()
	checkErr(err)
	err = viper.BindEnv("ACCEPT_EULA")
	checkErr(err)
	viper.AutomaticEnv() // read in environment variables that match
	err = viper.Unmarshal(&config)
	checkErr(err)

	trace("Config loaded from file: %v", viper.ConfigFileUsed())
}

func Save() {
	b, err := yaml.Marshal(&config)
	checkErr(err)
	err = viper.ReadConfig(bytes.NewReader(b))
	checkErr(err)
	createEmptyFileIfNotExistsCallback(viper.ConfigFileUsed())
	err = viper.WriteConfig()
	checkErr(err)
}

func GetConfigFileUsed() string {
	return viper.ConfigFileUsed()
}
