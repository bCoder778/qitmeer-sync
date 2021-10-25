package sqldb

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"time"

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

func (d *DB) UpdateBlockDatas(block *types.Block, txs []*types.Transaction,
	vinsMap map[string][]*types.Vin, voutsMap map[string][]*types.Vout,
	spentedVouts map[string][]*types.Vout, transfersMap map[string][]*types.Transfer) error {
	sess := d.engine.NewSession()
	defer sess.Close()
	fmt.Printf("Start UpdateBlockDatas %d\n", time.Now().Unix())
	if err := sess.Begin(); err != nil {
		return fmt.Errorf("failed to seesion begin, %s", err.Error())
	}
	fmt.Printf("Start updateBlock %d\n", time.Now().Unix())
	if err := updateBlock(sess, block); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateBlock %d\n", time.Now().Unix())
	if err := updateTransactions(sess, txs); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateTransactions %d\n", time.Now().Unix())
	if err := updateTransfers(sess, transfersMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateTransfers %d\n", time.Now().Unix())
	if err := updateSpentedVinouts(sess, spentedVouts); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateSpentedVinouts %d\n", time.Now().Unix())
	if err := updateVins(sess, vinsMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateVins %d\n", time.Now().Unix())
	if err := updateVouts(sess, voutsMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}
	fmt.Printf("End updateVouts %d\n", time.Now().Unix())
	if err := sess.Commit(); err != nil {
		return fmt.Errorf("failed to seesion coimmit, %s", err.Error())
	}
	fmt.Printf("End UpdateBlockDatas %d\n", time.Now().Unix())
	return nil
}

func (d *DB) UpdateTransactionDatas(txs []*types.Transaction, vinsMap map[string][]*types.Vin,
	voutsMap map[string][]*types.Vout, spentedVoutsMap map[string][]*types.Vout,
	transfersMap map[string][]*types.Transfer) error {
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

	if err := updateTransfers(sess, transfersMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateSpentedVinouts(sess, spentedVoutsMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVins(sess, vinsMap); err != nil {
		if errR := sess.Rollback(); errR != nil {
			return fmt.Errorf("roll back failed! %s", errR.Error())
		}
		return err
	}

	if err := updateVouts(sess, voutsMap); err != nil {
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

func updateVins(sess *xorm.Session, vinsMap map[string][]*types.Vin) error {
	// 更新vin
	for txId, vins := range vinsMap{
		/*for _, vin := range vins {
			queryVin := &types.Vin{}
			cols := []string{`order`, `timestamp`, `address`, `amount`, `spented_tx`, `vout`, `confirmations`}

			if ok, err := sess.Where("tx_id = ?  and number = ?", vin.TxId, vin.Number).Get(queryVin); err != nil {
				return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
			} else if ok {
				if vin.Duplicate{
					continue
				}
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
		}*/

		if len(vins) == 0{
			continue
		}
		queryVin := &types.Vin{}
		if exist, err := sess.Table(types.Vin{}).Where("tx_id = ?", txId).Limit(1).Get(queryVin);err != nil{
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		}else if exist{
			cols := []string{`order`, `timestamp`, `confirmations`}
			if queryVin.Stat != stat.TX_Confirmed{
				cols = append(cols, "stat")
			}
			if _, err = sess.Table(types.Vout{}).Where("tx_id = ?", txId).Cols(cols...).Update(vins[0]);err != nil{
				return err
			}
		}else{
			if _, err = sess.InsertMulti(vins);err != nil{
				return err
			}
		}
	}

	return nil
}

func updateVouts(sess *xorm.Session, voutsMap map[string][]*types.Vout) error {
	// 更新vout
	for txId, vouts := range voutsMap{
		/*for _, vout := range vouts {

			queryVout := &types.Vout{}
			cols := []string{`order`, `height`, `timestamp`, `address`, `amount`,`script_pub_key`, `is_blue`,`vout`, `lock`}
			if ok, err := sess.Where("tx_id = ?  and number = ?", vout.TxId, vout.Number).Get(queryVout); err != nil {
				return fmt.Errorf("faild to seesion exist vinout, %s", err.Error())
			} else if ok {
				if vout.Duplicate{
					continue
				}
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
		}*/
		if len(vouts) == 0{
			continue
		}
		queryVout := &types.Vout{}
		if exist, err := sess.Table(types.Vout{}).Where("tx_id = ?", txId).Limit(1).Get(queryVout);err != nil{
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		}else if exist{
			cols := []string{`order`, `height`, `timestamp`, `is_blue`}
			if queryVout.Stat != stat.TX_Confirmed{
				cols = append(cols, "stat")
			}
			if queryVout.Confirmations < vouts[0].Confirmations{
				cols = append(cols, `confirmations`)
			}
			if _, err = sess.Table(types.Vout{}).Where("tx_id = ?", txId).Cols(cols...).Update(vouts[0]); err != nil{
				return err
			}
		}else{
			insertCount := 2000
			times := len(vouts) / insertCount
			lastCout := len(vouts) % insertCount
			if lastCout != 0{
				times++
			}
			for i := 0;i < times;i++{
				start := i * insertCount
				end := i * insertCount + insertCount
				if lastCout != 0 && i == times - 1{
					end = i * insertCount + lastCout
				}
				if i == times - 1{}
				if _, err = sess.Table(types.Vout{}).InsertMulti(vouts[start:end]);err != nil{
					return err
				}
			}
		}
	}
	return nil
}

func updateSpentedVinouts(sess *xorm.Session, voutsMap map[string][]*types.Vout) error {
	// 更新spentedVouts
	for spentTx, vouts := range voutsMap{
		if len(vouts) == 0{
			continue
		}
		ids := ""
		for i, vout := range vouts {
			if i == len(vouts) - 1{
				ids += fmt.Sprintf("%d", vout.Id)
			}else{
				ids += fmt.Sprintf("%d,", vout.Id)
			}
		}
		// select * from vout where tx_id
		if _, err := sess.Table(types.Vout{}).Where("find_in_set(id, ?)", ids).
			Update(map[string]string{
				"spent_tx":spentTx,
		}); err != nil {
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

func updateTransfers(sess *xorm.Session, transfersMap map[string][]*types.Transfer) error {
	// 更新transaction
	/*
	fmt.Println(len(transfers))
	for _, tras := range transfers {
		queryTransfer := &types.Transfer{}
		if ok, err := sess.Where("tx_id = ? and address = ? and coin_id = ?", tras.TxId, tras.Address, tras.CoinId).Get(queryTransfer); err != nil {
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		} else if ok {
			if tras.Duplicate{
				continue
			}
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
	}*/
	for txId, transfers := range transfersMap{
		if len(transfers) == 0{
			continue
		}
		queryTransfer := &types.Transfer{}
		if exist, err := sess.Table(types.Transfer{}).Where("tx_id = ?", txId).Limit(1).Get(queryTransfer);err != nil{
			return fmt.Errorf("faild to seesion exist tx, %s", err.Error())
		}else if exist{
			cols := []string{`confirmations`, `txsvaild`, `is_blue`}
			if queryTransfer.Stat != stat.TX_Confirmed{
				cols = append(cols, "stat")
			}
			if _, err = sess.Table(types.Transfer{}).Where("tx_id = ?", txId).Cols(cols...).Update(transfers[0]);err != nil{
				return err
			}
		}else{
			if _, err = sess.InsertMulti(transfers);err != nil{
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

