package sql

import . "github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"

func (m *SqlMock) Connect(
	endpoint Endpoint,
	user *User,
	interactive bool,
) {
}

func (m *SqlMock) Query(text string) {

}
