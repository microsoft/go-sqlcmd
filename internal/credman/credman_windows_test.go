package credman

import (
	"testing"
	"time"
)

func TestWriteCredential(t *testing.T) {
	credential := &Credential{
		TargetName:     "TestTargetName",
		Comment:        "TestComment",
		LastWritten:    time.Now(),
		UserName:       "TestUsername",
		Persist:        PersistSession,
		CredentialBlob: []byte{0x65, 0x66},
	}

	err := WriteCredential(credential, CredTypeGeneric)
	if err != nil {
		t.Errorf("WriteCredential returned an error: %v", err)
	}

	// Check if the written credential can be retrieved
	credentials, _ := EnumerateCredentials("TestTargetName", false)
	if len(credentials) == 0 {
		t.Errorf("WriteCredential failed to write the credential")
	}
	if credentials[0].TargetName != "TestTargetName" {
		t.Errorf("WriteCredential wrote incorrect TargetName, got: %s, want: %s", credentials[0].TargetName, "TestTargetName")
	}
	if credentials[0].Comment != "TestComment" {
		t.Errorf("WriteCredential wrote incorrect Comment, got: %s, want: %s", credentials[0].Comment, "TestComment")
	}
	if credentials[0].UserName != "TestUsername" {
		t.Errorf("WriteCredential wrote incorrect UserName, got: %s, want: %s", credentials[0].UserName, "TestUsername")
	}
	if credentials[0].Persist != PersistSession {
		t.Errorf("WriteCredential wrote incorrect Persist, got: %d, want: %d", credentials[0].Persist, PersistSession)
	}

	// Cleanup the written credential
	DeleteCredential(credential, CredTypeGeneric)
}

func TestDeleteCredential(t *testing.T) {
	credential := &Credential{
		TargetName:  "TestTargetName",
		Comment:     "TestComment",
		LastWritten: time.Now(),
		UserName:    "TestUsername",
		Persist:     PersistSession,
	}

	// Write the credential first
	WriteCredential(credential, CredTypeGeneric)

	err := DeleteCredential(credential, CredTypeGeneric)
	if err != nil {
		t.Errorf("DeleteCredential returned an error: %v", err)
	}

	// Check if the deleted credential still exists
	credentials, _ := EnumerateCredentials("TestTargetName", false)
	if len(credentials) != 0 {
		t.Errorf("DeleteCredential failed to delete the credential")
	}
}

func TestNegDeleteCredential(t *testing.T) {
	credential := &Credential{}

	err := DeleteCredential(credential, CredTypeGeneric)
	if err == nil {
		t.Errorf("err should not be nil")
	}
}

func TestNegWriteCredential(t *testing.T) {
	credential := &Credential{}

	err := WriteCredential(credential, CredTypeGeneric)
	if err == nil {
		t.Errorf("err should not be nil")
	}
}

func TestNegConvertFromSystemCredential(t *testing.T) {
	result := convertFromSystemCredential(nil)
	if result != nil {
		t.Errorf("result should be nil")
	}
}

func TestNegConvertToSystemCredential(t *testing.T) {
	result := convertToSystemCredential(nil)
	if result != nil {
		t.Errorf("result should be nil")
	}
}

func TestNegcopyBytesToSlice(t *testing.T) {
	b := copyBytesToSlice(uintptr(0), 0)

	if len(b) != 0 {
		t.Errorf("bytes should be empty")
	}
}
