package extract

import "github.com/microsoft/go-sqlcmd/internal/container"

type tar struct {
	controller *container.Controller
}

func (e *tar) FileTypes() []string {
	return []string{"tar"}
}

func (e *tar) Initialize(controller *container.Controller) {
	e.controller = controller
}

func (e *tar) IsInstalled(containerId string) bool {
	return true
}

func (e *tar) Extract(srcFile string, destFolder string) (string, string) {
	return "", ""
}

func (e *tar) Install() {
}
