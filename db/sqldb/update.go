package sqldb

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	//"github.com/xormplus/xorm"
	"strings"

	//_ "github.com/lunny/godbc"
	_ "github.com/denisenkom/go-mssqldb"
)

type DB struct {
	engine *xorm.Engine
}

func ConnectMysql(conf *config.DB) (*DB, error) {
	path := strings.Join([]string{conf.User, ":", conf.Password, "@tcp(", conf.Address, ")/", conf.DBName}, "")
	engine, err := xorm.NewEngine("mysql", path)
	if err != nil {
		return nil, err
	}
	//engine.ShowSQL(true)

	if err = engine.Sync2(
		new(types.Block),
		new(types.Transaction),
		new(types.Vin),
		new(types.Vout),
		new(types.Transfer),
		new(types.Coin),
	); err != nil {
		return nil, err
	}

	return &DB{engine: engine}, nil
}

func ConnectSqlServer(conf *config.DB) (*DB, error) {
	path := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", conf.Address, conf.User, conf.Password, conf.DBName)
	engine, err := xorm.NewEngine("mssql", path)

	if err != nil {
		return nil, err
	}
	engine.ShowSQL(false)
	if err = engine.Sync2(
		new(types.Block),
		new(types.Transaction),
		new(types.Vin),
		new(types.Vout),
		new(types.Transfer),
		new(types.Coin),
	); err != nil {
		return nil, err
	}
	return &DB{engine}, nil
}

func (d *DB) Close() error {
	return d.engine.Close()
}

func (d *DB) Clear() error {
	d.engine.DropTables("block")
	d.engine.DropTables("transaction")
	d.engine.DropTables("vin")
	d.engine.DropTables("vout")
	d.engine.DropTables("transfer")
	d.engine.DropTables("coin")
	return nil
}

func (d *DB) UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vins []*types.Vin, vouts []*types.Vout, spentedVouts []*types.Vout, transfers []*types.Transfer) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}

	if err := updateBlock(sess, block); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateTransactions(sess, txs); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateTransfers(sess, transfers); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateSpentedVinouts(sess, spentedVouts); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVins(sess, vins); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVouts(sess, vouts); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func (d *DB) UpdateTransactionDatas(txs []*types.Transaction, vins []*types.Vin, vouts []*types.Vout, spentedVouts []*types.Vout, transfers []*types.Transfer) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}

	if err := updateTransactions(sess, txs); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateTransfers(sess, transfers); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateSpentedVinouts(sess, spentedVouts); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVins(sess, vins); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVouts(sess, vouts); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func updateVins(sess *xorm.Session, vins []*types.Vin) error {
	// 更新vin

	for _, vin := range vins {
		queryVin := &types.Vin{}
		cols := []string{`order`, `timestamp`, `address`, `amount`, `spented_tx`, `vout`, `confirmations`}

		if ok, err := sess.Where("tx_id = ?  and number = ?", vin.TxId, vin.Number).Get(queryVin); err != nil {
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if queryVin.Stat != stat.TX_Confirmed {
				cols = append(cols, "stat")
			}
			if _, err := sess.Where("tx_id = ?  and number = ?", vin.TxId, vin.Number).
				Cols(cols...).Update(vin); err != nil {
				return err
			}
		} else {
			if _, err := sess.Insert(vin); err != nil {
				return fmt.Errorf("insert block vin, %s", err)
			}
		}
	}
	return nil
}

func updateVouts(sess *xorm.Session, vouts []*types.Vout) error {
	// 更新vout

	for _, vout := range vouts {
		queryVout := &types.Vout{}
		cols := []string{`order`, `height`, `timestamp`, `address`, `amount`,`script_pub_key`, `is_blue`,`vout`, `lock`}
		if ok, err := sess.Where("tx_id = ?  and number = ?", vout.TxId, vout.Number).Get(queryVout); err != nil {
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if vout.SpentTx != "" {
				cols = append(cols, `spent_tx`)
			}
			if queryVout.Stat != stat.TX_Confirmed {
				cols = append(cols, `stat`)
			}
			if queryVout.Confirmations < vout.Confirmations{
				cols = append(cols, `confirmations`)
			}

			if _, err := sess.Where("tx_id = ? and number = ?", vout.TxId, vout.Number).
				Cols(cols...).Update(vout); err != nil {
				return err
			}
		} else {
			if _, err := sess.Insert(vout); err != nil {
				return fmt.Errorf("insert block vout, %s", err)
			}
		}
	}
	return nil
}

func updateSpentedVinouts(sess *xorm.Session, vouts []*types.Vout) error {
	// 更新spentedVouts
	for _, vout := range vouts {
		if _, err := sess.Where("tx_id = ? and number = ?", vout.TxId, vout.Number).
			Cols("spent_tx").Update(vout); err != nil {
			return err
		}
	}
	return nil
}

