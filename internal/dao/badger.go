package dao

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/modfin/henry/slicez"
	"github.com/modfin/idly"
	"time"
)

func NewBadger(uri string) (DAO, error) {

	db, err := badger.Open(badger.DefaultOptions(uri))
	if err != nil {
		return nil, err
	}

	return badger_{
		uri: uri,
		db:  db,
	}, nil
}

type badger_ struct {
	uri string
	db  *badger.DB
}

func (b badger_) ListLogins(prefix string) ([]idly.Login, error) {

	if !idly.IsLoginKeyPrefix(prefix) {
		return nil, fmt.Errorf("prefix %s, is not a valid prefix key, should be 'service/uid' format", prefix)
	}

	var logins []idly.Login

	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte(prefix)

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var login idly.Login
				err := json.Unmarshal(v, &login)
				if err != nil {
					return err
				}
				logins = append(logins, login)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	logins = slicez.SortFunc(logins, func(a, b idly.Login) bool {
		return !a.At.Before(b.At)
	})

	return logins, err
}

func (b badger_) StoreLogin(login idly.Login, ttl time.Duration) error {
	key := login.Key()
	value, err := login.Value()
	if err != nil {
		return fmt.Errorf("could not marshal login value: %w", err)
	}
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(
			badger.NewEntry(
				[]byte(key),
				[]byte(value),
			).WithTTL(ttl).WithDiscard())
	})
}

func (b badger_) Close() error {
	return b.db.Close()
}
