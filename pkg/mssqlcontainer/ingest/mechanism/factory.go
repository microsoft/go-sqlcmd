package mechanism

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
)

func NewMechanism(fileExtension string, name string, controller *container.Controller) Mechanism {
	trace("NewMechanism: fileExtension = %q, name = %q"+fileExtension, name)
	for _, m := range mechanisms {
		if m.Name() == name {
			m.Initialize(controller)

			trace("Returning: %q", m.Name())

			return m
		}
	}

	return NewMechanismByFileExt(fileExtension, controller)
}

func NewMechanismByFileExt(fileExtension string, controller *container.Controller) Mechanism {
	for _, m := range mechanisms {
		for _, ext := range m.FileTypes() {
			if ext == fileExtension {
				m.Initialize(controller)

				trace("Returning: %q", m.Name())

				return m
			}
		}
	}

	trace("No mechanism found for file extension %q", fileExtension)

	return nil
}

func Mechanisms() []string {
	m := []string{}
	for _, mechanism := range mechanisms {
		m = append(m, mechanism.Name())
	}
	return m
}
