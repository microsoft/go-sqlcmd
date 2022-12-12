package formatter

import "io"

// Options defines the options for creating a new Formatter instance.
type Options struct {
	SerializationFormat string
	StandardOutput      io.WriteCloser
	ErrorHandler        func(err error)
}
