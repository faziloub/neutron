package api

import (
	"time"

	"github.com/faziloub/neutron/backend/util"
)

const SessionTimeout = 10 * time.Minute

type Session struct {
	ID      string
	UserID  string
	Token   string
	Timeout *time.Timer
}

func NewSession(user string, expire func()) *Session {
	return &Session{
		ID:      util.GenerateId(),
		Token:   util.GenerateId(),
		UserID:  user,
		Timeout: time.AfterFunc(SessionTimeout, expire),
	}
}
