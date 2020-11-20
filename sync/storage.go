package sync

import (
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify"
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
	LastOrder() uint64
	UnconfirmedOrders() []uint64
	LastUnconfirmedOrder() uint64
}

type IUpdate interface {
	SaveBlock(block *rpc.Block) error
	SaveTransaction(tx *rpc.Transaction, order uint64, color int) error
	UpdateTransactionStat(txId string, stat verify.TxStat) error
}

type IQueryBlock interface {
}

type IQueryTransaction interface {
	QueryMemTransaction() []types.Transaction
	QueryUnconfirmedTranslateTransaction() []types.Transaction
}

type IList interface {
}
