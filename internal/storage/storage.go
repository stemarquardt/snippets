package storage

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Store struct {
	db   *bolt.DB
	path string
}

func New(p string) (*Store, error) {
	db, err := bolt.Open(p, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &Store{
		db:   db,
		path: p,
	}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Read(b, k []byte) ([]byte, error) {
	var ret []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(b)
		if bucket == nil {
			return fmt.Errorf("bucket %q not found", b)
		}
		v := bucket.Get(k)
		if v == nil {
			return fmt.Errorf("key %q not found in bucket %q", k, b)
		}
		ret = append(ret, v...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Store) Write(b, k, v []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(b)
		if err != nil {
			return err
		}
		return bucket.Put(k, v)
	})
}
