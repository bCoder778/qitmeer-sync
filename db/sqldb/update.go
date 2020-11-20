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
	engine.ShowSQL(false)

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
	for _, v := range d.engine.Tables {
		err := d.engine.DropTables(v.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) UpdateBlockDatas(block *types.Block, txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error {
	sess := d.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}

	// 更新block
	queryBlock := &types.Block{Hash: block.Hash}
	if ok, err := sess.Exist(queryBlock); err != nil {
		sess.Rollback()
		return fmt.Errorf("faild to seesion exist block, %s", err.Error())
	} else if ok {
		if _, err := sess.Where("hash = ?", block.Hash).Update(block); err != nil {
			sess.Rollback()
			return err
		}
	} else {
		if _, err := sess.Insert(block); err != nil {
			sess.Rollback()
			return err
		}
	}

	// 更新transaction
	queryTx := &types.Transaction{}
	for _, tx := range txs {
		queryTx.TxId = tx.TxId
		queryTx.BlockHash = tx.BlockHash
		if ok, err := sess.Exist(queryTx); err != nil {
			sess.Rollback()
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if _, err := sess.Where("tx_id = ? and block_hash = ?", tx.TxId, tx.BlockHash).Update(tx); err != nil {
				sess.Rollback()
				return err
			}
		} else {
			if _, err := sess.Insert(tx); err != nil {
				sess.Rollback()
				return err
			}
		}
	}

	// 更新spentedVouts
	for _, vinout := range spentedVouts {
		if _, err := sess.Where("tx_id = ? and type = ? and number = ?", vinout.TxId, vinout.Type, vinout.Number).
			Cols("spent_tx", "spent_number", "unconfirmed_spent_tx", "unconfirmed_spent_number").
			Update(vinout); err != nil {
			sess.Rollback()
			return err
		}

	}

	// 更新vinouts
	queryVinout := &types.Vinout{}
	for _, vinout := range vinouts {
		queryVinout.TxId = vinout.TxId
		queryVinout.Type = vinout.Type
		queryVinout.Number = vinout.Number
		if ok, err := sess.Exist(queryVinout); err != nil {
			sess.Rollback()
			return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
		} else if ok {
			if _, err := sess.Where("tx_id = ? and type = ? and number = ?", vinout.TxId, vinout.Type, vinout.Number).Update(vinout); err != nil {
				sess.Rollback()
				return err
			}
		} else {
			if _, err := sess.Insert(vinout); err != nil {
				sess.Rollback()
				return err
			}
		}
	}

	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	return nil
}

func (d *DB) UpdateTransactionDatas(txs []*types.Transaction, vinouts []*types.Vinout, spentedVouts []*types.Vinout) error {
	return nil
}

func (d *DB) UpdateBlock(block *types.Block) error {
	return nil
}

func (d *DB) UpdateTransaction(tx *types.Transaction) error {
	return nil
}

func (d *DB) UpdateVinout(inout *types.Vinout) error {
	return nil
}
