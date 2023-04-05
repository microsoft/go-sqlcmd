package location

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
)

func NewLocation(isLocal bool, uri string, controller *container.Controller) Location {
	if isLocal {
		return local{
			uri:        uri,
			controller: controller,
		}
	} else {
		return remote{
			uri:        uri,
			controller: controller,
		}
	}
}
