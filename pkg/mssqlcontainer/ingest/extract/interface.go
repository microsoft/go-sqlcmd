package extract

import "github.com/microsoft/go-sqlcmd/internal/container"

type Extractor interface {
	FileTypes() []string
	Initialize(controller *container.Controller)
	IsInstalled(containerId string) bool
	Install()
	Extract(srcFile string, destFolder string) (filename string, ldfFilename string)
}
