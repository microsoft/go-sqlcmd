package formatter

import (
	"fmt"
	"os"
)

// New creates a new instance of the Formatter interface. It takes an Options
// struct as input and sets default values for some of the fields if they are
// not specified. The SerializationFormat field of the Options struct is used
// to determine which implementation of the Formatter interface to return.
// If the specified format is not supported, the function will panic.
func New(options Options,
) (f Formatter) {
	if options.SerializationFormat == "" {
		options.SerializationFormat = "yaml"
	}
	if options.ErrorHandler == nil {
		options.ErrorHandler = func(err error) {}
	}
	if options.StandardOutput == nil {
		options.StandardOutput = os.Stdout
	}

	switch options.SerializationFormat {
	case "json":
		f = &Json{Base: Base{
			StandardOutput:       options.StandardOutput,
			ErrorHandlerCallback: options.ErrorHandler}}
	case "yaml":
		f = &Yaml{Base: Base{
			StandardOutput:       options.StandardOutput,
			ErrorHandlerCallback: options.ErrorHandler}}
	case "xml":
		f = &Xml{Base: Base{
			StandardOutput:       options.StandardOutput,
			ErrorHandlerCallback: options.ErrorHandler}}
	default:
		panic(fmt.Sprintf("Format '%v' not supported", options.SerializationFormat))
	}

	return
}
