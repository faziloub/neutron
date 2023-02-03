package memory

import (
	"github.com/faziloub/neutron/backend"
)

func Populate(b *backend.Backend) (err error) {
	domains, err := b.ListDomains()
	if err != nil {
		return
	}

	var domain *backend.Domain
	if len(domains) > 0 {
		domain = domains[0]
	} else {
		domain, err = b.InsertDomain(&backend.Domain{DomainName: "example.org"})
		if err != nil {
			return
		}
	}

	email := "neutron@" + domain.DomainName

	user, err := b.InsertUser(&backend.User{
		Name:        "neutron",
		DisplayName: "Neutron",
	}, "neutron")
	if err != nil {
		return
	}

	addr, _ := b.InsertAddress(user.ID, &backend.Address{
		DomainID: domain.ID,
		Email:    email,
		Send:     1,
		Receive:  1,
		Status:   1,
		Type:     1,
	})

	b.InsertContact(user.ID, &backend.Contact{
		Name:  "Myself :)",
		Email: email,
	})

	b.InsertLabel(user.ID, &backend.Label{
		Name:    "Hey!",
		Color:   "#7272a7",
		Display: 1,
		Order:   1,
	})

	b.InsertMessage(user.ID, &backend.Message{
		ID:             "message_id",
		ConversationID: "conversation_id",
		AddressID:      addr.ID,
		Subject:        "Hello World",
		Sender:         &backend.Email{email, "Neutron"},
		ToList:         []*backend.Email{&backend.Email{email, "Neutron"}},
		Time:           1458073557,
		Body:           "Hey! How are you today?",
		LabelIDs:       []string{backend.InboxLabel},
	})

	return
}
