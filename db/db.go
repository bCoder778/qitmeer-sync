package db

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db/sqldb"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
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
	UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vins []*types.Vin, vouts []*types.Vout, spentedVouts []*types.Vout, transfers []*types.Transfer) error
	UpdateTransactionDatas(txs []*types.Transaction, vins []*types.Vin, vouts []*types.Vout, spentedVouts []*types.Vout, transfers []*types.Transfer) error

	UpdateBlock(block *types.Block) error
	UpdateTransactionStat(txId string, confirmations uint64, stat stat.TxStat) error
	UpdateCoin(coins []types.Coin) error
}

type IGet interface {
	GetLastOrder() (uint64, error)
	GetLastId() (uint64, error)
	GetLastUnconfirmedOrder() (uint64, error)
	GetTransaction(txId string, blockHash string) (*types.Transaction, error)
	GetVout(txId string, vout int) (*types.Vout, error)
	GetConfirmedBlockCount() int64
	GetAllUtxoAndBlockCount() (map[string]float64, int64, error)
	GetConfirmedUtxoAndBlockCount() (float64, int64, error)
	TransactionExist(txId string) bool
}

type IQuery interface {
	QueryUnconfirmedTranslateTransaction() ([]types.Transaction, error)
	QueryMemTransaction() ([]types.Transaction, error)
	QueryUnConfirmedOrders() ([]uint64, error)
	QueryUnConfirmedIds() ([]uint64, error)
	QueryUnConfirmedIdsByCount(count int)([]uint64, error)
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
