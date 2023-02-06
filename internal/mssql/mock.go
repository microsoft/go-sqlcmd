package mssql

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
)

type MssqlType struct{}
type MssqlMock struct{}

func (m *MssqlMock) Connect(endpoint sqlconfig.Endpoint, user *sqlconfig.User, console sqlcmd.Console) *sqlcmd.Sqlcmd {
	return &sqlcmd.Sqlcmd{
		Exitcode:          0,
		Connect:           nil,
		Format:            nil,
		Query:             "",
		Cmd:               nil,
		PrintError:        nil,
		UnicodeOutputFile: false,
	}
}

func (m *MssqlMock) Query(s *sqlcmd.Sqlcmd, text string) {

}
