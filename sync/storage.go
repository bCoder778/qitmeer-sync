package sync

import (
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

type IStorage interface {
	ISyncInfo
	IUpdate
	IQueryBlock
	IQueryTransaction
	IList
	IVerify
	Close() error
}

type ISyncInfo interface {
	LastOrder() uint64
	LastId() uint64
	UnconfirmedOrders() []uint64
	UnconfirmedIds() []uint64
	UnconfirmedIdsByCount(count int) []uint64
	LastUnconfirmedOrder() uint64
	TransactionExist(txId string) bool
}

type IUpdate interface {
	SaveBlock(block *rpc.Block) error
	SaveTransaction(tx *rpc.Transaction, order, height uint64, color int) error
	UpdateTransactionStat(txId string, confirmations uint64, stat stat.TxStat) error
	UpdateCoins(coins []types.Coin)
	Set10GenesisUTXO(rpcBlock *rpc.Block) error
}

type IQueryBlock interface {
}

type IQueryTransaction interface {
	QueryMemTransaction() []types.Transaction
	QueryUnconfirmedTranslateTransaction() []types.Transaction
}

type IList interface {
}

type IVerify interface {
	VerifyQitmeer(block *rpc.Block) (bool, error)
}
