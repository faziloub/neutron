package smtp

import (
	"crypto/tls"
	"errors"
	"net"
	"net/smtp"

	"github.com/faziloub/neutron/backend"
	"github.com/faziloub/neutron/backend/util/textproto"
)

type SendBackend struct {
	PasswordsBackend

	config *Config
}

func (b *SendBackend) SendMessage(user string, msg *backend.OutgoingMessage) error {
	password, err := b.GetPassword(user)
	if err != nil {
		return err
	}

	cfg := b.config
	host := cfg.Host()

	var conn net.Conn
	if cfg.Tls {
		conn, err = tls.Dial("tcp", host, nil)
	} else {
		conn, err = net.Dial("tcp", host)
	}
	if err != nil {
		return err
	}

	smtpHost := cfg.SmtpHost
	if smtpHost == "" {
		smtpHost = cfg.Hostname
	}
	c, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return err
	}
	defer c.Close()

	if !cfg.Tls {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return errors.New("STMP server doesn't support STARTTLS")
		}

		tlsConfig := &tls.Config{ServerName: smtpHost}
		if err = c.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	auth := smtp.PlainAuth("", user+b.config.Suffix, password, smtpHost)
	if err = c.Auth(auth); err != nil {
		return err
	}

	from := msg.Sender.Address
	to := []string{msg.MessagePackage.Address}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	mail := []byte(textproto.FormatOutgoingMessage(msg))

	if _, err = w.Write(mail); err != nil {
		return err
	}

	if err = w.Close(); err != nil {
		return err
	}

	return c.Quit()
}

func New(config *Config, passwords PasswordsBackend) backend.SendBackend {
	if config.Port <= 0 {
		config.Port = 25
	}

	return &SendBackend{
		PasswordsBackend: passwords,

		config: config,
	}
}
