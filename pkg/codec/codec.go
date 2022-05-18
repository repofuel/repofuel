package codec

import (
	"crypto/cipher"
	"errors"
	"reflect"

	"github.com/repofuel/repofuel/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var tEncryptedCred = reflect.TypeOf((*credentials.EncryptedCred)(nil)).Elem()

func NewRegistryWithEncryption(gcm cipher.AEAD) *bsoncodec.RegistryBuilder {
	rb := bson.NewRegistryBuilder()
	NewCredentialsCodec(gcm).RegisterCredentialsCodecs(rb)
	return rb
}

type CredentialsCodec struct {
	gcm cipher.AEAD
}

func NewCredentialsCodec(gcm cipher.AEAD) *CredentialsCodec {
	return &CredentialsCodec{
		gcm: gcm,
	}
}
func (c *CredentialsCodec) RegisterCredentialsCodecs(rb *bsoncodec.RegistryBuilder) {
	if rb == nil {
		panic(errors.New("argument to RegisterPrimitiveCodecs must not be nil"))
	}

	var (
		tCredInterface = reflect.TypeOf((*credentials.Interface)(nil)).Elem()
		tCredString    = reflect.TypeOf((*credentials.String)(nil)).Elem()
		tCredBytes     = reflect.TypeOf((*credentials.Bytes)(nil)).Elem()
		tCredToken     = reflect.TypeOf((*credentials.Token)(nil)).Elem()
		tCredBasicAuth = reflect.TypeOf((*credentials.BasicAuth)(nil)).Elem()
	)

	rb.
		RegisterTypeEncoder(tCredInterface, c).
		RegisterTypeEncoder(tCredString, c).
		RegisterTypeEncoder(tCredToken, c).
		RegisterTypeEncoder(tCredBasicAuth, c).
		RegisterTypeEncoder(tCredBytes, c).
		RegisterTypeDecoder(tCredInterface, c).
		RegisterTypeDecoder(tCredString, c).
		RegisterTypeDecoder(tCredToken, c).
		RegisterTypeDecoder(tCredBasicAuth, c).
		RegisterTypeDecoder(tCredBytes, c)
}

func (c *CredentialsCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	var encrypted interface{}
	var err error
	switch val.Kind() {
	case reflect.Interface, reflect.Struct:
		encrypted, err = credentials.NewEncryptedCred(c.gcm, val.Interface())
	case reflect.String:
		encrypted, err = credentials.NewEncryptedCred(c.gcm, val.String())
	case reflect.Slice:
		encrypted, err = credentials.NewEncryptedCred(c.gcm, val.Bytes())
	default:
		return errors.New("unsupported credential kind")
	}

	if err != nil {
		return err
	}

	val = reflect.ValueOf(encrypted)
	ve, err := ec.LookupEncoder(val.Type())
	if err != nil {
		return err
	}
	return ve.EncodeValue(ec, vw, val)
}

func (c *CredentialsCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	vd, err := dc.LookupDecoder(tEncryptedCred)
	if err != nil {
		return err
	}

	cred := &credentials.EncryptedCred{}
	err = vd.DecodeValue(dc, vr, reflect.ValueOf(cred).Elem())
	if err != nil {
		return err
	}

	decrypted, err := cred.Decrypt(c.gcm)
	if err != nil {
		return err
	}

	if val.Kind() == reflect.Interface {
		val.Set(reflect.ValueOf(decrypted))
	} else {
		val.Set(reflect.Indirect(reflect.ValueOf(decrypted)))
	}

	return nil
}

type InterfaceCodec struct {
	typeFuncName string
	types        map[string]reflect.Type
}

func (cd *InterfaceCodec) RegisterInterfaceCodec(rb *bsoncodec.RegistryBuilder, i reflect.Type) {
	if i.Kind() != reflect.Interface {
		panic("InterfaceCodec can be registered only for a interface")
	}

	rb.
		RegisterTypeDecoder(i, cd).
		RegisterTypeEncoder(i, cd)
}

func NewInterfaceCodec(funName string, types ...interface{}) *InterfaceCodec {
	m := make(map[string]reflect.Type, len(types))
	for _, t := range types {
		val := reflect.ValueOf(t)
		name := val.MethodByName(funName).Call(nil)[0].String()
		m[name] = val.Type().Elem()
	}
	return &InterfaceCodec{
		typeFuncName: funName,
		types:        m,
	}
}

func (cd *InterfaceCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if val.IsNil() {
		return vw.WriteNull()
	}

	val = val.Elem()

	doc, err := vw.WriteDocument()
	if err != nil {
		return err
	}

	vw, err = doc.WriteDocumentElement(val.MethodByName(cd.typeFuncName).Call(nil)[0].String())
	if err != nil {
		return err
	}

	ve, err := ec.LookupEncoder(val.Type())
	if err != nil {
		return err
	}

	err = ve.EncodeValue(ec, vw, val)
	if err != nil {
		return err
	}

	return doc.WriteDocumentEnd()
}

func (cd *InterfaceCodec) DecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.EmbeddedDocument {
		return vr.Skip()
	}

	doc, err := vr.ReadDocument()
	if err != nil {
		return err
	}

	name, vr, err := doc.ReadElement()
	if err != nil {
		return err
	}

	t, ok := cd.types[name]
	if !ok {
		return errors.New("unsupported provider")
	}

	impl := reflect.New(t)
	val.Set(impl)

	dec, err := dc.LookupDecoder(impl.Elem().Type())
	if err != nil {
		return err
	}

	err = dec.DecodeValue(dc, vr, impl.Elem())
	if err != nil {
		return err
	}

	name, _, err = doc.ReadElement()
	if err == bsonrw.ErrEOD {
		return nil
	}
	if err != nil {
		return errors.New("unexpected element on an interface wrapper type")
	}

	return err
}
