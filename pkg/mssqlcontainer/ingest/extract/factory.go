package extract

func NewExtractor(fileExtension string) Extractor {
	for _, extractor := range extractors {
		for _, ext := range extractor.FileTypes() {
			if ext == fileExtension {
				return extractor
			}
		}
	}
	return nil
}

func FileTypes() []string {
	types := []string{}
	for _, extractor := range extractors {
		for _, ext := range extractor.FileTypes() {
			types = append(types, ext)
		}
	}
	return types
}
