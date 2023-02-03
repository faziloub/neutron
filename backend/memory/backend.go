// Stores data in memory.
package memory

import (
	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/events"
	"github.com/faziloub/neutron/backend/util"
)

func Use(bkd *backend.Backend) {
	evts := NewEvents()
	contacts := events.NewContacts(NewContacts(), evts)
	labels := events.NewLabels(NewLabels(), evts)
	attachments := NewAttachments()
	messages := NewMessages(attachments.(*Attachments))
	conversations := events.NewConversations(NewConversations(messages.(*Messages)), evts)
	send := util.NewEchoSend(conversations)
	domains := NewDomains()
	users := NewUsers()
	addresses := events.NewAddresses(NewAddresses(), evts)
	keys := NewKeys()

	bkd.Set(contacts, labels, conversations, send, domains, evts, users, addresses, attachments, keys)
}
