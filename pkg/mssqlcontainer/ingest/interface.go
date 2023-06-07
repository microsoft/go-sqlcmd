package ingest

type Ingest interface {
	IsRemoteUrl() bool
	IsValidScheme() bool
	IsValidFileExtension() bool

	SourceFileExists() bool
	DatabaseName() string
	UrlFilename() string
	OnlineMethod() string
	UserProvidedFileExt() string

	CopyToContainer(containerId string)
	BringOnline(query func(string), username string, password string)

	ValidSchemes() []string
}
