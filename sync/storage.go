package sync

import (
	"qitmeer-sync/rpc"
	"qitmeer-sync/storage/types"
)

type IStorage interface {
	ISyncInfo
	IUpdate
	IQueryBlock
	IQueryTransaction
	IList

	Close() error
}

type ISyncInfo interface {
	StartHeight() uint64
	UnconfirmedOrders() []uint64
}

type IUpdate interface {
	SaveBlock(block *rpc.Block) error
	SaveTransaction(tx *rpc.Transaction, order uint64, color int) error
	UpdateTxFailed(txId string) error
}

type IQueryBlock interface {
}

type IQueryTransaction interface {
	QueryMemTransaction() []types.Transaction
	QueryUnconfirmedTransaction() []types.Transaction
}

type IList interface {
}
