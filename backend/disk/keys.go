package disk

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/fazilb93/neutron/backend"
)

// Stores private & public keys on disk.
// Public keys are stored in DOMAIN/USERNAME.pub.gpg and private keys are
// in DOMAIN/USERNAME.priv.gpg.
type Keys struct {
	config *Config
	users  backend.UsersBackend
}

func parseEmail(email string) (username, domain string) {
	parts := strings.SplitN(email, "@", 2)
	username = parts[0]
	domain = parts[1]
	return
}

func (b *Keys) getKeyPath(email string, priv bool) (path string) {
	username, domain := parseEmail(email)

	path = b.config.Directory + "/" + domain + "/" + username
	if priv {
		path += ".priv"
	} else {
		path += ".pub"
	}
	path += ".gpg"
	return
}

func (b *Keys) getKey(email string, priv bool) (key string, err error) {
	path := b.getKeyPath(email, priv)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	key = string(data)
	return
}

func (b *Keys) GetPublicKey(email string) (string, error) {
	key, err := b.getKey(email, false)
	if os.IsNotExist(err) {
		return "", nil
	}
	return key, err
}

func (b *Keys) GetKeypair(email string) (keypair *backend.Keypair, err error) {
	pub, err := b.getKey(email, false)
	if err != nil {
		return
	}

	// TODO: use password to encrypt private key

	priv, err := b.getKey(email, true)
	if err != nil {
		return
	}

	keypair = &backend.Keypair{
		ID:         email,
		PublicKey:  pub,
		PrivateKey: priv,
	}
	return
}

func (b *Keys) InsertKeypair(email string, keypair *backend.Keypair) (*backend.Keypair, error) {
	return b.UpdateKeypair(email, keypair)
}

func (b *Keys) UpdateKeypair(email string, keypair *backend.Keypair) (updated *backend.Keypair, err error) {
	if keypair.ID == "" {
		keypair.ID = email
	}

	_, domain := parseEmail(email)
	parentPath := b.config.Directory + "/" + domain
	err = os.MkdirAll(parentPath, 0744)
	if err != nil {
		return
	}

	if keypair.PublicKey != "" {
		pubPath := b.getKeyPath(email, false)
		err = ioutil.WriteFile(pubPath, []byte(keypair.PublicKey), 0644)
		if err != nil {
			return
		}
	}

	privPath := b.getKeyPath(email, true)
	err = ioutil.WriteFile(privPath, []byte(keypair.PrivateKey), 0644)
	if err != nil {
		return
	}

	updated = keypair
	return
}

func NewKeys(config *Config, users backend.UsersBackend) backend.KeysBackend {
	return &Keys{
		config: config,
		users:  users,
	}
}

func UseKeys(bkd *backend.Backend, config *Config) {
	bkd.Set(NewKeys(config, bkd))
}
