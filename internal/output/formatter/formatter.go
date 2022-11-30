package formatter

import "fmt"
import "io"

func NewFormatter(
	serializationFormat string,
	standardOutput io.WriteCloser,
	errorHandler func(err error),
) (f Formatter) {
	switch serializationFormat {
	case "json":
		f = &Json{Base: Base{
			StandardOutput:       standardOutput,
			ErrorHandlerCallback: errorHandler}}
	case "yaml":
		f = &Yaml{Base: Base{
			StandardOutput:       standardOutput,
			ErrorHandlerCallback: errorHandler}}
	case "xml":
		f = &Xml{Base: Base{
			StandardOutput:       standardOutput,
			ErrorHandlerCallback: errorHandler}}
	default:
		panic(fmt.Sprintf("Format '%v' not supported", serializationFormat))
	}

	return
}
