package mechanism

func NewMechanism(fileExtension string, name string) Mechanism {
	for _, m := range mechanisms {
		if m.Name() == name {
			return m
		}
	}

	return NewMechanismByFileExt(fileExtension)
}

func NewMechanismByFileExt(fileExtension string) Mechanism {
	for _, m := range mechanisms {
		for _, ext := range m.FileTypes() {
			if ext == fileExtension {
				return m
			}
		}
	}
	return nil
}
