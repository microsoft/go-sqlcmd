// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

// Formatter defines a formatter for serializing an input object into a byte slice.
// The Serialize method serializes the input object and returns the resulting
// byte slice. The CheckErr method handles any error encountered during
// the serialization process.
type Formatter interface {
	Serialize(in interface{}) (bytes []byte)
	CheckErr(err error)
}
