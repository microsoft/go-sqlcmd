// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build windows

package sqlcmd

import (
	"errors"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

const (
	// MB_ERR_INVALID_CHARS causes MultiByteToWideChar to fail if it encounters
	// an invalid character in the source string (including incomplete sequences)
	mbErrInvalidChars = 0x00000008
	// Maximum bytes that might form a single character in any Windows codepage
	// (most DBCS codepages use 2 bytes, but we use 4 for safety)
	maxMultibyteCharLen = 4
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

// windowsDecoder converts from a Windows codepage to UTF-8.
// It buffers incomplete multibyte sequences between Transform calls.
type windowsDecoder struct {
	codepage uint32
	buf      [maxMultibyteCharLen]byte // buffer for incomplete sequences
	bufLen   int                       // number of bytes in buffer
}

func (d *windowsDecoder) Reset() {
	d.bufLen = 0
}

func (d *windowsDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// Prepend any buffered bytes from previous call
	var input []byte
	if d.bufLen > 0 {
		input = make([]byte, d.bufLen+len(src))
		copy(input, d.buf[:d.bufLen])
		copy(input[d.bufLen:], src)
	} else {
		input = src
	}

	if len(input) == 0 {
		return 0, 0, nil
	}

	// Try to convert with MB_ERR_INVALID_CHARS to detect incomplete sequences
	n, _, errno := procMultiByteToWideChar.Call(
		uintptr(d.codepage),
		mbErrInvalidChars,
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(len(input)),
		0,
		0,
	)

	// If conversion failed, it might be due to incomplete trailing sequence
	if n == 0 && errno == windows.ERROR_NO_UNICODE_TRANSLATION {
		if atEOF {
			// At EOF with incomplete sequence - this is an error
			d.bufLen = 0
			return 0, len(src), errors.New("incomplete multibyte sequence at end of input")
		}

		// Not at EOF - try removing bytes from the end until conversion succeeds
		// This finds the incomplete trailing sequence
		for trimLen := 1; trimLen <= len(input) && trimLen <= maxMultibyteCharLen; trimLen++ {
			tryLen := len(input) - trimLen
			if tryLen <= 0 {
				// Need more input - buffer what we have
				if len(input) <= maxMultibyteCharLen {
					copy(d.buf[:], input)
					d.bufLen = len(input)
					return 0, len(src), transform.ErrShortSrc
				}
				break
			}

			n, _, errno = procMultiByteToWideChar.Call(
				uintptr(d.codepage),
				mbErrInvalidChars,
				uintptr(unsafe.Pointer(&input[0])),
				uintptr(tryLen),
				0,
				0,
			)
			if n > 0 || errno != windows.ERROR_NO_UNICODE_TRANSLATION {
				// Found a valid prefix - buffer the trailing bytes
				trailingBytes := input[tryLen:]
				copy(d.buf[:], trailingBytes)
				d.bufLen = len(trailingBytes)
				input = input[:tryLen]
				break
			}
		}

		// If still failing, buffer everything and wait for more
		if n == 0 {
			if len(input) <= maxMultibyteCharLen {
				copy(d.buf[:], input)
				d.bufLen = len(input)
				return 0, len(src), transform.ErrShortSrc
			}
			// Input is larger than max char length but still invalid - real error
			d.bufLen = 0
			return 0, len(src), errors.New("invalid multibyte sequence")
		}
	} else if n == 0 {
		if errno != windows.ERROR_SUCCESS {
			d.bufLen = 0
			return 0, 0, errno
		}
		d.bufLen = 0
		return 0, 0, errors.New("MultiByteToWideChar failed")
	} else {
		// Success - clear buffer since we'll consume all input
		d.bufLen = 0
	}

	// Allocate wide char buffer and do the actual conversion
	wideChars := make([]uint16, n)
	n, _, errno = procMultiByteToWideChar.Call(
		uintptr(d.codepage),
		0, // Don't use MB_ERR_INVALID_CHARS here - we already validated
		uintptr(unsafe.Pointer(&input[0])),
		uintptr(len(input)),
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
	return len(utf8Bytes), len(src), err
}

// windowsEncoder converts from UTF-8 to a Windows codepage.
// It buffers incomplete UTF-8 sequences between Transform calls.
type windowsEncoder struct {
	codepage uint32
	buf      [utf8.UTFMax]byte // buffer for incomplete UTF-8 sequences
	bufLen   int               // number of bytes in buffer
}

func (e *windowsEncoder) Reset() {
	e.bufLen = 0
}

func (e *windowsEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// Prepend any buffered bytes from previous call
	var input []byte
	if e.bufLen > 0 {
		input = make([]byte, e.bufLen+len(src))
		copy(input, e.buf[:e.bufLen])
		copy(input[e.bufLen:], src)
	} else {
		input = src
	}

	if len(input) == 0 {
		return 0, 0, nil
	}

	// Find the last complete UTF-8 sequence
	validLen := len(input)
	for validLen > 0 && !utf8.Valid(input[:validLen]) {
		validLen--
	}

	// Check for incomplete trailing sequence
	if validLen < len(input) {
		trailingBytes := input[validLen:]
		if atEOF {
			// At EOF with incomplete UTF-8 - this is an error
			e.bufLen = 0
			return 0, len(src), errors.New("incomplete UTF-8 sequence at end of input")
		}
		// Buffer the incomplete trailing bytes for next call
		if len(trailingBytes) <= utf8.UTFMax {
			copy(e.buf[:], trailingBytes)
			e.bufLen = len(trailingBytes)
		} else {
			// Shouldn't happen with valid partial UTF-8, but handle it
			e.bufLen = 0
			return 0, len(src), errors.New("invalid UTF-8 sequence")
		}
		input = input[:validLen]
	} else {
		e.bufLen = 0
	}

	if len(input) == 0 {
		// Only incomplete sequence - need more input
		return 0, len(src), transform.ErrShortSrc
	}

	// Convert UTF-8 to UTF-16
	runes := []rune(string(input))
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

	return int(n), len(src), err
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
// not in our built-in registry. If the codepage is not available, it returns
// a nil encoding and a non-nil error.
func getSystemCodePageEncoding(codepage int) (encoding.Encoding, error) {
	cp := uint32(codepage)
	if !isCodePageValid(cp) {
		return nil, localizer.Errorf("unsupported codepage %s", strconv.Itoa(codepage))
	}
	return &windowsCodePageEncoding{codepage: cp}, nil
}
