package localizer

import (
	// Import the internal/translations so that it's init() function
	// is run. It's really important that we do this here so that the
	// default message catalog is updated to use our translations
	// *before* we initialize the message.Printer instances below.
	"os"

	_ "github.com/microsoft/go-sqlcmd/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Translator Localizer

// Define a Localizer type which stores the relevant locale ID and a
// (deliberately unexported) message.Printer instance for the locale.
type Localizer struct {
	ID      string
	printer *message.Printer
}

// Initialize a slice which holds the initialized Localizer types for
// each of our supported locales.
var locales = []Localizer{
	{
		ID:      "de-de",
		printer: message.NewPrinter(language.MustParse("de-DE")),
	},
	{
		ID:      "fr-ch",
		printer: message.NewPrinter(language.MustParse("fr-CH")),
	},
	{
		ID:      "en-us",
		printer: message.NewPrinter(language.MustParse("en-US")),
	},
}

// The Get() function accepts a locale ID and returns the corresponding
// Localizer for that locale. If the locale ID is not supported then
// this returns `false` as the second return value.
func Get(id string) Localizer {
	for _, locale := range locales {
		if id == locale.ID {

			return locale
		}
	}
	return Localizer{
		ID:      "en-us",
		printer: message.NewPrinter(language.MustParse("en-US")),
	}

}

func init() {
	localeName := os.Getenv("SQLCMD_LANG")
	Translator = Get(localeName)
}

// We also add a Translate() method to the Localizer type. This acts
// as a wrapper around the unexported message.Printer's Sprintf()
// function and returns the appropriate translation for the given
// message and arguments.
func Translate(key message.Reference, args ...interface{}) string {
	return Translator.printer.Sprintf(key, args...)
}
