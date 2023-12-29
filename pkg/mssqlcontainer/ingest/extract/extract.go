package extract

var extractors = []Extractor{
	&tar{},
	&sevenZip{},
}
