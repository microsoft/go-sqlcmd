module github.com/microsoft/go-sqlcmd

go 1.16

require (
	github.com/alecthomas/kong v0.2.18-0.20210621093454-54558f65e86f
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20210722231415-061457976a23 // indirect
	github.com/denisenkom/go-mssqldb v0.12.0
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c
	github.com/golang-sql/sqlexp v0.0.0-20170517235910-f1bb20e5a188
	github.com/google/uuid v1.3.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/denisenkom/go-mssqldb => github.com/shueybubbles/go-mssqldb v0.10.1-0.20220303143659-8896461e4ec7
