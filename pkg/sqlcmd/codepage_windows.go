// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build windows

package sqlcmd

import (
	"errors"
	"strconv"
	"unicode/utf16"
	"unsafe"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

var (
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procMultiByteToWideChar = kernel32.NewProc("MultiByteToWideChar")
	procWideCharToMultiByte = kernel32.NewProc("WideCharToMultiByte")
)

// windowsCodePageEncoding implements encoding.Encoding using Windows API
type windowsCodePageEncoding struct {
	codepage uint32
}

func (e *windowsCodePageEncoding) NewDecoder() *encoding.Decoder {
	return &encoding.Decoder{Transformer: &windowsDecoder{codepage: e.codepage}}
}

func (e *windowsCodePageEncoding) NewEncoder() *encoding.Encoder {
	return &encoding.Encoder{Transformer: &windowsEncoder{codepage: e.codepage}}
}

// windowsDecoder converts from a Windows codepage to UTF-8
type windowsDecoder struct {
	codepage uint32
}

func (d *windowsDecoder) Reset() {}

func (d *windowsDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if len(src) == 0 {
		return 0, 0, nil
	}

	// First call to get required buffer size for wide chars
	n, _, errno := procMultiByteToWideChar.Call(
		uintptr(d.codepage),
		0,
		uintptr(unsafe.Pointer(&src[0])),
		uintptr(len(src)),
		0,
		0,
	)
	if n == 0 {
		if errno != windows.ERROR_SUCCESS {
			return 0, 0, errno
		}
		return 0, 0, errors.New("MultiByteToWideChar failed")
	}

	// Allocate wide char buffer
	wideChars := make([]uint16, n)

	// Convert to wide chars
	n, _, errno = procMultiByteToWideChar.Call(
		uintptr(d.codepage),
		0,
		uintptr(unsafe.Pointer(&src[0])),
		uintptr(len(src)),
		uintptr(unsafe.Pointer(&wideChars[0])),
		uintptr(len(wideChars)),
	)
	if n == 0 {
		if errno != windows.ERROR_SUCCESS {
			return 0, 0, errno
		}
		return 0, 0, errors.New("MultiByteToWideChar failed")
	}

	// Convert UTF-16 to UTF-8
	runes := utf16.Decode(wideChars[:n])
	utf8Bytes := []byte(string(runes))

	if len(utf8Bytes) > len(dst) {
		return 0, 0, transform.ErrShortDst
	}

	copy(dst, utf8Bytes)
	return len(utf8Bytes), len(src), nil
}

// windowsEncoder converts from UTF-8 to a Windows codepage
type windowsEncoder struct {
	codepage uint32
}

func (e *windowsEncoder) Reset() {}

func (e *windowsEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if len(src) == 0 {
		return 0, 0, nil
	}

	// Convert UTF-8 to UTF-16
	runes := []rune(string(src))
	wideChars := utf16.Encode(runes)

	if len(wideChars) == 0 {
		return 0, len(src), nil
	}

	// First call to get required buffer size
	n, _, errno := procWideCharToMultiByte.Call(
		uintptr(e.codepage),
		0,
		uintptr(unsafe.Pointer(&wideChars[0])),
		uintptr(len(wideChars)),
		0,
		0,
		0,
		0,
	)
	if n == 0 {
		if errno != windows.ERROR_SUCCESS {
			return 0, 0, errno
		}
		return 0, 0, errors.New("WideCharToMultiByte failed")
	}

	if int(n) > len(dst) {
		return 0, 0, transform.ErrShortDst
	}

	// Convert to multibyte
	n, _, errno = procWideCharToMultiByte.Call(
		uintptr(e.codepage),
		0,
		uintptr(unsafe.Pointer(&wideChars[0])),
		uintptr(len(wideChars)),
		uintptr(unsafe.Pointer(&dst[0])),
		uintptr(len(dst)),
		0,
		0,
	)
	if n == 0 {
		if errno != windows.ERROR_SUCCESS {
			return 0, 0, errno
		}
		return 0, 0, errors.New("WideCharToMultiByte failed")
	}

	return int(n), len(src), nil
}

// isCodePageValid checks if a codepage is valid/installed on Windows
func isCodePageValid(codepage uint32) bool {
	// Try to convert a simple byte - if the codepage is invalid, this will fail
	src := []byte{0x41} // 'A'
	n, _, _ := procMultiByteToWideChar.Call(
		uintptr(codepage),
		0,
		uintptr(unsafe.Pointer(&src[0])),
		1,
		0,
		0,
	)
	return n > 0
}

// getSystemCodePageEncoding returns an encoding using Windows API for codepages
// not in our built-in registry. Returns nil if the codepage is not available.
func getSystemCodePageEncoding(codepage int) (encoding.Encoding, error) {
	cp := uint32(codepage)
	if !isCodePageValid(cp) {
		// Use %s with strconv.Itoa to avoid locale-based number formatting
		// that would add thousands separators (e.g., "99,999" instead of "99999")
		return nil, localizer.Errorf("unsupported codepage %s", strconv.Itoa(codepage))
	}
	return &windowsCodePageEncoding{codepage: cp}, nil
}
