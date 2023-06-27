// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sql

type SqlOptions struct {
	UnitTesting bool
}

func NewSql(options SqlOptions) Sql {
	if options.UnitTesting {
		return &mock{}
	} else {
		return &mssql{}
	}
}
