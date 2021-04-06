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
		cols := []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`, `spented_tx`, `vout`, "confirmations", `sequence`, `script_sig`, `stat`}
		if ok, err := sess.Where("tx_id = ?  and number = ?", vin.TxId, vin.Number).Get(queryVin); err != nil {
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if queryVin.Stat != stat.TX_Confirmed {
				cols = []string{`order`, `timestamp`, `address`, `amount`,
					`spented_tx`, `vout`, "confirmations", `sequence`,
					`script_sig`, `stat`}

				if _, err := sess.Where("tx_id = ?  and number = ?", vin.TxId, vin.Number).
					Cols(cols...).Update(vin); err != nil {
					return err
				}
			}
		} else {
			if _, err := sess.Insert(vin); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateVouts(sess *xorm.Session, vouts []*types.Vout) error {
	// 更新vout

	for _, vout := range vouts {
		queryVout := &types.Vout{}
		cols := []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`, `spented_tx`, `vout`, "confirmations", `sequence`, `script_sig`, `stat`}
		if ok, err := sess.Where("tx_id = ?  and number = ?", vout.TxId, vout.Number).Get(queryVout); err != nil {
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if queryVout.Stat != stat.TX_Confirmed {
				if vout.SpentTx != "" {
					cols = []string{`order`, `timestamp`, `address`, `amount`,
						`script_pub_key`, `spent_tx`, "confirmations",
						`stat`}
				} else if vout.SpentTx == "" {
					cols = []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`,
						`spented_tx`, "confirmations", `sequence`, `script_sig`, `stat`}
				}

				if _, err := sess.Where("tx_id = ? and number = ?", vout.TxId, vout.Number).
					Cols(cols...).Update(vout); err != nil {
					return err
				}
			}
		} else {
			if _, err := sess.Insert(vout); err != nil {
				return err
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
		if ok, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).Get(queryTx); err != nil {
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if queryTx.Stat != stat.TX_Confirmed {
				if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).
					Cols(`block_order`, `tx_hash`, `size`, `version`, `locktime`,
						`timestamp`, `expire`, `confirmations`, `txsvaild`, `is_coinbase`,
						`vins`, `vouts`, `total_vin`, `total_vout`, `fees`, `duplicate`,
						`stat`).Update(tx); err != nil {
					return err
				}
			}

		} else {
			if _, err := sess.Insert(tx); err != nil {
				return err
			}
		}

	}
	return nil
}

func updateTransfers(sess *xorm.Session, transfers []*types.Transfer) error {
	// 更新transaction
	for _, tras := range transfers {
		queryTransfer := &types.Transfer{}
		if ok, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).Get(queryTransfer); err != nil {
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if queryTransfer.Stat != stat.TX_Confirmed {
				if _, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).
					Cols(`change`, `fees`, `confirmations`, `txsvaild`, `stat`).Update(tras); err != nil {
					return err
				}
			}
		} else {
			if _, err := sess.Insert(tras); err != nil {
				return err
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
				`difficulty`, `pow_name`, `pow_type`, `nonce`, `edge_bits`, `circle_nonces`, `address`,
				`amount`, `stat`).Update(block); err != nil {
			return err
		}

	} else {
		if _, err := sess.Insert(block); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) UpdateBlock(block *types.Block) error {
	return nil
}

func (d *DB) UpdateTransactionStat(tx *types.Transaction) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}
	if _, err := sess.Where("tx_id = ?", tx.TxId).
		Cols(`stat`).
		Update(tx); err != nil {
		return err
	}

	if _, err := sess.Where("tx_id = ?", tx.TxId).
		Cols(`stat`).
		Update(&types.Vin{TxId: tx.TxId, Stat: tx.Stat}); err != nil {
		return err
	}
	if _, err := sess.Where("tx_id = ?", tx.TxId).
		Cols(`stat`).
		Update(&types.Vout{TxId: tx.TxId, Stat: tx.Stat}); err != nil {
		return err
	}
	if _, err := sess.Where("tx_id = ?", tx.TxId).
		Cols(`stat`).
		Update(&types.Transfer{TxId: tx.TxId, Stat: tx.Stat}); err != nil {
		return err
	}
	if _, err := sess.Where("spent_tx = ?", tx.TxId).
		Cols("spent_tx").
		Update(&types.Vout{SpentTx: ""}); err != nil {
		return err
	}
	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func (d *DB) DeleteTransaction(tx *types.Transaction) error {
	var deleted types.Transaction
	_, err := d.engine.Id(tx.Id).Delete(&deleted)
	return err
}
