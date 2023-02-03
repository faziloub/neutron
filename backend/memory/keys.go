package memory

import (
	"errors"

	"github.com/fazilb93/neutron/backend"
)

type Keys struct {
	keys map[string]*backend.Keypair
}

func (b *Keys) getKeypair(email string) (*backend.Keypair, error) {
	kp, ok := b.keys[email]
	if !ok {
		return nil, errors.New("No such keypair")
	}
	return kp, nil
}

func (b *Keys) GetPublicKey(email string) (string, error) {
	kp, err := b.getKeypair(email)
	if err != nil {
		return "", nil
	}
	return kp.PublicKey, nil
}

func (b *Keys) GetKeypair(email string) (*backend.Keypair, error) {
	kp, err := b.getKeypair(email)
	if err != nil {
		return nil, err
	}
	return kp, nil
}

func (b *Keys) InsertKeypair(email string, keypair *backend.Keypair) (inserted *backend.Keypair, err error) {
	keypair.ID = email
	b.keys[email] = keypair
	inserted = keypair
	return
}

func (b *Keys) UpdateKeypair(email string, keypair *backend.Keypair) (updated *backend.Keypair, err error) {
	updated = b.keys[email]

	if keypair.PublicKey != "" {
		updated.PublicKey = keypair.PublicKey
	}
	updated.PrivateKey = keypair.PrivateKey

	return
}

func NewKeys() backend.KeysBackend {
	return &Keys{
		keys: map[string]*backend.Keypair{},
	}
}
