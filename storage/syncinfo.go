package storage

func (s *Storage) StartHeight() uint64 {
	order, _ := s.db.GetLastOrder()
	return order
}

func (s *Storage) UnconfirmedOrders() []uint64 {
	orders, _ := s.db.QueryUnConfirmedOrders()
	return orders
}
