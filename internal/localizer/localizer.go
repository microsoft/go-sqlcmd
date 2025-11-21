// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package localizer

import (
	// Import the internal/translations so that its init() function
	// is run. It's really important that we do this here so that the
	// default message catalog is updated to use our translations
	// *before* we initialize the message.Printer instances below.

	"fmt"
	"os"
	"strings"

	_ "github.com/microsoft/go-sqlcmd/internal/translations"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var Translator *message.Printer

var supportedLanguages = map[string]string{
	"de-de":   "de-DE",
	"fr-fr":   "fr-FR",
	"en-us":   "en-US",
	"zh-hans": "zh-CN",
	"zh-cn":   "zh-CN",
	"zh-hant": "zh-TW",
	"zh-tw":   "zh-TW",
	"it-it":   "it-IT",
	"ja-jp":   "ja-JP",
	"ko-kr":   "ko-KR",
	"pt-br":   "pt-BR",
	"ru-ru":   "ru-RU",
	"es-es":   "es-ES",
}

// init() initializes the language automatically
// based on env var SQLCMD_LANG which expects language
// tag such as en-us, de-de, fr-ch, etc.
func init() {
	localeName := strings.ToLower(os.Getenv("SQLCMD_LANG"))
	if _, ok := supportedLanguages[localeName]; !ok {
		localeName = "en-us"
	}
	Translator = message.NewPrinter(language.MustParse(supportedLanguages[localeName]))
}

// Errorf() is wrapper function to create localized errors
func Errorf(format string, a ...any) error {
	errMsg := Translator.Sprintf(format, a...)
	return fmt.Errorf("%s", errMsg)
}

// Sprintf() is wrapper function to create localized string
func Sprintf(key message.Reference, args ...interface{}) string {
	return Translator.Sprintf(key, args...)
}

// ProductBanner() returns the localized product banner string
func ProductBanner() string {
	return Sprintf("sqlcmd: Install/Create/Query SQL Server, Azure SQL, and Tools")
}
