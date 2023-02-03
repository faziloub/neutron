package events

import (
	"github.com/faziloub/neutron/backend"
)

type Addresses struct {
	backend.AddressesBackend
	events backend.EventsBackend
}

func (b *Addresses) InsertAddress(user string, addr *backend.Address) (inserted *backend.Address, err error) {
	inserted, err = b.AddressesBackend.InsertAddress(user, addr)
	if err != nil {
		return
	}

	event := backend.NewUserEvent(&backend.User{ID: user})
	b.events.InsertEvent(user, event)
	return
}

func (b *Addresses) UpdateAddress(user string, update *backend.AddressUpdate) (updated *backend.Address, err error) {
	updated, err = b.AddressesBackend.UpdateAddress(user, update)
	if err != nil {
		return
	}

	event := backend.NewUserEvent(&backend.User{ID: user})
	b.events.InsertEvent(user, event)
	return
}

func (b *Addresses) DeleteAddress(user, id string) (err error) {
	err = b.AddressesBackend.DeleteAddress(user, id)
	if err != nil {
		return
	}

	event := backend.NewUserEvent(&backend.User{ID: user})
	b.events.InsertEvent(user, event)
	return
}

func NewAddresses(addrs backend.AddressesBackend, events backend.EventsBackend) backend.AddressesBackend {
	return &Addresses{
		AddressesBackend: addrs,
		events:           events,
	}
}
