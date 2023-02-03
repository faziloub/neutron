package memory

import (
	"errors"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/util"
)

type attachment struct {
	*backend.Attachment
	Contents []byte
}

type Attachments struct {
	attachments map[string][]*attachment
}

func (b *Attachments) getAttachmentIndex(user, id string) (int, error) {
	for i, att := range b.attachments[user] {
		if att.ID == id {
			return i, nil
		}
	}
	return -1, errors.New("No such attachment")
}

func (b *Attachments) ListAttachments(user, msgId string) (atts []*backend.Attachment, err error) {
	for _, att := range b.attachments[user] {
		if att.MessageID == msgId {
			atts = append(atts, att.Attachment)
		}
	}
	return
}

func (b *Attachments) ReadAttachment(user, id string) (*backend.Attachment, []byte, error) {
	i, err := b.getAttachmentIndex(user, id)
	if err != nil {
		return nil, nil, err
	}

	att := b.attachments[user][i]
	return att.Attachment, att.Contents, nil
}

func (b *Attachments) InsertAttachment(user string, att *backend.Attachment, contents []byte) (*backend.Attachment, error) {
	att.ID = util.GenerateId()
	att.Size = len(contents)
	b.attachments[user] = append(b.attachments[user], &attachment{
		Attachment: att,
		Contents:   contents,
	})
	return att, nil
}

func (b *Attachments) DeleteAttachment(user, id string) (err error) {
	i, err := b.getAttachmentIndex(user, id)
	if err != nil {
		return
	}

	b.attachments[user] = append(b.attachments[user][:i], b.attachments[user][i+1:]...)
	return
}

// Additional function needed for imap.Messages backend
func (b *Attachments) UpdateAttachmentMessage(user, id, msgId string) error {
	i, err := b.getAttachmentIndex(user, id)
	if err != nil {
		return err
	}

	b.attachments[user][i].MessageID = msgId
	return nil
}

func NewAttachments() backend.AttachmentsBackend {
	return &Attachments{
		attachments: map[string][]*attachment{},
	}
}
