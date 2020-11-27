package sqldb

import (
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

func (d *DB) GetTransaction(txId string, blockHash string) (*types.Transaction, error) {
	tx := &types.Transaction{}
	_, err := d.engine.Table(new(types.Transaction)).Where("tx_id = ? and block_hash = ?", txId, blockHash).Get(tx)
	return tx, err
}

func (d *DB) GetVout(txId string, vout int) (*types.Vinout, error) {
	vinout := &types.Vinout{}
	_, err := d.engine.Where("tx_id = ? and type = ? and number = ?", txId, stat.TX_Vout, vout).Get(vinout)
	return vinout, err
}

func (d *DB) GetLastOrder() (uint64, error) {
	var block = &types.Block{}
	_, err := d.engine.Table(new(types.Block)).Desc("order").Get(block)
	return block.Order, err
}

func (d *DB) GetLastUnconfirmedOrder() (uint64, error) {
	var block = &types.Block{}
	_, err := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Unconfirmed).OrderBy("`order`").Get(block)
	return block.Order, err
}

func (d *DB) GetConfirmedUtxo() float64 {
	amount, _ := d.engine.Where("spent_tx = ? and type = ? and stat = ? and confirmations > ?", "", stat.TX_Vout, stat.TX_Confirmed, stat.Block_Confirmed_Value).Sum(new(types.Vinout), "amount")
	return amount
}

func (d *DB) GetConfirmedBlockCount() int64 {
	count, _ := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Confirmed).Count()
	return count
}

func (d *DB) GetConfirmedUtxoAndBlockCount() (float64, int64, error) {
	sess := d.engine.NewSession()
	defer sess.Close()

	utxo, err := sess.Where("spent_tx = ? and type = ? and stat = ? and confirmations > ?", "", stat.TX_Vout, stat.TX_Confirmed, stat.Block_Confirmed_Value).Sum(new(types.Vinout), "amount")
	if err != nil {
		return 0, 0, err
	}
	count, err := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Confirmed).Count()
	return utxo, count, err
}
