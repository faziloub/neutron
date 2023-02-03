package util

import (
	"github.com/faziloub/neutron/backend"
)

// A conversations backend that builds one conversation per message (no threads).
type DummyConversations struct {
	backend.MessagesBackend
}

func (b *DummyConversations) ListConversationMessages(user, id string) ([]*backend.Message, error) {
	msg, err := b.GetMessage(user, id)
	if err != nil {
		return nil, err
	}
	msg.ConversationID = id
	return []*backend.Message{msg}, nil
}

func (b *DummyConversations) buildConversation(msg *backend.Message) *backend.Conversation {
	conv := &backend.Conversation{
		ID:          msg.ID,
		Order:       msg.Order,
		NumMessages: 1,
		NumUnread:   1 - msg.IsRead,
		Time:        msg.Time,
		Subject:     msg.Subject,
		Senders:     []*backend.Email{msg.Sender},
		Recipients:  msg.ToList,
		TotalSize:   msg.Size,
		LabelIDs:    msg.LabelIDs,
	}

	for _, lbl := range msg.LabelIDs {
		conv.Labels = append(conv.Labels, &backend.ConversationLabel{
			ID:          lbl,
			NumMessages: 1,
			NumUnread:   1 - msg.IsRead,
		})
	}

	return conv
}

func (b *DummyConversations) GetConversation(user, id string) (*backend.Conversation, error) {
	msg, err := b.GetMessage(user, id)
	if err != nil {
		return nil, err
	}
	return b.buildConversation(msg), nil
}

func (b *DummyConversations) ListConversations(user string, filter *backend.MessagesFilter) ([]*backend.Conversation, int, error) {
	msgs, total, err := b.ListMessages(user, filter)
	if err != nil {
		return nil, -1, err
	}

	convs := make([]*backend.Conversation, len(msgs))
	for i, msg := range msgs {
		convs[i] = b.buildConversation(msg)
	}

	return convs, total, nil
}

func (b *DummyConversations) CountConversations(user string) ([]*backend.MessagesCount, error) {
	return b.CountMessages(user)
}

func (b *DummyConversations) DeleteConversation(user, id string) error {
	return b.DeleteMessage(user, id)
}

func (b *DummyConversations) GetMessage(user, id string) (*backend.Message, error) {
	msg, err := b.MessagesBackend.GetMessage(user, id)

	if err == nil {
		msg.ConversationID = msg.ID
	}

	return msg, err
}

func (b *DummyConversations) ListMessages(user string, filter *backend.MessagesFilter) ([]*backend.Message, int, error) {
	msgs, total, err := b.MessagesBackend.ListMessages(user, filter)

	if err == nil {
		for _, msg := range msgs {
			msg.ConversationID = msg.ID
		}
	}

	return msgs, total, err
}

func (b *DummyConversations) InsertMessage(user string, msg *backend.Message) (*backend.Message, error) {
	msg, err := b.MessagesBackend.InsertMessage(user, msg)

	if err == nil {
		msg.ConversationID = msg.ID
	}

	return msg, err
}

func (b *DummyConversations) UpdateMessage(user string, update *backend.MessageUpdate) (*backend.Message, error) {
	msg, err := b.MessagesBackend.UpdateMessage(user, update)

	if err == nil {
		msg.ConversationID = msg.ID
	}

	return msg, err
}

func NewDummyConversations(messages backend.MessagesBackend) backend.ConversationsBackend {
	return &DummyConversations{messages}
}
