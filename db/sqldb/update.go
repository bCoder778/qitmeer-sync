package sqldb

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/storage/types"
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
		new(types.Vinout),
	); err != nil {
		return nil, err
	}

	return &DB{engine: engine}, nil
}

func ConnectSqlServer(conf *config.DB) (*DB, error) {
	path := fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s", conf.Address, conf.User, conf.Password, conf.DBName)
	engine, err := xorm.NewEngine("mssql", path)
	engine.ShowSQL(false)
	if err != nil {
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
	d.engine.DropTables("vinout")
	return nil
}

func (d *DB) UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}

	if err := updateBlock(sess, block); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := updateTransactions(sess, txs); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := updateSpentedVinouts(sess, spentedVouts); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := updateVinouts(sess, vinouts); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func (d *DB) UpdateTransactionDatas(txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}

	if err := updateTransactions(sess, txs); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := updateVinouts(sess, vinouts); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := updateSpentedVinouts(sess, spentedVouts); err != nil {
		if err := sess.Rollback(); err != nil {
			return fmt.Errorf("roll back failed! %s", err.Error())
		}
		return err
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func updateVinouts(sess *xorm.Session, vinouts []*types.Vinout) error {
	// 更新vinouts

	for _, vinout := range vinouts {
		queryVinout := &types.Vinout{}
		cols := []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`, `spented_tx`, `vout`, "confirmations", `sequence`, `script_sig`, `stat`}
		if ok, err := sess.Where("tx_id = ? and type = ? and number = ?", vinout.TxId, vinout.Type, vinout.Number).Get(queryVinout); err != nil {
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if vinout.SpentTx != "" {
				cols = []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`,
					`spent_tx`, `spent_number`, `spented_tx`, `vout`, "confirmations",
					`sequence`, `script_sig`, `stat`}
			} else if vinout.SpentTx == "" && vinout.UnconfirmedSpentTx != "" {
				cols = []string{`order`, `timestamp`, `address`, `amount`, `script_pub_key`,
					`unconfirmed_spent_tx`, `unconfirmed_spent_number`, `spented_tx`, `vout`,
					"confirmations", `sequence`, `script_sig`, `stat`}
			}

			if _, err := sess.Where("tx_id = ? and type = ? and number = ?", vinout.TxId, vinout.Type, vinout.Number).
				Cols(cols...).Update(vinout); err != nil {
				return err
			}
		} else {
			if _, err := sess.Insert(vinout); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateSpentedVinouts(sess *xorm.Session, vinouts []*types.Vinout) error {
	// 更新spentedVouts
	for _, vinout := range vinouts {
		if _, err := sess.Where("tx_id = ? and type = ? and number = ?", vinout.TxId, vinout.Type, vinout.Number).
			Cols("spent_tx", "spent_number", "unconfirmed_spent_tx", "unconfirmed_spent_number").Update(vinout); err != nil {
			return err
		}
	}
	return nil
}

func updateTransactions(sess *xorm.Session, txs []*types.Transaction) error {
	// 更新transaction
	queryTx := &types.Transaction{}
	for _, tx := range txs {
		queryTx.TxId = tx.TxId
		queryTx.BlockHash = tx.BlockHash
		if ok, err := sess.Exist(queryTx); err != nil {
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).
				Cols(`block_order`, `tx_hash`, `size`, `version`, `locktime`,
					`timestamp`, `expire`, `confirmations`, `txsvaild`, `is_coinbase`,
					`vins`, `vouts`, `total_vin`, `total_vout`, `fees`, `duplicate`,
					`stat`).Update(tx); err != nil {
				return err
			}
		} else {
			if _, err := sess.Insert(tx); err != nil {
				return err
			}
		}

	}
	return nil
}

func updateBlock(sess *xorm.Session, block *types.Block) error {
	// 更新block
	queryBlock := &types.Block{Hash: block.Hash}
	if ok, err := sess.Exist(queryBlock); err != nil {
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

func (d *DB) UpdateTransaction(tx *types.Transaction) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).
		Cols(`block_order`, `tx_hash`, `size`, `version`, `locktime`,
			`timestamp`, `expire`, `confirmations`, `txsvaild`, `is_coinbase`,
			`vins`, `vouts`, `total_vin`, `total_vout`, `fees`, `duplicate`,
			`stat`).
		Update(tx); err != nil {
		return err
	}
	return nil
}

func (d *DB) DeleteTransaction(tx *types.Transaction) error {
	var deleted types.Transaction
	_, err := d.engine.Id(tx.Id).Delete(&deleted)
	return err
}

func (d *DB) UpdateVinout(inout *types.Vinout) error {
	return nil
}
