package sql

func New(unitTesting bool) Sql {
	if unitTesting {
		return &SqlMock{}
	} else {
		return &SqlType{}
	}
}
