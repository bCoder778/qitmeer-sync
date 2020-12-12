package db

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db/sqldb"
	"github.com/bCoder778/qitmeer-sync/storage/types"
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
	UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout, transfers []*types.Transfer) error
	UpdateTransactionDatas(txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout, transfers []*types.Transfer) error

	UpdateBlock(block *types.Block) error
	UpdateTransaction(tx *types.Transaction) error
	UpdateVinout(inout *types.Vinout) error

	DeleteTransaction(tx *types.Transaction) error
}

type IGet interface {
	GetLastOrder() (uint64, error)
	GetLastUnconfirmedOrder() (uint64, error)
	GetTransaction(txId string, blockHash string) (*types.Transaction, error)
	GetVout(txId string, vout int) (*types.Vinout, error)
	GetConfirmedUtxo() float64
	GetConfirmedBlockCount() int64
	GetAllUtxoAndBlockCount() (float64, int64, error)
	GetConfirmedUtxoAndBlockCount() (float64, int64, error)
}

type IQuery interface {
	QueryUnconfirmedTranslateTransaction() ([]types.Transaction, error)
	QueryMemTransaction() ([]types.Transaction, error)
	QueryUnConfirmedOrders() ([]uint64, error)
	QueryTransactions(txId string) ([]types.Transaction, error)
}

type IList interface {
}

func ConnectDB(setting *config.Config) (IDB, error) {
	var (
		db  IDB
		err error
	)
	switch setting.DB.DBType {
	case "mysql":
		if db, err = sqldb.ConnectMysql(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	case "sqlserver":
		if db, err = sqldb.ConnectSqlServer(setting.DB); err != nil {
			return nil, fmt.Errorf("failed to connect mysql, error:%v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database %s", setting.DB.DBType)
	}
	return db, nil
}
