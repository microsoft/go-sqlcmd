package location

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/http"
)

type remote struct {
	uri        string
	controller *container.Controller
}

func (l remote) IsLocal() bool {
	return false
}

func (l remote) ValidSchemes() []string {
	return []string{"https", "http"}
}

// Verify the file exists at the URL
func (l remote) Exists() bool {
	return http.UrlExists(l.uri)
}

func (l remote) CopyToContainer(containerId string, destFolder string) {
	l.controller.DownloadFile(
		containerId,
		l.uri,
		destFolder,
	)
}
