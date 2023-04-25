package storage

import "github.com/bCoder778/qitmeer-sync/storage/types"

func (s *Storage) GetLastEVmHeightBlock() *types.Block {
	return s.db.GetLastEVmHeightBlock()
}

func (s *Storage) LastOrder() uint64 {
	order, _ := s.db.GetLastOrder()
	return order
}

func (s *Storage) LastId() uint64 {
	order, _ := s.db.GetLastId()
	return order
}

func (s *Storage) UnconfirmedOrders() []uint64 {
	orders, _ := s.db.QueryUnConfirmedOrders()
	return orders
}

func (s *Storage) UnconfirmedIds() []uint64 {
	ids, _ := s.db.QueryUnConfirmedIds()
	return ids
}

func (s *Storage) UnconfirmedIdsByCount(count int) []uint64 {
	ids, _ := s.db.QueryUnConfirmedIdsByCount(count)
	return ids
}

func (s *Storage) LastUnconfirmedOrder() uint64 {
	order, _ := s.db.GetLastUnconfirmedOrder()
	return order
}

func (s *Storage) TransactionExist(txId string) bool {
	return s.db.TransactionExist(txId)
}
