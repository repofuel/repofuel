package codec

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/repofuel/repofuel/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

var gcm, _ = credentials.RandomCipherGCM()
var reg = NewRegistryWithEncryption(gcm).Build()

func TestCredentialsCodec_Interface(t *testing.T) {
	type testStruct struct {
		Cred1 credentials.Interface
		Cred2 credentials.Interface
		Cred3 credentials.Interface
	}

	cred := credentials.BasicAuth{
		Username: "me",
		Password: "pass",
	}

	org := &testStruct{
		Cred1: cred,
		Cred2: &cred,
		Cred3: "to secret",
	}

	b, err := bson.MarshalWithRegistry(reg, org)
	if err != nil {
		t.Fatal(err)
	}

	var doc testStruct
	err = bson.UnmarshalWithRegistry(reg, b, &doc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, org, org)
}

func TestCredentialsCodec_String(t *testing.T) {
	type testStruct struct {
		Cred credentials.String
	}

	b, err := bson.MarshalWithRegistry(reg, &testStruct{
		Cred: "top secret",
	})
	if err != nil {
		t.Fatal(err)
	}

	var doc testStruct
	err = bson.UnmarshalWithRegistry(reg, b, &doc)
	if err != nil {
		t.Fatal(err)
	}

	if doc.Cred != "top secret" {
		t.Fatal("unexpected value:", doc.Cred)
	}
}

func TestCredentialsCodec_Struct(t *testing.T) {
	type testStruct struct {
		Cred *credentials.BasicAuth
	}

	ba := &credentials.BasicAuth{
		Username: "MyUserName",
		Password: "MyPassword",
	}

	b, err := bson.MarshalWithRegistry(reg, &testStruct{
		Cred: ba,
	})
	if err != nil {
		t.Fatal(err)
	}

	var doc testStruct
	err = bson.UnmarshalWithRegistry(reg, b, &doc)
	if err != nil {
		t.Fatal(err)
	}

	if *doc.Cred != *ba {
		t.Fatal("unexpected value:", doc.Cred)
	}
}

func TestCredentialsCodec_Bytes(t *testing.T) {
	type testStruct struct {
		Cred credentials.Bytes
	}

	original := testStruct{
		Cred: []byte("top secret"),
	}

	b, err := bson.MarshalWithRegistry(reg, original)
	if err != nil {
		t.Fatal(err)
	}

	var doc testStruct
	err = bson.UnmarshalWithRegistry(reg, b, &doc)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(doc.Cred, original.Cred) {
		t.Fatal("unexpected value:", doc.Cred)
	}
}

type testInterface interface{}

type testConfig1 struct {
	F1 string
	F2 int64
}

func (t testConfig1) String() string {
	return "T1"
}

type testConfig2 struct {
	F1 float64
	F2 []byte
}

func (t testConfig2) String() string {
	return "T2"
}

func TestInterfaceCodec(t *testing.T) {
	interfaceCodec := NewInterfaceCodec("String",
		&testConfig1{},
		&testConfig2{},
	)
	tInterface := reflect.TypeOf((*testInterface)(nil)).Elem()

	reg := bson.NewRegistryBuilder().
		RegisterTypeDecoder(tInterface, interfaceCodec).
		RegisterTypeEncoder(tInterface, interfaceCodec).
		Build()

	type testStruct struct {
		First   string
		Config1 testInterface
		Config2 testInterface
		Name    string
	}

	original := testStruct{
		First: "First",
		Config1: &testConfig1{
			F1: "Test text",
			F2: 150,
		},
		Config2: &testConfig2{
			F1: 22.2,
		},
		Name: "Test",
	}

	b, err := bson.MarshalWithRegistry(reg, original)
	if err != nil {
		t.Fatal(err)
	}

	var doc testStruct
	err = bson.UnmarshalWithRegistry(reg, b, &doc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, original, doc)
}
