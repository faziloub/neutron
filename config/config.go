// Parses the configuration file.
package config

import (
	"github.com/faziloub/neutron/backend/disk"
	"github.com/faziloub/neutron/backend/imap"
	"github.com/faziloub/neutron/backend/smtp"
)

// Configuration for all backends.
// Backends omitted or set to null won't be activated.
type Config struct {
	// Memory config.
	Memory *MemoryConfig

	// IMAP config.
	Imap *ImapConfig

	// SMTP config.
	Smtp *SmtpConfig

	// Disk config.
	Disk *DiskConfig
}

type BackendConfig struct {
	Enabled bool
}

type MemoryConfig struct {
	*BackendConfig
	Populate bool
	Domains  []string
}

type ImapConfig struct {
	*BackendConfig
	*imap.Config
}

type SmtpConfig struct {
	*BackendConfig
	*smtp.Config
}

type DiskConfig struct {
	*BackendConfig
	*disk.Config

	Contacts      *DiskConfig
	Keys          *DiskConfig
	UsersSettings *DiskConfig
	Addresses     *DiskConfig
}
