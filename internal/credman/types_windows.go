// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package credman

import (
	syscall "golang.org/x/sys/windows"
	"time"
)

const (
	CredTypeGeneric CredentialType = 0x1
)

// Credential is the basic credential structure.
// A credential is identified by its target name.
// The actual credential secret is available in the CredentialBlob field.
type Credential struct {
	TargetName     string
	Comment        string
	LastWritten    time.Time
	CredentialBlob []byte
	TargetAlias    string
	UserName       string
	Persist        CredentialPersistence
}

// CredentialPersistence describes one of three persistence modes of a credential.
// A detailed description of the available modes can be found on:
//
//	https://learn.microsoft.com/en-us/windows/win32/api/wincred/ns-wincred-credentiala
type CredentialPersistence uint32

const (
	// PersistSession indicates that the credential only persists for the life
	// of the current Windows login session. Such a credential is not visible in
	// any other logon session, even from the same user.
	PersistSession CredentialPersistence = 0x1
)

// CredentialAttribute represents an application-specific attribute of a credential.
type CredentialAttribute struct {
	Keyword string
	Value   []byte
}

// Interface for syscall.Proc
type proc interface {
	Call(a ...uintptr) (r1, r2 uintptr, lastErr error)
}

type CREDENTIAL struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     uintptr
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

type CREDENTIAL_ATTRIBUTE struct {
	Keyword   *uint16
	Flags     uint32
	ValueSize uint32
	Value     uintptr
}

type CredentialType uint32