func updateTransactions(sess *xorm.Session, txs []*types.Transaction) error {
	// 更新transaction
	for _, tx := range txs {
		queryTx := &types.Transaction{}


		cols := []string{`vin_amount`,`vout_amount`,`block_order`, `block_hash`, `tx_hash`, `expire`, `confirmations`, `txsvaild`, `duplicate`}

		if tx.BlockHash == "" || tx.Stat == stat.TX_Memry{
			if ok, err := sess.Where("tx_id = ?", tx.TxId).Get(queryTx); err != nil {
				return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
			} else if ok {
				return nil
			} else {
				if _, err := sess.Insert(tx); err != nil {
					return fmt.Errorf("insert transction %s error, %s", tx.TxId, err)
				}
			}
		}else{
			if ok, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, "").Get(queryTx); err != nil {
				return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
			} else if ok {
				cols = append(cols, "stat")
				if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, "").
					Cols(cols...).Update(tx); err != nil {
					return err
				}
			} else {
				if ok, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).Get(queryTx); err != nil {
					return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
				} else if ok {
					if queryTx.Stat != stat.TX_Confirmed{
						cols = append(cols, "stat")
					}
					if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).
						Cols(cols...).Update(tx); err != nil {
						return err
					}
				} else {
					if _, err := sess.Insert(tx); err != nil {
						return fmt.Errorf("insert transction %s error, %s", tx.TxId, err)
					}
				}
			}
		}
	}
	return nil
}

func updateTransfers(sess *xorm.Session, transfers []*types.Transfer) error {
	// 更新transaction
	cols := []string{`change`, `fees`, `confirmations`, `txsvaild`, `is_blue`}
	colsHasStat := append(cols, "stat")

	for _, tras := range transfers {
		queryTransfer := &types.Transfer{}
		if ok, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).Get(queryTransfer); err != nil {
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if queryTransfer.Stat == stat.TX_Confirmed {
				if _, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).
					Cols(cols...).Update(tras); err != nil {
					return err
				}
			}else{
				if _, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).
					Cols(colsHasStat...).Update(tras); err != nil {
					return err
				}
			}
		} else {
			if _, err := sess.Insert(tras); err != nil {
				return fmt.Errorf("insert transfer  error, %s", err)
			}
		}
	}
	return nil
}

func updateBlock(sess *xorm.Session, block *types.Block) error {
	// 更新block
	queryBlock := &types.Block{}
	if ok, err := sess.Where("hash = ?", block.Hash).Get(queryBlock); err != nil {
		return fmt.Errorf("faild to seesion exist block, %s", err.Error())
	} else if ok {
		if _, err := sess.Where("hash = ?", block.Hash).
			Cols(`txvalid`, `confirmations`, `version`, `weight`, `height`, `tx_root`, `order`,
				`transactions`, `state_root`, `bits`, `timestamp`, `parent_root`, `parents`, `children`,
				`difficulty`, `pow_name`, `pow_type`, `peer_id`, `nonce`, `edge_bits`, `circle_nonces`, `address`,
				`amount`, `color`, `stat`).Update(block); err != nil {
			return err
		}

	} else {
		// 更细为确认交易，无法获取blockID，会导致插入报错
		if block.Id == 0 && block.Height != 0 {
			return nil
		}
		if _, err := sess.Insert(block); err != nil {
			return fmt.Errorf("insert block error, %s", err)
		}
	}
	return nil
}

func (d *DB) UpdateBlock(block *types.Block) error {
	return nil
}

func (d *DB) UpdateCoin(coins []types.Coin) error {
	for _, coin := range coins {
		queryCoin := &types.Coin{}
		if ok, err := d.engine.Where("id = ?", coin.CoinId).Get(queryCoin); err != nil {
			return fmt.Errorf("faild to seesion exist block, %s", err.Error())
		} else if ok {
			return nil
		} else {

			if _, err := d.engine.Insert(coin); err != nil {
				return fmt.Errorf("insert block error, %s", err)
			}
		}
	}
	return nil
}

func (d *DB) UpdateTransactionStat(txId string, confirmations uint64, txStat stat.TxStat) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}
	if _, err := sess.Where("tx_id = ?", txId).
		Cols(`stat`, `confirmations`).
		Update(&types.Transaction{Stat:txStat, Confirmations: confirmations}); err != nil {
		return err
	}

	if _, err := sess.Where("tx_id = ?", txId).
		Cols(`stat`,  `confirmations`).
		Update(&types.Vin{TxId: txId, Stat: txStat, Confirmations: confirmations}); err != nil {
		return err
	}
	if _, err := sess.Where("tx_id = ?", txId).
		Cols(`stat`,  `confirmations`).
		Update(&types.Vout{TxId: txId, Stat: txStat, Confirmations: confirmations}); err != nil {
		return err
	}
	if _, err := sess.Where("tx_id = ?", txId).
		Cols(`stat`,  `confirmations`).
		Update(&types.Transfer{TxId: txId, Stat: txStat, Confirmations: confirmations}); err != nil {
		return err
	}
	if txStat == stat.TX_Failed{
		if _, err := sess.Where("spent_tx = ?", txId).
			Cols("spent_tx").
			Update(&types.Vout{SpentTx: ""}); err != nil {
			return err
		}
	}
	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

