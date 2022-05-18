package credentials

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"
)

var gcm, _ = RandomCipherGCM()

func TestTypeBasicAuth(t *testing.T) {
	row := &BasicAuth{
		Username: "MyUserName",
		Password: "This is the Password",
	}
	encrypted, err := NewEncryptedCred(gcm, row)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := encrypted.Decrypt(gcm)
	if err != nil {
		t.Fatal(err)
	}

	switch decrypted := decrypted.(type) {
	case *BasicAuth:
		if *decrypted != *row {
			t.Fatal("unexpected decrypted value")
		}
	default:
		t.Fatal("unexpected credentials type")
	}
}

func TestTypePlainText(t *testing.T) {
	row := "This is the Secret"

	encrypted, err := NewEncryptedCred(gcm, row)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := encrypted.Decrypt(gcm)
	if err != nil {
		t.Fatal(err)
	}

	switch decrypted := decrypted.(type) {
	case String:
		if decrypted != String(row) {
			t.Fatal("unexpected decrypted value. expected:", row, "go:", decrypted)
		}
	default:
		t.Fatal("unexpected credentials type")
	}
}

func TestTypeBinary(t *testing.T) {
	row := []byte("This is the Secret")

	encrypted, err := NewEncryptedCred(gcm, row)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := encrypted.Decrypt(gcm)
	if err != nil {
		t.Fatal(err)
	}

	switch decrypted := decrypted.(type) {
	case Bytes:
		if !bytes.Equal(decrypted, row) {
			t.Fatal("unexpected decrypted value. \n\texpected:", row, "\n\tgo:      ", decrypted)
		}
	default:
		t.Fatal("unexpected credentials type")
	}
}

func TestTypePrivateKey(t *testing.T) {
	row, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := NewEncryptedCred(gcm, row)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := encrypted.Decrypt(gcm)
	if err != nil {
		t.Fatal(err)
	}

	switch decrypted := decrypted.(type) {
	case *rsa.PrivateKey:
		if !IsPublicKeysEqual(decrypted, row) {
			t.Fatal("unexpected decrypted value. \n\texpected:", row, "\n\tgo:      ", decrypted)
		}
	default:
		t.Fatalf("unexpected credentials type: %T", decrypted)
	}
}

func IsPublicKeysEqual(k1, k2 crypto.PrivateKey) bool {
	b1, err := x509.MarshalPKCS8PrivateKey(k1)
	if err != nil {
		return false
	}

	b2, err := x509.MarshalPKCS8PrivateKey(k2)
	if err != nil {
		return false
	}

	return bytes.Equal(b1, b2)
}
