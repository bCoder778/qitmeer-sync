package cache

import "github.com/bCoder778/qitmeer-sync/storage/types"

type Cache struct {
	Blocks       []types.Block
	Transactions []types.Transaction
	Vinout       []types.VinOut
}

func NewCache() *Cache {
	return &Cache{
		Blocks:       []types.Block{},
		Transactions: []types.Transaction{},
		Vinout:       []types.VinOut{},
	}
}
