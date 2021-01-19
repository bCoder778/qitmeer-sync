package storage

import "github.com/bCoder778/qitmeer-sync/rpc"

func (s *Storage) VerifyQitmeer(block *rpc.Block) (bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.verify.VerifyQitmeer(block)
}
