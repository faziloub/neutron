package disk

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/events"
	"github.com/faziloub/neutron/backend/util"
)

type Contacts struct {
	config *Config
}

func (b *Contacts) getContactsPath(user string) string {
	return b.config.Directory + "/" + user + ".json"
}

func (b *Contacts) loadContacts(user string) (contacts []*backend.Contact, err error) {
	data, err := ioutil.ReadFile(b.getContactsPath(user))
	if os.IsNotExist(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &contacts)
	return
}

func (b *Contacts) saveContacts(user string, contacts []*backend.Contact) (err error) {
	data, err := json.Marshal(contacts)
	if err != nil {
		return
	}

	err = os.MkdirAll(b.config.Directory, 0744)
	if err != nil {
		return
	}

	return ioutil.WriteFile(b.getContactsPath(user), data, 0644)
}

func (b *Contacts) ListContacts(user string) ([]*backend.Contact, error) {
	return b.loadContacts(user)
}

func (b *Contacts) InsertContact(user string, contact *backend.Contact) (*backend.Contact, error) {
	contacts, err := b.loadContacts(user)
	if err != nil {
		return nil, err
	}

	contact.ID = util.GenerateId()
	contacts = append(contacts, contact)

	err = b.saveContacts(user, contacts)
	if err != nil {
		return nil, err
	}

	return contact, nil
}

func getContactIndex(contacts []*backend.Contact, id string) (int, error) {
	for i, contact := range contacts {
		if contact.ID == id {
			return i, nil
		}
	}

	return -1, errors.New("No such contact")
}

func (b *Contacts) UpdateContact(user string, update *backend.ContactUpdate) (*backend.Contact, error) {
	contacts, err := b.loadContacts(user)
	if err != nil {
		return nil, err
	}

	i, err := getContactIndex(contacts, update.Contact.ID)
	if err != nil {
		return nil, err
	}

	contact := contacts[i]
	update.Apply(contact)

	err = b.saveContacts(user, contacts)
	if err != nil {
		return nil, err
	}

	return contact, nil
}

func (b *Contacts) DeleteContact(user, id string) error {
	contacts, err := b.loadContacts(user)
	if err != nil {
		return err
	}

	i, err := getContactIndex(contacts, id)
	if err != nil {
		return err
	}

	contacts = append(contacts[:i], contacts[i+1:]...)

	err = b.saveContacts(user, contacts)
	if err != nil {
		return err
	}

	return nil
}

func (b *Contacts) DeleteAllContacts(user string) error {
	contacts := []*backend.Contact{}
	return b.saveContacts(user, contacts)
}

func NewContacts(config *Config) backend.ContactsBackend {
	return &Contacts{
		config: config,
	}
}

func UseContacts(bkd *backend.Backend, config *Config) {
	bkd.Set(events.NewContacts(NewContacts(config), bkd))
}
