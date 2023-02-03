package imap

import (
	"errors"
	"io/ioutil"

	"github.com/faziloub/go-imap"
	"github.com/faziloub/neutron/backend"
)

func (b *Messages) ListAttachments(user, msg string) ([]*backend.Attachment, error) {
	return nil, errors.New("Not yet implemented")
}

func (b *Messages) ReadAttachment(user, id string) (att *backend.Attachment, out []byte, err error) {
	// First, try to get attachment from temporary backend
	att, out, err = b.tmpAtts.ReadAttachment(user, id)
	if err == nil {
		return
	}

	// Not found in tmp backend, get it from the server

	mailbox, uid, partId, err := parseAttachmentId(id)
	if err != nil {
		return
	}

	err = b.selectMailbox(user, mailbox)
	if err != nil {
		return
	}

	c, unlock, err := b.getConn(user)
	if err != nil {
		return
	}
	defer unlock()

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	items := []imap.FetchItem{imap.FetchItem("BODY.PEEK[" + partId + "]")}

	messages := make(chan *imap.Message, 1)
	if err = c.UidFetch(seqset, items, messages); err != nil {
		return
	}

	data := <-messages
	if data == nil {
		err = errors.New("No such attachment (cannot find parent message)")
		return
	}

	var bodySectionName *imap.BodySectionName
	bodySectionName, err = imap.ParseBodySectionName(imap.FetchItem("BODY[" + partId + "]"))
	if err != nil {
		return
	}
	att, r := parseAttachment(data.GetBody(bodySectionName))

	out, err = ioutil.ReadAll(r)
	return
}

func (b *Messages) InsertAttachment(user string, attachment *backend.Attachment, data []byte) (*backend.Attachment, error) {
	return b.tmpAtts.InsertAttachment(user, attachment, data)
}

func (b *Messages) DeleteAttachment(user, id string) error {
	return b.tmpAtts.DeleteAttachment(user, id)
}
