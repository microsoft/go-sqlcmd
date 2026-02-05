// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build windows

package sqlcmd

import (
	"syscall"
	"unsafe"

	"golang.org/x/text/language"
)

var (
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultLCID = kernel32.NewProc("GetUserDefaultLCID")
)

// detectUserLocale returns the user's locale from Windows settings
func detectUserLocale() language.Tag {
	// Get user default locale
	ret, _, _ := procGetUserDefaultLCID.Call()
	lcid := uint32(ret)
	locale := lcidToLanguageTag(lcid)
	if tag, err := language.Parse(locale); err == nil {
		return tag
	}
	return language.English
}

// suppressUnused is used to prevent "imported and not used" errors
var _ = unsafe.Sizeof(0)

// lcidToLanguageTag converts a Windows LCID to a BCP 47 language tag
func lcidToLanguageTag(lcid uint32) string {
	// Common LCID mappings
	// See: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-lcid
	switch lcid {
	case 0x0409:
		return "en-US"
	case 0x0809:
		return "en-GB"
	case 0x0c09:
		return "en-AU"
	case 0x1009:
		return "en-CA"
	case 0x0407:
		return "de-DE"
	case 0x0807:
		return "de-CH"
	case 0x0c07:
		return "de-AT"
	case 0x040c:
		return "fr-FR"
	case 0x080c:
		return "fr-BE"
	case 0x0c0c:
		return "fr-CA"
	case 0x100c:
		return "fr-CH"
	case 0x0410:
		return "it-IT"
	case 0x0810:
		return "it-CH"
	case 0x0c0a:
		return "es-ES"
	case 0x080a:
		return "es-MX"
	case 0x2c0a:
		return "es-AR"
	case 0x0416:
		return "pt-BR"
	case 0x0816:
		return "pt-PT"
	case 0x0413:
		return "nl-NL"
	case 0x0813:
		return "nl-BE"
	case 0x0419:
		return "ru-RU"
	case 0x0415:
		return "pl-PL"
	case 0x0405:
		return "cs-CZ"
	case 0x041b:
		return "sk-SK"
	case 0x040e:
		return "hu-HU"
	case 0x0418:
		return "ro-RO"
	case 0x0402:
		return "bg-BG"
	case 0x041a:
		return "hr-HR"
	case 0x0424:
		return "sl-SI"
	case 0x0c1a:
		return "sr-Latn-RS"
	case 0x081a:
		return "sr-Cyrl-RS"
	case 0x041f:
		return "tr-TR"
	case 0x0408:
		return "el-GR"
	case 0x0422:
		return "uk-UA"
	case 0x0423:
		return "be-BY"
	case 0x040b:
		return "fi-FI"
	case 0x041d:
		return "sv-SE"
	case 0x0414:
		return "nb-NO"
	case 0x0814:
		return "nn-NO"
	case 0x0406:
		return "da-DK"
	case 0x040f:
		return "is-IS"
	case 0x0411:
		return "ja-JP"
	case 0x0412:
		return "ko-KR"
	case 0x0804:
		return "zh-CN"
	case 0x0404:
		return "zh-TW"
	case 0x0c04:
		return "zh-HK"
	default:
		// Default to US English
		return "en-US"
	}
}
