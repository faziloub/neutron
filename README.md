# neutron

[![Build Status](https://travis-ci.org/emersion/neutron.svg?branch=master)](https://travis-ci.org/emersion/neutron)
[![GoDoc](https://godoc.org/github.com/faziloub/neutron?status.svg)](https://godoc.org/github.com/faziloub/neutron)

Self-hosted server for [Protonmail client](https://github.com/ProtonMail/WebClient).

> This project is not affiliated or supported by ProtonMail.

## What is it?

Neutron is a server that will allow the ProtonMail client to be used with
_backends_. Several backends are available right now:
* IMAP: this will read and store messages on your IMAP server. Received messages
  will stay as is (that is, unencrypted) but messages saved from the web client
  will be encrypted. You will login to the web client with your IMAP username
  and password.
* SMTP: this will send messages using your SMTP server. Messages are sent
  encrypted to the server. If a recipient's public key is not found, the server
  will decrypt the message before sending it to this recipient.
* Filesystem: settings, contacts, keys are stored on disk. Keys are always
  stored encrypted.
* Memory: all is stored in memory and will be destroyed when the server is
  stopped.

Neutron is modular so it's easy to create new backends and handle more scenarios.

Keep in mind that Neutron is less secure than ProtonMail: most servers don't
use full-disk encryption and aren't under 1,000 meters of granite rock in
Switzerland. Also, SRP is not yet supported ([#35](https://github.com/faziloub/neutron/issues/35)).
If you use Neutron, make sure to [donate to ProtonMail](https://protonmail.com/donate)!

## Install

* Debian, Ubuntu & Fedora: ~~install from https://packager.io/gh/emersion/neutron
  and run with `neutronmail run web`~~
* Other platforms: no packages yet, you'll have to build from source (see below)

### Configuration

See `config.json`. You'll have to change IMAP and SMTP settings to match your
mail server config.

```js
{
	"Memory": {
		"Enabled": true,
		"Populate": false, // Populate server with default neutron user
		"Domains": ["emersion.fr"] // Available e-mail domains
	},
	"Imap": { // IMAP server config
		"Enabled": true,
		"Hostname": "mail.gandi.net",
		"Tls": true,
		"Suffix": "@emersion.fr" // Will be appended to username when authenticating
	},
	"Smtp": { // SMTP server config
		"Enabled": true,
		"Hostname": "mail.gandi.net",
		"Port": 587,
		"Suffix": "@emersion.fr" // Will be appended to username when authenticating
	},
	"Disk": { // Store keys, contacts and settings on disk
		"Enabled": true,
		"Keys": { "Directory": "db/keys" }, // PGP keys location
		"Contacts": { "Directory": "db/contacts" },
		"UsersSettings": { "Directory": "db/settings" },
		"Addresses": { "Directory": "db/addresses" }
	}
}
```

### Usage

To generate keys for a new user the first time, just click _Sign up_ on the
login page and enter your IMAP credentials.

### Options

* `-config`: specify a custom config file
* `-help`: show help

## Build

Requirements:
* Go (to build the server)
* Node, NPM (to build the client)

```shell
# Get the code
go get -u github.com/faziloub/neutron
cd $GOPATH/src/github.com/faziloub/neutron

# Build the client
git submodule init
git submodule update
make build-client

# Start the server
make start
```

### Docker

```shell
make build-docker
docker build -t neutron .
docker create -p 4000:4000 -v $PWD/config.json:/config.json -v $PWD/db:/db neutron
```

## Backends

All backends must implement the [backend interface](https://github.com/faziloub/neutron/blob/master/backend/backend.go).
The main backend interface is split into multiple other backend interfaces for
different roles: `ContactsBackend`, `LabelsBackend` and so on. This allows to
build modular backends, e.g. a `MessagesBackend` which stores messages on an
IMAP server with a `ContactsBackend` which stores contacts on a LDAP server and
a `SendBackend` which sends outgoing messages to a SMTP server.

Writing a backend is just a matter of implementing the necessary functions. You
can read the [`memory` backend](https://github.com/faziloub/neutron/tree/master/backend/memory)
to understand how to do that. Docs for the backend are available here:
https://godoc.org/github.com/faziloub/neutron/backend#Backend

## License

MIT
