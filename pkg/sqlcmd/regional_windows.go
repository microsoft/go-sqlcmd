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
	procLCIDToLocaleName   = kernel32.NewProc("LCIDToLocaleName")
)

const localeNameMaxLength = 85

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

// lcidToLanguageTag converts a Windows LCID to a BCP 47 language tag
func lcidToLanguageTag(lcid uint32) string {
	buffer := make([]uint16, localeNameMaxLength)
	ret, _, _ := procLCIDToLocaleName.Call(
		uintptr(lcid),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		0,
	)
	if ret == 0 {
		return "en-US"
	}
	return syscall.UTF16ToString(buffer)
}
