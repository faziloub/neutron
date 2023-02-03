package events

import (
	"github.com/faziloub/neutron/backend"
)

type Conversations struct {
	backend.ConversationsBackend
	messages backend.MessagesBackend
	events   backend.EventsBackend
}

func (b *Conversations) InsertMessage(user string, msg *backend.Message) (*backend.Message, error) {
	hadConversation := (msg.ConversationID != "")

	msg, err := b.messages.InsertMessage(user, msg)

	if err == nil && msg.ConversationID != "" {
		action := backend.EventCreate
		if hadConversation {
			action = backend.EventUpdate
		}

		conv, err := b.GetConversation(user, msg.ConversationID)
		if err == nil {
			event := backend.NewConversationDeltaEvent(msg.ConversationID, action, conv)
			b.events.InsertEvent(user, event)
		}

		// TODO: add ConversationCounts to event
	}

	return msg, err
}

func (b *Conversations) UpdateMessage(user string, update *backend.MessageUpdate) (*backend.Message, error) {
	msg, err := b.messages.UpdateMessage(user, update)

	if err == nil && msg.ConversationID != "" {
		conv, err := b.GetConversation(user, msg.ConversationID)
		if err == nil {
			event := backend.NewConversationDeltaEvent(msg.ConversationID, backend.EventUpdate, conv)
			b.events.InsertEvent(user, event)
		}
	}

	return msg, err
}

func (b *Conversations) DeleteMessage(user, id string) error {
	msg, _ := b.GetMessage(user, id)

	err := b.messages.DeleteMessage(user, id)

	if err == nil && msg != nil && msg.ConversationID != "" {
		conv, _ := b.GetConversation(user, msg.ConversationID)

		action := backend.EventUpdate
		if conv == nil {
			action = backend.EventDelete
		}

		event := backend.NewConversationDeltaEvent(msg.ConversationID, action, conv)
		b.events.InsertEvent(user, event)

		// TODO: add ConversationCounts to event
	}

	return err
}

func NewConversations(bkd backend.ConversationsBackend, events backend.EventsBackend) backend.ConversationsBackend {
	return &Conversations{
		ConversationsBackend: bkd,
		messages:             NewMessages(bkd, events),
		events:               events,
	}
}
