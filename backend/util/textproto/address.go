package textproto

import (
	"net/mail"

	"github.com/faziloub/neutron/backend"
)

func ParseEmail(addr *mail.Address) *backend.Email {
	return &backend.Email{Name: addr.Name, Address: addr.Address}
}

func FormatEmail(email *backend.Email) string {
	addr := &mail.Address{Name: email.Name, Address: email.Address}
	return addr.String()
}
