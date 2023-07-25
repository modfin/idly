package dao

import (
	"github.com/modfin/idly"
	"time"
)

type DAO interface {
	ListLogins(prefix string) ([]idly.Login, error)
	StoreLogin(login idly.Login, ttl time.Duration) error

	Close() error
}
