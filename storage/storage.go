package storage

import (
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/verify"
	"sync"
)

type Storage struct {
	mutex  sync.RWMutex
	db     db.IDB
	verify *verify.QitmeerVerify
}

func NewStorage(db db.IDB, v *verify.QitmeerVerify) *Storage {
	return &Storage{db: db, verify: v}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
