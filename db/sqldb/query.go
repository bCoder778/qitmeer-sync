package sqldb

import (
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

func (d *DB) QueryUnConfirmedOrders() ([]uint64, error) {
	orders := []uint64{}
	err := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Unconfirmed).Cols("order").Find(&orders)
	return orders, err
}

func (d *DB) QueryUnConfirmedIds() ([]uint64, error) {
	lastHeight ,_ := d.GetLastHeight()
	if lastHeight > 1000{
		lastHeight -= 1000
	}
	ids := []uint64{}
	err := d.engine.Table(new(types.Block)).Where("stat = ? and (height > ? or block.order != 0)", stat.Block_Unconfirmed, lastHeight).Cols("id").Find(&ids)
	return ids, err
}

func (d *DB) QueryUnConfirmedIdsByCount(count int) ([]uint64, error) {
	ids := []uint64{}
	err := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Unconfirmed).Cols("id").Desc(`id`).Limit(count).Find(&ids)
	return ids, err
}

func (d *DB) QueryUnconfirmedTranslateTransaction() ([]types.Transaction, error) {
	txs := []types.Transaction{}
	err := d.engine.Where("is_coinbase = ?", 0).And("confirmations < ? and stat in ( ?, ?)", stat.Tx_Confirmed_Value, stat.TX_Unconfirmed, stat.TX_Memry).Find(&txs)
	return txs, err
}

func (d *DB) QueryMemTransaction() ([]types.Transaction, error) {
	txs := []types.Transaction{}
	err := d.engine.Where("stat = ?", stat.TX_Memry).Find(&txs)
	return txs, err
}

func (d *DB) QueryTransactions(txId string) ([]types.Transaction, error) {
	txs := []types.Transaction{}
	err := d.engine.Where("tx_id = ?", txId).Find(&txs)
	return txs, err
}
