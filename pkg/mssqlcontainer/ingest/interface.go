package ingest

type Ingest interface {
	IsRemoteUrl() bool
	IsValidScheme() bool
	IsValidFileExtension() bool
	IsExtractionNeeded() bool

	SourceFileExists() bool
	DatabaseName() string
	UrlFilename() string
	OnlineMethod() string
	UserProvidedFileExt() string

	CopyToContainer(containerId string)
	Extract()
	BringOnline(query func(string), username string, password string)

	ValidSchemes() []string
}
