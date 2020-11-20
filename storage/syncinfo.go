package storage

func (s *Storage) LastOrder() uint64 {
	order, _ := s.db.GetLastOrder()
	return order
}

func (s *Storage) UnconfirmedOrders() []uint64 {
	orders, _ := s.db.QueryUnConfirmedOrders()
	return orders
}

func (s *Storage) LastUnconfirmedOrder() uint64 {
	order, _ := s.db.GetLastUnconfirmedOrder()
	return order
}
