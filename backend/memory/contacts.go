package memory

import (
	"errors"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/util"
)

type Contacts struct {
	contacts map[string][]*backend.Contact
}

func (b *Contacts) ListContacts(user string) (contacts []*backend.Contact, err error) {
	contacts = b.contacts[user]
	return
}

func (b *Contacts) InsertContact(user string, contact *backend.Contact) (*backend.Contact, error) {
	contact.ID = util.GenerateId()
	b.contacts[user] = append(b.contacts[user], contact)
	return contact, nil
}

func (b *Contacts) getContactIndex(user, id string) (int, error) {
	for i, contact := range b.contacts[user] {
		if contact.ID == id {
			return i, nil
		}
	}

	return -1, errors.New("No such contact")
}

func (b *Contacts) UpdateContact(user string, update *backend.ContactUpdate) (*backend.Contact, error) {
	i, err := b.getContactIndex(user, update.Contact.ID)
	if err != nil {
		return nil, err
	}

	contact := b.contacts[user][i]
	update.Apply(contact)
	return contact, nil
}

func (b *Contacts) DeleteContact(user, id string) error {
	i, err := b.getContactIndex(user, id)
	if err != nil {
		return err
	}

	contacts := b.contacts[user]
	b.contacts[user] = append(contacts[:i], contacts[i+1:]...)

	return nil
}

func (b *Contacts) DeleteAllContacts(user string) error {
	b.contacts[user] = nil
	return nil
}

func NewContacts() backend.ContactsBackend {
	return &Contacts{
		contacts: map[string][]*backend.Contact{},
	}
}
