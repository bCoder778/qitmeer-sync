package sqldb

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

func (d *DB) GetTransaction(txId string, blockHash string) (*types.Transaction, error) {
	tx := &types.Transaction{}
	_, err := d.engine.Table(new(types.Transaction)).Where("tx_id = ? and block_hash = ?", txId, blockHash).Get(tx)
	return tx, err
}

func (d *DB) GetVout(txId string, vout int) (*types.Vout, error) {
	vinout := &types.Vout{}
	_, err := d.engine.Where("tx_id = ? and number = ?", txId, vout).Get(vinout)
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

func (d *DB) GetConfirmedBlockCount() int64 {
	count, _ := d.engine.Table(new(types.Block)).Where("stat = ?", stat.Block_Confirmed).Count()
	return count
}

func (d *DB) GetAllUtxoAndBlockCount() (map[string]float64, int64, error) {
	sess := d.engine.NewSession()
	defer sess.Close()

	table := new(types.Vout)
	txIds := []string{}
	coinIds := []string{}
	utxos := map[string]float64{}
	sess.Table(table).Select("coin_id").GroupBy("coin_id").Find(&coinIds)
	if len(coinIds) == 0 {
		return nil, 0, fmt.Errorf("no coin")
	}

	for _, coinId := range coinIds {
		sess.Table(table).Select("tx_id").Where("coin_id = ?, stat = ?", coinId, stat.TX_Memry).Find(&txIds)
		params := []interface{}{""}
		for _, txId := range txIds {
			params = append(params, txId)
		}

		utxo, err := sess.In("spent_tx", params...).Or("spent_tx = ?", "").
			And("coin_id = ? and stat in (?, ?)", coinId, stat.TX_Confirmed, stat.TX_Unconfirmed).Sum(new(types.Vout), "amount")
		if err != nil {
			return nil, 0, err
		}
		utxos[coinId] = utxo
	}

	count, err := d.engine.Table(new(types.Block)).Where("stat in (?, ?)", stat.Block_Confirmed, stat.Block_Unconfirmed).Count()
	return utxos, count, err
}

func (d *DB) GetConfirmedUtxoAndBlockCount() (float64, int64, error) {
	sess := d.engine.NewSession()
	defer sess.Close()

	table := new(types.Vout)
	block := new(types.Block)
	var orderValue int64

	//select `order` from vout where confirmations < 720 ORDER BY `order` limit 1)

	ok, err := sess.Table(block).Select("`order`").Where("confirmations < ? and `order` <> ?", stat.Block_Confirmed_Value, 0).OrderBy("`order`").Limit(1).Get(&orderValue)

	if !ok {
		_, err = sess.Table(block).Select("`order`").Where("`order` <> ?", 0).Desc("`order`").Limit(1).Get(&orderValue)
	}
	// select * from vout where confirmations > 720 and `order` < orderValue order by `order` desc limit 1
	sess.Table(block).Select("`order`").Where("confirmations > ? and `order` < ?", stat.Block_Confirmed_Value, orderValue).Desc(`order`).Limit(1).Get(&orderValue)

	txIds := []string{}
	sess.Table(table).Select("tx_id").Where("`order` > ?", orderValue).Find(&txIds)

	params := []interface{}{}
	for _, txId := range txIds {
		params = append(params, txId)
	}

	utxo, err := sess.In("spent_tx", params...).Or("spent_tx = ?", "").
		And("`order` <= ? and stat = ?", orderValue, stat.TX_Confirmed).
		Sum(table, "amount")

	count, err := d.engine.Table(block).Where("stat = ? and `order` <= ?", stat.Block_Confirmed, orderValue).Count()
	return utxo, count, err
}
