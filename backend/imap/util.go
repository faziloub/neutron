package imap

import (
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/faziloub/go-imap"
	"github.com/faziloub/neutron/backend"
	_textproto "github.com/faziloub/neutron/backend/util/textproto"
)

func formatAttachmentId(mailbox string, uid uint32, part string) string {
	raw := mailbox + "/" + strconv.Itoa(int(uid))
	if part != "" {
		raw += "#" + part
	}
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func formatMessageId(mailbox string, uid uint32) string {
	return formatAttachmentId(mailbox, uid, "")
}

func parseAttachmentId(id string) (mailbox string, uid uint32, part string, err error) {
	decoded, err := base64.URLEncoding.DecodeString(id)
	if err != nil {
		return
	}

	fstParts := strings.SplitN(string(decoded), "/", 2)
	if len(fstParts) != 2 {
		err = errors.New("Invalid message ID: does not contain separator")
		return
	}
	sndParts := strings.SplitN(fstParts[1], "#", 2)

	uidInt, err := strconv.Atoi(sndParts[0])
	if err != nil {
		return
	}

	mailbox = fstParts[0]
	uid = uint32(uidInt)

	if len(sndParts) == 2 {
		part = sndParts[1]
	}
	return
}

func parseMessageId(id string) (mailbox string, uid uint32, err error) {
	mailbox, uid, _, err = parseAttachmentId(id)
	return
}

func parseMessage(msg *backend.Message, src *imap.Message) {
	msg.Order = int(src.SeqNum)
	msg.Size = int(src.Size)

	for _, flag := range src.Flags {
		switch flag {
		case imap.SeenFlag:
			msg.IsRead = 1
		case imap.AnsweredFlag:
			msg.IsReplied = 1
		case imap.FlaggedFlag:
			msg.Starred = 1
			msg.LabelIDs = append(msg.LabelIDs, backend.StarredLabel)
		case imap.DraftFlag:
			msg.Type = backend.DraftType
		}
	}
}

func bodyStructureAttachments(structure *imap.BodyStructure) []*backend.Attachment {
	// Non-multipart messages don't contain attachments
	if structure.MIMEType != "multipart" || structure.MIMESubType == "alternative" {
		return nil
	}

	var attachments []*backend.Attachment
	for i, part := range structure.Parts {
		if part.MIMEType == "multipart" {
			attachments = append(attachments, bodyStructureAttachments(part)...)
			continue
		}

		// Apple Mail doesn't format well headers
		// First child is message content
		if part.MIMEType == "text" && i == 0 {
			continue
		}

		attachments = append(attachments, &backend.Attachment{
			ID:       part.Id,
			Name:     part.Params["name"],
			MIMEType: part.MIMEType + "/" + part.MIMESubType,
			Size:     int(part.Size),
		})
	}

	return attachments
}

func getPreferredPart(structure *imap.BodyStructure) (path string, part *imap.BodyStructure) {
	part = structure

	for i, p := range structure.Parts {
		if p.MIMEType == "multipart" && p.MIMESubType == "alternative" {
			path, part = getPreferredPart(p)
			path = strconv.Itoa(i+1) + "." + path
		}
		if p.MIMEType != "text" {
			continue
		}
		if part.MIMEType == "multipart" || p.MIMESubType == "html" {
			part = p
			path = strconv.Itoa(i + 1)
		}
	}

	return
}

func decodePart(part *imap.BodyStructure, r io.Reader) io.Reader {
	return _textproto.Decode(r, part.Encoding, part.Params["charset"])
}

func parseAttachment(r io.Reader) (att *backend.Attachment, body io.Reader) {
	br := bufio.NewReader(r)
	h, err := textproto.NewReader(br).ReadMIMEHeader()
	if err != nil {
		return
	}

	mediaType, params, _ := mime.ParseMediaType(h.Get("Content-Type"))

	att = &backend.Attachment{
		ID:       h.Get("Content-Id"),
		Name:     params["name"],
		MIMEType: mediaType,
	}

	if size := h.Get("Content-Size"); size != "" {
		att.Size, _ = strconv.Atoi(size)
	}

	body = _textproto.Decode(br, h.Get("Content-Encoding"), params["charset"])
	return
}

func parseAddress(addr *imap.Address) *backend.Email {
	return &backend.Email{
		Name:    _textproto.DecodeWord(addr.PersonalName),
		Address: addr.MailboxName + "@" + addr.HostName,
	}
}

func parseAddressList(list []*imap.Address) []*backend.Email {
	emails := make([]*backend.Email, len(list))
	for i, addr := range list {
		emails[i] = parseAddress(addr)
	}
	return emails
}

func parseEnvelope(msg *backend.Message, envelope *imap.Envelope) {
	if !envelope.Date.IsZero() {
		msg.Time = envelope.Date.Unix()
	}

	msg.Subject = envelope.Subject // _textproto.DecodeWord()

	if len(envelope.Sender) > 0 {
		msg.Sender = parseAddress(envelope.Sender[0])
	}

	if len(envelope.ReplyTo) > 0 {
		msg.ReplyTo = parseAddress(envelope.ReplyTo[0])
	}

	msg.ToList = parseAddressList(envelope.To)
	msg.CCList = parseAddressList(envelope.Cc)
	msg.BCCList = parseAddressList(envelope.Bcc)
}
