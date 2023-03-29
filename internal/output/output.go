// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package Output provides a number of methods for logging and handling
// errors, including Debugf, Errorf, Fatalf, FatalErr, Infof, Panic, Panicf,
// Struct, Tracef, and Warnf. These methods allow the caller to specify the
// desired verbosity level, adds a newline to the end of the log message if
// necessary, and handle errors and hints in a variety of ways.
//
// Trace("Something very low level.") - not localized
// Debug("Useful debugging information.") - not localized
// Info("Something noteworthy happened!") - localized
// Warn("You should probably take a look at this.") - localized
// Error("Something failed but I'm not quitting.") - localized
// Fatal("Bye.") - localized, calls os.Exit(1) after logging
// Panic("I'm bailing.") - not localized, calls panic() after logging
package output

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/pkg/errors"
	"regexp"
	"strings"
)

func (o Output) Debugf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Debug {
		format = o.ensureEol(format)
		o.printf("DEBUG: "+format, a...)
	}
}

func (o Output) Errorf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Error {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "ERROR: " + format
		}
		o.printf(format, a...)
	}
}

func (o Output) Fatal(a ...any) {
	o.fatal([]string{}, a...)
}
func (o Output) FatalErr(err error) {
	o.errorCallback(err)
}

func (o Output) Fatalf(format string, a ...any) {
	o.fatalf([]string{}, format, a...)
}

func (o Output) FatalfErrorWithHints(err error, hints []string, format string, a ...any) {
	o.hintCallback(hints)
	s := fmt.Sprintf(format, a...)
	o.errorCallback(fmt.Errorf(s+": %w", err))
}

func (o Output) FatalfWithHints(hints []string, format string, a ...any) {
	o.fatalf(hints, format, a...)
}

func (o Output) FatalfWithHintExamples(hintExamples [][]string, format string, a ...any) {
	err := errors.New(fmt.Sprintf(format, a...))
	o.displayHintExamples(hintExamples)
	o.errorCallback(err)
}

func (o Output) FatalWithHints(hints []string, a ...any) {
	o.fatal(hints, a...)
}

func (o Output) Infof(format string, a ...any) {
	o.infofWithHints([]string{}, format, a...)
}

func (o Output) InfofWithHints(hints []string, format string, a ...any) {
	o.infofWithHints(hints, format, a...)
}

// InfofWithHintExamples logs an info-level message with a given format and
// arguments. It also displays additional hints with example usage in the
// output.
func (o Output) InfofWithHintExamples(hintExamples [][]string, format string, a ...any) {
	if o.loggingLevel >= verbosity.Info {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "INFO:  " + format
		}
		o.printf(format, a...)
		o.displayHintExamples(hintExamples)
	}
}

func (o Output) Panic(a ...any) {
	panic(a)
}

func (o Output) Panicf(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}

func (o Output) Struct(in interface{}) (bytes []byte) {
	bytes = o.formatter.Serialize(in)

	return
}

func (o Output) Tracef(format string, a ...any) {
	if o.loggingLevel >= verbosity.Trace {
		format = o.ensureEol(format)
		o.printf("TRACE: "+format, a...)
	}
}

func (o Output) Warnf(format string, a ...any) {
	if o.loggingLevel >= verbosity.Warn {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "WARN:  " + format
		}
		o.printf(format, a...)
	}
}

// displayHintExamples takes an array of hint examples and displays them in
// a formatted way. It first calculates the maximum length of the description
// in the hint examples, and then creates a string for each hint example with
// the description padded to the maximum length, followed by the example.
// Finally, it calls the hint callback function with the array of formatted hints.
func (o Output) displayHintExamples(hintExamples [][]string) {
	var hints []string

	maxLengthHintText := 0
	for _, hintExample := range hintExamples {
		if len(hintExample) != 2 {
			panic("Hintexample must be 2 elements, a description, and an example")
		}

		if len(hintExample[0]) > maxLengthHintText {
			maxLengthHintText = len(hintExample[0])
		}
	}

	for _, hintExample := range hintExamples {
		padLength := maxLengthHintText - len(hintExample[0])
		hints = append(hints, fmt.Sprintf(
			"%v: %v%s",
			hintExample[0],
			strings.Repeat(" ", padLength),
			hintExample[1],
		))
	}
	o.hintCallback(hints)
}

// ensureEol ensures that the provided format string ends with a line break character.
// It does this by checking if the format string already ends with a line break character,
// and if not, it appends a line break character to the format string. If the format
// string is shorter than the length of the line break character, it returns the line
// break character on its own. This function is useful for ensuring that output to
// the console will always be properly formatted and easy to read.
func (o Output) ensureEol(format string) string {
	if len(format) >= len(pal.LineBreak()) {
		if !strings.HasSuffix(format, pal.LineBreak()) {
			format = format + pal.LineBreak()
		}
	} else {
		format = pal.LineBreak()
	}
	return format
}

func (o Output) fatal(hints []string, a ...any) {
	err := errors.New(fmt.Sprintf("%v", a...))
	o.hintCallback(hints)
	o.errorCallback(err)
}

func (o Output) fatalf(hints []string, format string, a ...any) {
	err := errors.New(fmt.Sprintf(format, a...))
	o.hintCallback(hints)
	o.errorCallback(err)
}

// infofWithHints is used to print out an "INFO" message with additional hints.
// The format argument specifies the text to be printed, which can include placeholders
// for dynamic values. The a argument is a variadic parameter containing the
// values to be used to replace the placeholders in the format string. The hints
// argument is a slice of strings representing additional hints to be printed along
// with the message. The function checks if the logging level is set to at least
// "Info" before printing the message. If the logging level is set to "Debug"
// or higher, the string "INFO: " is prepended to the message before it is printed.
// The function also calls the hintCallback function to print the hints, if any are provided.
func (o Output) infofWithHints(hints []string, format string, a ...any) {
	if o.loggingLevel >= verbosity.Info {
		format = o.ensureEol(format)
		if o.loggingLevel >= verbosity.Debug {
			format = "INFO:  " + format
		}
		o.printf(format, a...)
		o.hintCallback(hints)
	}
}

// maskSecrets takes a string as input and masks any password found in the
// string using the PASSWORD.*\s?=.*\s?N?') regular expression. It
// returns the resulting masked string.
func (o Output) maskSecrets(text string) string {

	// Mask password from T/SQL e.g. ALTER LOGIN [sa] WITH PASSWORD = N'foo';
	r := regexp.MustCompile(`(PASSWORD.*\s?=.*\s?N?')(.*)(')`)
	text = r.ReplaceAllString(text, "$1********$3")

	// Mask password from sqlpackage.exe command line e.g. /TargetPassword:foo
	r = regexp.MustCompile(`(/TargetPassword:)(.*)( )`)
	text = r.ReplaceAllString(text, "$1********$3")

	return text
}

func (o Output) printf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	text = o.maskSecrets(text)
	_, err := o.standardWriteCloser.Write([]byte(text))
	o.errorCallback(err)
}
