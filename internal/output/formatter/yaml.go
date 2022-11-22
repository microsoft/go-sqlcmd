// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"gopkg.in/yaml.v2"
)

type Yaml struct {
	Base
}

func (f *Yaml) Serialize(in interface{}) (bytes []byte) {
	var err error

	bytes, err = yaml.Marshal(in)
	f.Base.CheckErr(err)
	f.Base.Output(bytes)

	return
}
