package store

import (
	"github.com/jcmturner/passvault/entry"
	"github.com/prologic/bitcask"
)

func Put(name, username, password string, masterpasswd []byte, cask *bitcask.Bitcask) error {
	// Create entry and encrypt
	e := entry.New(name)
	e.Secret.Username = username
	e.Secret.Password = password
	b, err := e.Encrypt([]byte(masterpasswd))
	if err != nil {
		return err
	}

	// Put to the store
	err = cask.Put(name, b)
	if err != nil {
		return err
	}
	return cask.Sync()
}

func Get(name string, masterpasswd []byte, cask *bitcask.Bitcask) (entry.Entry, error) {
	b, err := cask.Get(name)
	if err != nil {
		return entry.Entry{}, err
	}
	return entry.Decrypt(b, []byte(masterpasswd))
}
