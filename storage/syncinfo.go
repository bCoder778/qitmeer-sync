package storage

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

func (s *Storage) LastUnconfirmedOrder() uint64 {
	order, _ := s.db.GetLastUnconfirmedOrder()
	return order
}
