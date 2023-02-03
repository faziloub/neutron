package disk

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/fazilb93/neutron/backend"
	"github.com/fazilb93/neutron/backend/events"
	"github.com/fazilb93/neutron/backend/util"
)

type Addresses struct {
	config *Config
}

func (b *Addresses) getAddressesPath(user string) string {
	return b.config.Directory + "/" + user + ".json"
}

func (b *Addresses) loadAddresses(user string) (adresses []*backend.Address, err error) {
	data, err := ioutil.ReadFile(b.getAddressesPath(user))
	if os.IsNotExist(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &adresses)
	return
}

func (b *Addresses) saveAddresses(user string, adresses []*backend.Address) (err error) {
	data, err := json.Marshal(adresses)
	if err != nil {
		return
	}

	err = os.MkdirAll(b.config.Directory, 0744)
	if err != nil {
		return
	}

	return ioutil.WriteFile(b.getAddressesPath(user), data, 0644)
}

func (b *Addresses) ListAddresses(user string) ([]*backend.Address, error) {
	return b.loadAddresses(user)
}

func getAddressIndex(adresses []*backend.Address, id string) (int, error) {
	for i, addr := range adresses {
		if addr.ID == id {
			return i, nil
		}
	}

	return -1, errors.New("No such address")
}

func (b *Addresses) GetAddress(user, id string) (*backend.Address, error) {
	addresses, err := b.loadAddresses(user)
	if err != nil {
		return nil, err
	}

	i, err := getAddressIndex(addresses, id)
	if err != nil {
		return nil, err
	}

	return addresses[i], nil
}

func (b *Addresses) InsertAddress(user string, address *backend.Address) (*backend.Address, error) {
	addresses, err := b.loadAddresses(user)
	if err != nil {
		return nil, err
	}

	address.ID = util.GenerateId()
	addresses = append(addresses, address)

	err = b.saveAddresses(user, addresses)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (b *Addresses) UpdateAddress(user string, update *backend.AddressUpdate) (*backend.Address, error) {
	addresses, err := b.loadAddresses(user)
	if err != nil {
		return nil, err
	}

	i, err := getAddressIndex(addresses, update.Address.ID)
	if err != nil {
		return nil, err
	}

	address := addresses[i]
	update.Apply(address)

	err = b.saveAddresses(user, addresses)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func (b *Addresses) DeleteAddress(user, id string) error {
	addresses, err := b.loadAddresses(user)
	if err != nil {
		return err
	}

	i, err := getAddressIndex(addresses, id)
	if err != nil {
		return err
	}

	addresses = append(addresses[:i], addresses[i+1:]...)

	err = b.saveAddresses(user, addresses)
	if err != nil {
		return err
	}

	return nil
}

func NewAddresses(config *Config) backend.AddressesBackend {
	return &Addresses{
		config: config,
	}
}

func UseAddresses(bkd *backend.Backend, config *Config) {
	bkd.Set(events.NewAddresses(NewAddresses(config), bkd))
}
