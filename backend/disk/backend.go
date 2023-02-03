// Stores data in files on disk.
package disk

import (
	"github.com/faziloub/neutron/backend"
)

type Config struct {
	Directory string
}

func Use(bkd *backend.Backend, config *Config) {
	keys := NewKeys(config, bkd)

	bkd.Set(keys)
}
