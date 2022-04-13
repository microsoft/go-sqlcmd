module github.com/microsoft/go-sqlcmd

go 1.16

require (
	github.com/alecthomas/kong v0.5.0
	github.com/denisenkom/go-mssqldb v0.12.0
	github.com/golang-sql/sqlexp v0.0.0-20170517235910-f1bb20e5a188
	github.com/google/uuid v1.3.0
	github.com/peterh/liner v1.2.2
	github.com/stretchr/testify v1.7.1
	golang.org/x/text v0.3.6
)

replace github.com/denisenkom/go-mssqldb => github.com/shueybubbles/go-mssqldb v0.10.1-0.20220317022252-fafb9d92e469
