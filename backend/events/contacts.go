package events

import (
	"github.com/fazilb93/neutron/backend"
)

type Contacts struct {
	backend.ContactsBackend
	events backend.EventsBackend
}

func (b *Contacts) InsertContact(user string, contact *backend.Contact) (*backend.Contact, error) {
	contact, err := b.ContactsBackend.InsertContact(user, contact)

	if err == nil {
		event := backend.NewContactDeltaEvent(contact.ID, backend.EventCreate, contact)
		b.events.InsertEvent(user, event)
	}

	return contact, err
}

func (b *Contacts) UpdateContact(user string, update *backend.ContactUpdate) (*backend.Contact, error) {
	contact, err := b.ContactsBackend.UpdateContact(user, update)

	if err == nil {
		event := backend.NewContactDeltaEvent(contact.ID, backend.EventUpdate, contact)
		b.events.InsertEvent(user, event)
	}

	return contact, err
}

func (b *Contacts) DeleteContact(user, id string) error {
	err := b.ContactsBackend.DeleteContact(user, id)

	if err == nil {
		event := backend.NewContactDeltaEvent(id, backend.EventDelete, nil)
		b.events.InsertEvent(user, event)
	}

	return err
}

func NewContacts(bkd backend.ContactsBackend, events backend.EventsBackend) backend.ContactsBackend {
	return &Contacts{
		ContactsBackend: bkd,
		events:          events,
	}
}
