// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

type Formatter interface {
	Serialize(in interface{}) (bytes []byte)
	CheckErr(err error)
}
