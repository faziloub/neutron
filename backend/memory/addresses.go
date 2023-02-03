package memory

import (
	"errors"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/util"
)

type Addresses struct {
	addresses map[string][]*backend.Address
}

func (b *Addresses) getAddressIndex(user, id string) (int, error) {
	for i, addr := range b.addresses[user] {
		if addr.ID != id {
			continue
		}

		return i, nil
	}

	return -1, errors.New("No such address")
}

func (b *Addresses) GetAddress(user, id string) (address *backend.Address, err error) {
	i, err := b.getAddressIndex(user, id)
	if err != nil {
		return
	}

	address = b.addresses[user][i]
	return
}

func (b *Addresses) ListAddresses(user string) (addrs []*backend.Address, err error) {
	addrs = b.addresses[user]
	return
}

func (b *Addresses) InsertAddress(user string, addr *backend.Address) (*backend.Address, error) {
	addr.ID = util.GenerateId()
	b.addresses[user] = append(b.addresses[user], addr)
	return addr, nil
}

func (b *Addresses) UpdateAddress(user string, update *backend.AddressUpdate) (addr *backend.Address, err error) {
	i, err := b.getAddressIndex(user, update.Address.ID)
	if err != nil {
		return
	}

	addr = b.addresses[user][i]
	update.Apply(addr)
	return
}

func (b *Addresses) DeleteAddress(user, id string) (err error) {
	i, err := b.getAddressIndex(user, id)
	if err != nil {
		return
	}

	addresses := b.addresses[user]
	b.addresses[user] = append(addresses[:i], addresses[i+1:]...)
	return
}

func NewAddresses() backend.AddressesBackend {
	return &Addresses{
		addresses: map[string][]*backend.Address{},
	}
}
