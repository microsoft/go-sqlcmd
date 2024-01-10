package mechanism

var mechanisms = []Mechanism{
	&attach{},
	&dacfx{},
	&git2{},
	&restore{},
	&script{},
}

func FileTypes() []string {
	fileTypes := []string{}
	for _, m := range mechanisms {
		fileTypes = append(fileTypes, m.FileTypes()...)
	}
	return fileTypes
}
