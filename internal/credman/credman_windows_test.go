package credman

import (
	"github.com/stretchr/testify/assert"
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
	assert.NoErrorf(t, err, "WriteCredential returned an error: %v", err)

	// Check if the written credential can be retrieved
	credentials, _ := EnumerateCredentials("TestTargetName", false)
	assert.NotEqual(t, len(credentials), 0, "WriteCredential failed to write the credential")
	assert.Equal(t, credentials[0].TargetName, "TestTargetName", "WriteCredential wrote incorrect TargetName, got: %s, want: %s", credentials[0].TargetName, "TestTargetName")
	assert.Equal(t, credentials[0].Comment, "TestComment", "WriteCredential wrote incorrect Comment, got: %s, want: %s", credentials[0].Comment, "TestComment")
	assert.Equal(t, credentials[0].UserName, "TestUsername", "WriteCredential wrote incorrect UserName, got: %s, want: %s", credentials[0].UserName, "TestUsername")
	assert.Equal(t, credentials[0].Persist, PersistSession, "WriteCredential wrote incorrect Persist, got: %d, want: %d", credentials[0].Persist, PersistSession)

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
	assert.NoErrorf(t, err, "DeleteCredential returned an error: %v", err)

	// Check if the deleted credential still exists
	credentials, _ := EnumerateCredentials("TestTargetName", false)
	assert.Equal(t, len(credentials), 0, "DeleteCredential failed to delete the credential")
}

func TestNegDeleteCredential(t *testing.T) {
	credential := &Credential{}

	err := DeleteCredential(credential, CredTypeGeneric)
	assert.Errorf(t, err, "err should not be nil")
}

func TestNegWriteCredential(t *testing.T) {
	credential := &Credential{}

	err := WriteCredential(credential, CredTypeGeneric)
	assert.Errorf(t, err, "err should not be nil")
}

func TestNegConvertFromSystemCredential(t *testing.T) {
	result := convertFromSystemCredential(nil)
	assert.Nil(t, result, "result should be nil")
}

func TestNegConvertToSystemCredential(t *testing.T) {
	result := convertToSystemCredential(nil)
	assert.Nil(t, result, "result should be nil")
}

func TestNegcopyBytesToSlice(t *testing.T) {
	b := copyBytesToSlice(uintptr(0), 0)
	assert.Len(t, b, 0, "bytes should be empty")
}
