// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pal

func init() {
	Initialize(func(err error) {
		if err != nil {
			panic(err)
		}
	})
}

func Initialize(handler func(err error)) {
	errorCallback = handler
}
