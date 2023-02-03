package util

import (
	"time"

	"github.com/fazilb93/neutron/backend"
)

// A SendBackend that does nothing.
type NoopSend struct{}

func (b *NoopSend) SendMessage(user string, msg *backend.OutgoingMessage) error {
	return nil // Do nothing
}

func NewNoopSend() backend.SendBackend {
	return &NoopSend{}
}

// A SendBackend that forwards all sent messages to a MessagesBackend.
type EchoSend struct {
	target backend.MessagesBackend
}

func (b *EchoSend) SendMessage(user string, msg *backend.OutgoingMessage) error {
	newMsg := *msg.Message // Copy msg
	newMsg.Subject = "[EchoSend forwarded message] " + newMsg.Subject
	newMsg.Body = msg.MessagePackage.Body
	newMsg.Time = time.Now().Unix()
	newMsg.LabelIDs = []string{backend.InboxLabel}
	newMsg.Type = 0
	newMsg.IsRead = 0
	newMsg.Attachments = msg.Message.Attachments

	_, err := b.target.InsertMessage(user, &newMsg)
	return err
}

func NewEchoSend(target backend.MessagesBackend) backend.SendBackend {
	return &EchoSend{
		target: target,
	}
}
