package keys

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const pemStart = "-----BEGIN "

var (
	ErrKeyMustBePEMEncoded = errors.New("key must be PEM encoded PKCS1 or PKCS8 private key")
	ErrNotRSAPrivateKey    = errors.New("key is not a valid RSA private key")
)

type ServiceKeys struct {
	PrivateKey    *ecdsa.PrivateKey
	PublicKeys    map[string]crypto.PublicKey
	EncryptionKey cipher.AEAD
}

// UnmarshalYAML check at the first if the value existed in the
// environment variables, if omitted, then it check the value in the
// YAML. The value could be a pem string or path for a pem file.
func (kys *ServiceKeys) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var raw struct {
		PrivateKey    string            `yaml:"private_key"`
		PublicKeys    map[string]string `yaml:"public_keys"`
		EncryptionKey string            `yaml:"encryption_key"`
	}

	if err = unmarshal(&raw); err != nil {
		return err
	}

	if raw.PrivateKey != "" {
		if kys.PrivateKey, err = LoadECPrivateKey(raw.PrivateKey); err != nil {
			return err
		}
	}

	if kys.PublicKeys == nil {
		kys.PublicKeys = make(map[string]crypto.PublicKey, len(raw.PublicKeys))
	}
	for k := range raw.PublicKeys {

		v, ok := os.LookupEnv("PUBLIC_KEY_SRV_" + strings.ToUpper(k))
		if !ok {
			v = raw.PublicKeys[k]
		}

		kys.PublicKeys[k], err = LoadPKIXPublicKey(v)
		if err != nil {
			return err
		}
	}

	if v, ok := os.LookupEnv("ENCRYPTION_KEY"); ok {
		raw.EncryptionKey = v
	}

	if raw.EncryptionKey != "" {
		c, err := aes.NewCipher([]byte(raw.EncryptionKey))
		if err != nil {
			return err
		}

		kys.EncryptionKey, err = cipher.NewGCM(c)
		if err != nil {
			return err
		}
	}

	return nil
}

type RSAPrivateKey struct {
	key *rsa.PrivateKey
}

func (k *RSAPrivateKey) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	k.key, err = LoadRSAPrivateKey(v)
	if err != nil {
		return err
	}

	return nil
}

func (k *RSAPrivateKey) Load(file string) error {
	key, err := LoadRSAPrivateKey(file)
	if err != nil {
		return err
	}

	k.key = key

	return nil
}

func (k *RSAPrivateKey) Key() *rsa.PrivateKey {
	return k.key
}

func getPem(file string) (*pem.Block, error) {
	var key []byte
	// if file is a path, read the file
	if !strings.HasPrefix(file, pemStart) {
		var err error
		if key, err = ioutil.ReadFile(file); err != nil {
			return nil, fmt.Errorf("cannot read the key: %w", err)
		}
	} else {
		key = []byte(file)
	}

	block, _ := pem.Decode(key)
	if block == nil {
		return nil, ErrKeyMustBePEMEncoded
	}
	return block, nil
}

func GenKey() {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey := &privateKey.PublicKey
	a, b := encode(privateKey, publicKey)
	print(a, "\n\n\n", b)
}

func encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncoded), string(pemEncodedPub)
}

func decode(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return privateKey, publicKey
}

func LoadECPrivateKey(file string) (*ecdsa.PrivateKey, error) {
	block, err := getPem(file)
	if err != nil {
		return nil, err
	}

	return x509.ParseECPrivateKey(block.Bytes)
}

func LoadPKIXPublicKey(file string) (crypto.PublicKey, error) {
	block, err := getPem(file)
	if err != nil {
		return nil, err
	}

	return x509.ParsePKIXPublicKey(block.Bytes)
}

func LoadRSAPrivateKey(filename string) (*rsa.PrivateKey, error) {

	block, err := getPem(filename)
	if err != nil {
		return nil, ErrKeyMustBePEMEncoded
	}

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, err
		}
	}

	pkey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, ErrNotRSAPrivateKey
	}

	return pkey, nil
}
