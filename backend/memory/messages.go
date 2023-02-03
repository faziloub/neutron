package memory

import (
	"errors"

	"github.com/fazilb93/neutron/backend"
	"github.com/fazilb93/neutron/backend/util"
)

type Messages struct {
	*Attachments

	messages map[string][]*backend.Message
}

func (b *Messages) getMessageIndex(user, id string) (int, error) {
	for i, m := range b.messages[user] {
		if m.ID == id {
			return i, nil
		}
	}

	return -1, errors.New("No such message")
}

func (b *Messages) GetMessage(user, id string) (msg *backend.Message, err error) {
	i, err := b.getMessageIndex(user, id)
	if err != nil {
		return
	}

	msg = b.messages[user][i]

	atts, err := b.ListAttachments(user, id)
	if err != nil {
		return
	}
	msg.Attachments = atts
	return
}

func (b *Messages) ListMessages(user string, filter *backend.MessagesFilter) (msgs []*backend.Message, total int, err error) {
	all := b.messages[user]
	filtered := []*backend.Message{}

	for _, msg := range all {
		if filter.Label != "" {
			matches := false
			for _, lbl := range msg.LabelIDs {
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

		msg.Attachments, _ = b.ListAttachments(user, msg.ID)
		filtered = append(filtered, msg)
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

		msgs = filtered[from:to]
	} else {
		msgs = filtered
	}

	return
}

func (b *Messages) CountMessages(user string) (counts []*backend.MessagesCount, err error) {
	indexes := map[string]int{}

	for _, msg := range b.messages[user] {
		for _, label := range msg.LabelIDs {
			var count *backend.MessagesCount
			if i, ok := indexes[label]; ok {
				count = counts[i]
			} else {
				indexes[label] = len(counts)
				count = &backend.MessagesCount{LabelID: label}
				counts = append(counts, count)
			}

			count.Total++
			if msg.IsRead == 0 {
				count.Unread++
			}
		}
	}

	return
}

func (b *Messages) InsertMessage(user string, msg *backend.Message) (*backend.Message, error) {
	msg.ID = util.GenerateId()
	msg.Order = len(b.messages[user])
	msg.NumAttachments = len(msg.Attachments)
	b.messages[user] = append(b.messages[user], msg)
	return msg, nil
}

func (b *Messages) UpdateMessage(user string, update *backend.MessageUpdate) (msg *backend.Message, err error) {
	i, err := b.getMessageIndex(user, update.Message.ID)
	if err != nil {
		return
	}

	msg = b.messages[user][i]
	update.Apply(msg)
	return
}

func (b *Messages) DeleteMessage(user, id string) error {
	i, err := b.getMessageIndex(user, id)
	if err != nil {
		return err
	}

	messages := b.messages[user]
	b.messages[user] = append(messages[:i], messages[i+1:]...)

	return nil
}

func NewMessages(attachments *Attachments) backend.MessagesBackend {
	return &Messages{
		Attachments: attachments,
		messages:    map[string][]*backend.Message{},
	}
}
