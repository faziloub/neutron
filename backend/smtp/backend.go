// Sends messages with a SMTP server.
package smtp

import (
	"strconv"

	"github.com/faziloub/neutron/backend"
)

type PasswordsBackend interface {
	GetPassword(user string) (string, error)
}

type Config struct {
	Hostname string
	Port     int
	Suffix   string
	Tls      bool
	SmtpHost string
}

func (c *Config) Host() string {
	port := c.Port
	if port <= 0 {
		if c.Tls {
			port = 465
		} else {
			port = 25
		}
	}

	return c.Hostname + ":" + strconv.Itoa(port)
}

func Use(bkd *backend.Backend, config *Config, passwords PasswordsBackend) {
	send := New(config, passwords)

	bkd.Set(send)
}
