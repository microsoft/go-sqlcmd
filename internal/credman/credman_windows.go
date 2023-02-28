// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Package credman provides functions for interacting with the Windows Credential
// Manager API, which allows to read, write, and delete saved credentials on
// the local machine.
package credman

import (
	syscall "golang.org/x/sys/windows"
	"reflect"
	"time"
	"unsafe"
)

// Load the Windows DLL "advapi32.dll", then set up Go wrapper functions for
// Windows System APIs: "CredWriteW", "CredDeleteW", "CredFree", and
// "CredEnumerateW", these APIs are related to Windows Credential Management,
// which provides the ability to securely store user credentials, such as usernames
// and passwords, on the local system.
var advapi32 = syscall.NewLazyDLL("advapi32.dll")
var credWrite proc = advapi32.NewProc("CredWriteW")
var credDelete proc = advapi32.NewProc("CredDeleteW")
var credFree proc = advapi32.NewProc("CredFree")
var credEnumerate proc = advapi32.NewProc("CredEnumerateW")

// WriteCredential writes a credential to the Windows Credential Manager
func WriteCredential(credential *Credential, credentialType CredentialType) error {
	systemCredential := convertToSystemCredential(credential)
	systemCredential.Type = uint32(credentialType)
	ret, _, err := credWrite.Call(
		uintptr(unsafe.Pointer(systemCredential)),
		0,
	)
	if ret == 0 {
		return err
	}

	return nil
}

// DeleteCredential deletes a credential from the Windows Credential Manager
func DeleteCredential(credential *Credential, credentialType CredentialType) error {
	targetNamePtr, _ := syscall.UTF16PtrFromString(credential.TargetName)
	ret, _, err := credDelete.Call(
		uintptr(unsafe.Pointer(targetNamePtr)),
		uintptr(credentialType),
		0,
	)
	if ret == 0 {
		return err
	}

	return nil
}

// EnumerateCredentials returns a slice of Credentials from the Windows
// Credential Manager
func EnumerateCredentials(filter string, all bool) ([]*Credential, error) {
	var count int
	var systemCredential uintptr
	var filterPtr *uint16
	if !all {
		filterPtr, _ = syscall.UTF16PtrFromString(filter)
	}
	ret, _, err := credEnumerate.Call(
		uintptr(unsafe.Pointer(filterPtr)),
		0,
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(&systemCredential)),
	)
	if ret == 0 {
		return nil, err
	}
	defer credFree.Call(systemCredential)
	systemCredentials := *(*[]*CREDENTIAL)(unsafe.Pointer(&reflect.SliceHeader{
		Data: systemCredential,
		Len:  count,
		Cap:  count,
	}))
	credentials := make([]*Credential, count, count)
	for i, c := range systemCredentials {
		credentials[i] = convertFromSystemCredential(c)
	}

	return credentials, nil
}

// convertFromSystemCredential converts the given CREDENTIAL struct to
// a more usable structure for golang
func convertFromSystemCredential(cred *CREDENTIAL) (result *Credential) {
	if cred == nil {
		return nil
	}
	result = new(Credential)
	result.Comment = syscall.UTF16PtrToString(cred.Comment)
	result.TargetName = syscall.UTF16PtrToString(cred.TargetName)
	result.TargetAlias = syscall.UTF16PtrToString(cred.TargetAlias)
	result.UserName = syscall.UTF16PtrToString(cred.UserName)
	result.LastWritten = time.Unix(0, cred.LastWritten.Nanoseconds())
	result.Persist = CredentialPersistence(cred.Persist)
	result.CredentialBlob = copyBytesToSlice(cred.CredentialBlob, cred.CredentialBlobSize)
	return result
}

// convertToSystemCredential, converts the given Credential object back to
// a CREDENTIAL struct, which can be used for calling the Windows APIs
func convertToSystemCredential(cred *Credential) (result *CREDENTIAL) {
	if cred == nil {
		return nil
	}
	result = new(CREDENTIAL)
	result.Flags = 0
	result.Type = 0
	result.TargetName, _ = syscall.UTF16PtrFromString(cred.TargetName)
	result.Comment, _ = syscall.UTF16PtrFromString(cred.Comment)
	result.LastWritten = syscall.NsecToFiletime(cred.LastWritten.UnixNano())
	result.CredentialBlobSize = uint32(len(cred.CredentialBlob))
	if len(cred.CredentialBlob) > 0 {
		result.CredentialBlob = uintptr(unsafe.Pointer(&cred.CredentialBlob[0]))
	} else {
		result.CredentialBlob = 0
	}
	result.Persist = uint32(cred.Persist)
	result.TargetAlias, _ = syscall.UTF16PtrFromString(cred.TargetAlias)
	result.UserName, _ = syscall.UTF16PtrFromString(cred.UserName)

	return
}

func copyBytesToSlice(src uintptr, len uint32) (bytes []byte) {
	if src == uintptr(0) {
		return []byte{}
	}
	bytes = make([]byte, len)
	copy(bytes, *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: src,
		Len:  int(len),
		Cap:  int(len),
	})))
	return
}
