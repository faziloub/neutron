package disk

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/fazilb93/neutron/backend"
)

type UsersSettings struct {
	backend.UsersBackend

	config *Config
}

func (b *UsersSettings) getUserSettingsPath(id string) string {
	return b.config.Directory + "/" + id + ".json"
}

func (b *UsersSettings) copyUserSettings(src, dst *backend.User) {
	if src == nil {
		return
	}

	dst.NotificationEmail = src.NotificationEmail
	dst.Signature = src.Signature
	dst.NumMessagePerPage = src.NumMessagePerPage
	dst.Notify = src.Notify
	dst.AutoSaveContacts = src.AutoSaveContacts
	dst.Language = src.Language
	dst.LogAuth = src.LogAuth
	dst.ComposerMode = src.ComposerMode
	dst.MessageButtons = src.MessageButtons
	dst.ShowImages = src.ShowImages
	dst.ViewMode = src.ViewMode
	dst.ViewLayout = src.ViewLayout
	dst.SwipeLeft = src.SwipeLeft
	dst.SwipeRight = src.SwipeRight
	dst.Theme = src.Theme
	dst.Currency = src.Currency
	dst.DisplayName = src.DisplayName
}

func (b *UsersSettings) loadUserSettings(id string, user *backend.User) (err error) {
	data, err := ioutil.ReadFile(b.getUserSettingsPath(id))
	if os.IsNotExist(err) {
		err = nil
		return
	}
	if err != nil {
		return
	}

	settings := &backend.User{}
	err = json.Unmarshal(data, settings)
	b.copyUserSettings(settings, user)
	return
}

func (b *UsersSettings) saveUserSettings(id string, user *backend.User) (err error) {
	settings := &backend.User{}
	b.copyUserSettings(user, settings)

	data, err := json.Marshal(settings)
	if err != nil {
		return
	}

	err = os.MkdirAll(b.config.Directory, 0744)
	if err != nil {
		return
	}

	return ioutil.WriteFile(b.getUserSettingsPath(id), data, 0644)
}

func (b *UsersSettings) GetUser(id string) (user *backend.User, err error) {
	user, err = b.UsersBackend.GetUser(id)
	if err != nil {
		return
	}

	err = b.loadUserSettings(id, user)
	return
}

func (b *UsersSettings) Auth(username, password string) (user *backend.User, err error) {
	user, err = b.UsersBackend.Auth(username, password)
	if err != nil {
		return
	}

	err = b.loadUserSettings(user.ID, user)
	return
}

func (b *UsersSettings) InsertUser(user *backend.User, password string) (inserted *backend.User, err error) {
	inserted, err = b.UsersBackend.InsertUser(user, password)
	if err != nil {
		return
	}

	err = b.saveUserSettings(inserted.ID, user)
	if err != nil {
		return
	}

	b.copyUserSettings(user, inserted)
	return
}

func (b *UsersSettings) UpdateUser(update *backend.UserUpdate) (err error) {
	settings := &backend.User{}
	err = b.loadUserSettings(update.User.ID, settings)
	if err != nil {
		return
	}

	settings.ID = update.User.ID
	update.Apply(settings)

	err = b.saveUserSettings(update.User.ID, settings)
	return
}

func (b *UsersSettings) DeleteUser(id string) error {
	return errors.New("Not yet implemented") // TODO
}

func NewUsersSettings(config *Config, users backend.UsersBackend) backend.UsersBackend {
	return &UsersSettings{
		UsersBackend: users,
		config:       config,
	}
}

func UseUsersSettings(bkd *backend.Backend, config *Config) {
	bkd.Set(NewUsersSettings(config, bkd.UsersBackend))
}
