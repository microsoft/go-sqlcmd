package mechanism

import "github.com/microsoft/go-sqlcmd/internal/container"

func NewMechanism(fileExtension string, name string, controller *container.Controller) Mechanism {
	for _, m := range mechanisms {
		if m.Name() == name {
			m.Initialize(controller)
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
				return m
			}
		}
	}
	return nil
}
