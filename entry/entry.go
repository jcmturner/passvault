package entry

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

type Entry struct {
	Name      string
	Salt      []byte `json:"-"`
	Encrypted []byte `json:"-"`
	Secret    Secret
}

type Secret struct {
	Username   string
	Password   string
	Notes      string
	Properties map[string]string
}

func New(name string) Entry {
	return Entry{
		Name: name,
		Secret: Secret{
			Properties: make(map[string]string),
		},
	}
}

func (e Entry) Encrypt(passphrase []byte) ([]byte, error) {
	key, salt := DeriveKey(passphrase, nil)
	e.Salt = salt

	// Marshal the secret into bytes
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(e.Secret)
	if err != nil {
		return []byte{}, err
	}

	// Encrypt
	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(c)
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return []byte{}, err
	}
	e.Encrypted = gcm.Seal(nonce, nonce, buf.Bytes(), nil)

	// Clear secret
	e.Secret = Secret{}
	// Marshal the entry into bytes
	buf = new(bytes.Buffer)
	enc = gob.NewEncoder(buf)
	err = enc.Encode(e)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func (e Entry) String() string {
	return fmt.Sprintf(
		`Name: %s
Username: %s
Password: %s`,
		e.Name, e.Secret.Username, e.Secret.Password)
}

func (e Entry) Redacted() string {
	return fmt.Sprintf(
		`Name: %s
Username: %s
Password: ****`,
		e.Name, e.Secret.Username)
}

func (e Entry) JSON() string {
	jb, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"Error": "Could not marshal entry to JSON: %v"}`, err)
	}
	return string(jb)
}

func (e Entry) AddProperty(k string, v string) {
	e.Secret.Properties[k] = v
}

func (e Entry) GetProperty(k string) (string, bool) {
	v, ok := e.Secret.Properties[k]
	return v, ok
}

func Decrypt(b, passphrase []byte) (Entry, error) {
	var e Entry
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&e)
	if err != nil {
		return e, fmt.Errorf("malformed entry: %v", err)
	}
	key, _ := DeriveKey(passphrase, e.Salt)

	c, err := aes.NewCipher(key)
	if err != nil {
		return e, err
	}
	gcm, err := cipher.NewGCM(c)
	nonceSize := gcm.NonceSize()
	if len(e.Encrypted) < nonceSize {
		return e, errors.New("invalid encrypted value")
	}
	nonce, ct := e.Encrypted[:nonceSize], e.Encrypted[nonceSize:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return e, err
	}
	s := Secret{}
	buf = bytes.NewBuffer(pt)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&s)
	if err != nil {
		return e, fmt.Errorf("malformed entry: %v", err)
	}
	e.Secret = s
	return e, nil
}

func DeriveKey(passphrase, salt []byte) ([]byte, []byte) {
	if salt == nil {
		salt = make([]byte, 8)
		rand.Read(salt)
	}
	return pbkdf2.Key(passphrase, salt, 1000, 32, sha256.New), salt
}
