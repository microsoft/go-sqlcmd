package extract

type Extractor interface {
	FileTypes() []string
	IsInstalled(containerId string) bool
	Install()
	Extract(srcFile string, destFolder string) (filename string, ldfFilename string)
}
