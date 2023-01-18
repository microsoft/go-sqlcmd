package mssql

func New(unitTesting bool) MssqlInterface {
	if unitTesting {
		return &MssqlMock{}
	} else {
		return &MssqlType{}
	}
}
