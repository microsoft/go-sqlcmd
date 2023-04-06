package mechanism

var mechanisms = []Mechanism{
	&attach{},
	&bacpac{},
	&restore{},
}

func FileTypes() []string {
	fileTypes := []string{}
	for _, m := range mechanisms {
		fileTypes = append(fileTypes, m.FileTypes()...)
	}
	return fileTypes
}
