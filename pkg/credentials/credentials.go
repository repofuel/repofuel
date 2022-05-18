package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type credType uint8

const (
	typeBytes credType = iota + 1
	typePlainText
	typeBasicAuth
	typeToken
	typePrivateKey
)

type Interface interface{}

type String string

type Bytes []byte

type Token struct {
	Token       string `json:"token"`
	TokenSecret string `json:"secret"`
}

type BasicAuth struct {
	Username string `json:"user"`
	Password string `json:"pass"`
}

type EncryptedCred struct {
	Type  credType `bson:"t"`
	Bytes []byte   `bson:"b"`
}

func (EncryptedCred) String() string {
	return "*******"
}

func NewEncryptedCred(gcm cipher.AEAD, plain interface{}) (*EncryptedCred, error) {
	var cred EncryptedCred
	var err error
	var b []byte

	switch plain := plain.(type) {
	case BasicAuth, *BasicAuth:
		cred.Type = typeBasicAuth
		b, err = json.Marshal(plain)
		if err != nil {
			return nil, err
		}

	case Token, *Token:
		cred.Type = typeToken
		b, err = json.Marshal(plain)
		if err != nil {
			return nil, err
		}

	case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
		cred.Type = typePrivateKey
		b, err = x509.MarshalPKCS8PrivateKey(plain)
		if err != nil {
			return nil, err
		}

	case string:
		cred.Type = typePlainText
		b = []byte(plain)

	case String:
		cred.Type = typePlainText
		b = []byte(plain)

	case []byte:
		cred.Type = typeBytes
		b = plain

	case Bytes:
		cred.Type = typeBytes
		b = plain

	default:
		return nil, fmt.Errorf("unsupported credential type: %T", plain)
	}

	cred.Bytes, err = encrypt(gcm, b)
	if err != nil {
		return nil, err
	}

	return &cred, nil
}

func encrypt(gcm cipher.AEAD, b []byte) ([]byte, error) {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, b, nil), nil
}

func (auth *EncryptedCred) Decrypt(gcm cipher.AEAD) (interface{}, error) {
	decrypted, err := decrypt(gcm, auth.Bytes)
	if err != nil {
		return nil, err
	}

	var r interface{}
	switch auth.Type {
	case typePlainText:
		return String(decrypted), nil
	case typeBytes:
		return Bytes(decrypted), nil
	case typePrivateKey:
		return x509.ParsePKCS8PrivateKey(decrypted)
	case typeBasicAuth:
		r = new(BasicAuth)
	case typeToken:
		r = new(Token)
	default:
		return nil, errors.New("unsupported credential type")
	}

	err = json.Unmarshal(decrypted, &r)

	return r, err
}

func decrypt(gcm cipher.AEAD, b []byte) ([]byte, error) {
	nonceSize := gcm.NonceSize()
	if len(b) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := b[:nonceSize], b[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func RandomCipherGCM() (cipher.AEAD, error) {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(c)
}
