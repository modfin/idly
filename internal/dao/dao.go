package dao

import (
	"idly"
	"time"
)

type DAO interface {
	ListLogins(prefix string) ([]idly.Login, error)
	StoreLogin(login idly.Login, ttl time.Duration) error

	Close() error
}
