package localizer

import (
	// Import the internal/translations so that it's init() function
	// is run. It's really important that we do this here so that the
	// default message catalog is updated to use our translations
	// *before* we initialize the message.Printer instances below.

	"fmt"
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
		ID:      "fr-fr",
		printer: message.NewPrinter(language.MustParse("fr-FR")),
	},
	{
		ID:      "en-us",
		printer: message.NewPrinter(language.MustParse("en-US")),
	},
	{
		ID:      "zh-cn",
		printer: message.NewPrinter(language.MustParse("zh-CN")),
	},
	{
		ID:      "zh-tw",
		printer: message.NewPrinter(language.MustParse("zh-TW")),
	},
	{
		ID:      "it-it",
		printer: message.NewPrinter(language.MustParse("it-IT")),
	},
	{
		ID:      "ja-jp",
		printer: message.NewPrinter(language.MustParse("ja-JP")),
	},
	{
		ID:      "ko-kr",
		printer: message.NewPrinter(language.MustParse("ko-KR")),
	},
	{
		ID:      "pt-br",
		printer: message.NewPrinter(language.MustParse("pt-BR")),
	},
	{
		ID:      "ru-ru",
		printer: message.NewPrinter(language.MustParse("ru-RU")),
	},
	{
		ID:      "es-es",
		printer: message.NewPrinter(language.MustParse("es-ES")),
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

// init() initializes the language automatically
// based on env var SQLCMD_LANG which expects language
// tag such as en-us, de-de, fr-ch, etc.
func init() {
	localeName := os.Getenv("SQLCMD_LANG")
	Translator = Get(localeName)
}

// Errorf() is wrapper function to create localized errors
func Errorf(format string, a ...any) error {
	errMsg := Translator.printer.Sprintf(format, a...)
	return fmt.Errorf(errMsg)
}

// Srrorf() is wrapper function to create localized string
func Sprintf(key message.Reference, args ...interface{}) string {
	return Translator.printer.Sprintf(key, args...)
}
