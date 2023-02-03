package memory

import (
	"errors"

	"github.com/fazilb93/neutron/backend"
	"github.com/fazilb93/neutron/backend/util"
)

type Conversations struct {
	*Messages
}

func isEmailInList(needle *backend.Email, haystack []*backend.Email) bool {
	for _, email := range haystack {
		if needle.Address == email.Address {
			return true
		}
	}
	return false
}

func populateConversation(conv *backend.Conversation, msg *backend.Message) {
	conv.NumMessages++
	if msg.IsRead == 0 {
		conv.NumUnread++
	}

	if msg.Time > conv.Time {
		conv.Time = msg.Time
		conv.Subject = msg.Subject
	}

	if !isEmailInList(msg.Sender, conv.Senders) {
		conv.Senders = append(conv.Senders, msg.Sender)
	}

	for _, email := range msg.ToList {
		if !isEmailInList(email, conv.Recipients) {
			conv.Recipients = append(conv.Recipients, email)
		}
	}

	for _, labelId := range msg.LabelIDs {
		var label *backend.ConversationLabel
		for _, l := range conv.Labels {
			if l.ID == labelId {
				label = l
				break
			}
		}

		if label == nil {
			label = &backend.ConversationLabel{ID: labelId}
			conv.Labels = append(conv.Labels, label)
			conv.LabelIDs = append(conv.LabelIDs, labelId)
		}

		label.NumMessages++
		if msg.IsRead == 0 {
			label.NumUnread++
		}
	}
}

func (b *Conversations) ListConversationMessages(user, id string) (msgs []*backend.Message, err error) {
	for _, msg := range b.messages[user] {
		if msg.ConversationID == id {
			msgs = append(msgs, msg)
		}
	}
	return
}

func (b *Conversations) listConversations(user string) (convs []*backend.Conversation, err error) {
	for _, msg := range b.messages[user] {
		var conv *backend.Conversation
		for _, c := range convs {
			if c.ID == msg.ConversationID {
				conv = c
				break
			}
		}

		if conv == nil {
			conv = &backend.Conversation{ID: msg.ConversationID}
			convs = append(convs, conv)
		}

		populateConversation(conv, msg)
	}

	return
}

func (b *Conversations) ListConversations(user string, filter *backend.MessagesFilter) (convs []*backend.Conversation, total int, err error) {
	all, err := b.listConversations(user)
	if err != nil {
		return
	}

	filtered := []*backend.Conversation{}

	for _, c := range all {
		if filter.Label != "" {
			matches := false
			for _, lbl := range c.LabelIDs {
				if lbl == filter.Label {
					matches = true
					break
				}
			}

			if !matches {
				continue
			}
		}

		// TODO: other filter fields support

		filtered = append(filtered, c)
	}

	total = len(filtered)

	if filter.Limit > 0 && filter.Page >= 0 {
		from := filter.Limit * filter.Page
		to := filter.Limit * (filter.Page + 1)
		if from < 0 {
			from = 0
		}
		if to > total {
			to = total
		}

		convs = filtered[from:to]
	} else {
		convs = filtered
	}

	return
}

func (b *Conversations) CountConversations(user string) (counts []*backend.MessagesCount, err error) {
	convs, err := b.listConversations(user)
	if err != nil {
		return
	}

	indexes := map[string]int{}

	for _, c := range convs {
		for _, label := range c.LabelIDs {
			var count *backend.MessagesCount
			if i, ok := indexes[label]; ok {
				count = counts[i]
			} else {
				indexes[label] = len(counts)
				count = &backend.MessagesCount{LabelID: label}
				counts = append(counts, count)
			}

			count.Total++
			if c.NumUnread > 0 {
				count.Unread++
			}
		}
	}

	return
}

func (b *Conversations) GetConversation(user, id string) (conv *backend.Conversation, err error) {
	for _, msg := range b.messages[user] {
		if msg.ConversationID == id {
			if conv == nil {
				conv = &backend.Conversation{ID: id}
			}

			populateConversation(conv, msg)
		}
	}

	if conv == nil {
		err = errors.New("No such conversation")
	}
	return
}

func (b *Conversations) InsertMessage(user string, msg *backend.Message) (*backend.Message, error) {
	if msg.ConversationID == "" {
		msg.ConversationID = util.GenerateId()
	}

	return b.Messages.InsertMessage(user, msg)
}

func (b *Conversations) DeleteConversation(user, id string) (err error) {
	// Delete all messages in conversation
	msgs, err := b.ListConversationMessages(user, id)
	if err != nil {
		return
	}

	for _, msg := range msgs {
		err = b.DeleteMessage(user, msg.ID)
		if err != nil {
			return
		}
	}

	return
}

func NewConversations(messages *Messages) backend.ConversationsBackend {
	return &Conversations{
		Messages: messages,
	}
}
