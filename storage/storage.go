package storage

import (
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify"
	"sync"
)

type IDB interface {
	IUpdate
	IGet
	IQuery
	IList
	Close() error
	Clear() error
}

type IUpdate interface {
	UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error
	UpdateTransactionDatas(txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error

	UpdateBlock(block *types.Block) error
	UpdateTransaction(tx *types.Transaction) error
	UpdateVinout(inout *types.Vinout) error
}

type IGet interface {
	GetLastOrder() (uint64, error)
	GetTransaction(txId string) (*types.Transaction, error)
	GetVout(txId string, vout int) (*types.Vinout, error)
}

type IQuery interface {
	QueryUnconfirmedTransaction() ([]types.Transaction, error)
	QueryMemTransaction() ([]types.Transaction, error)
	QueryUnConfirmedOrders() ([]uint64, error)
}

type IList interface {
}

type Storage struct {
	mutex  sync.RWMutex
	db     IDB
	verify *verify.QitmeerVerify
}

func NewStorage(db IDB) *Storage {
	return &Storage{db: db, verify: verify.NewQitmeerVerfiy()}
}

func (s *Storage) Close() error {
	return s.db.Close()
}
