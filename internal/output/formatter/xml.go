// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"encoding/xml"
)

type Xml struct {
	Base
}

func (f *Xml) Serialize(in interface{}) (bytes []byte) {
	var err error

	bytes, err = xml.MarshalIndent(in, "", "    ")
	f.Base.CheckErr(err)
	f.Base.Output(bytes)

	return
}
