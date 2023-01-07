// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"encoding/json"
)

type Json struct {
	Base
}

func (f *Json) Serialize(in interface{}) (bytes []byte) {
	var err error

	bytes, err = json.MarshalIndent(in, "", "  ")
	f.Base.CheckErr(err)
	f.Base.Output(bytes)

	return
}
