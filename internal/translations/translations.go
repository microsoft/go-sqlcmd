package translations

//go:generate gotext -srclang=en-US update -out=catalog.go -lang=en-US,de-DE,fr-CH github.com/microsoft/go-sqlcmd/cmd/sqlcmd github.com/microsoft/go-sqlcmd/pkg/sqlcmd
