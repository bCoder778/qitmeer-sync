package storage

import "github.com/bCoder778/qitmeer-sync/storage/types"

func (s *Storage) QueryMemTransaction() []types.Transaction {
	txs, _ := s.db.QueryMemTransaction()
	return txs
}

func (s *Storage) QueryUnconfirmedTranslateTransaction() []types.Transaction {
	txs, _ := s.db.QueryUnconfirmedTranslateTransaction()
	return txs
}
