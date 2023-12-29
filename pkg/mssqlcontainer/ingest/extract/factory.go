package extract

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
)

func NewExtractor(fileExtension string, controller *container.Controller) Extractor {
	for _, extractor := range extractors {
		for _, ext := range extractor.FileTypes() {
			if ext == fileExtension {
				extractor.Initialize(controller)
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
