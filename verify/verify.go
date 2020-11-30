package verify

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

const (
	BlockReward = 12000000000
	GenesisUTXO = 6524293004366634
)

type QitmeerVerify struct {
	conf *config.Verify
	db   db.IDB
}

func NewQitmeerVerfiy(conf *config.Verify, db db.IDB) *QitmeerVerify {
	return &QitmeerVerify{conf: conf, db: db}
}

func (qv *QitmeerVerify) BlockStat(block *rpc.Block) stat.BlockStat {
	if block.Confirmations <= stat.Block_Confirmed_Value {
		return stat.Block_Unconfirmed
	}
	if !block.Txsvalid {
		return stat.Block_InValid
	}

	switch block.IsBlue {
	case 0:
		return stat.Block_Red
	case 1:
		// coinbase 是重复交易的情况
		if qv.isDuplicateCoinBase(block) {
			return stat.Block_Duplicate
		}
		return stat.Block_Confirmed
	case 2:
		return stat.Block_Unconfirmed
	}
	return stat.Block_Unconfirmed
}

func (qv *QitmeerVerify) TransactionStat(tx *rpc.Transaction, color int) stat.TxStat {
	if qv.IsCoinBase(tx) {
		if tx.Confirmations <= stat.Block_Confirmed_Value {
			return stat.TX_Unconfirmed
		}
		if !tx.Txsvalid {
			return stat.TX_Failed
		}

		switch color {
		case 0:
			return stat.TX_Failed
		case 1:
			return stat.TX_Confirmed
		case 2:
			return stat.TX_Unconfirmed
		default:
			return stat.TX_Unconfirmed
		}
	} else {
		if tx.BlockHash == "" {
			return stat.TX_Memry
		}
		if tx.Confirmations <= stat.Tx_Confirmed_Value {
			return stat.TX_Unconfirmed
		} else {
			if !tx.Txsvalid {
				return stat.TX_Failed
			}
			return stat.TX_Confirmed
		}
	}
	return stat.TX_Unconfirmed
}

func (qv *QitmeerVerify) IsCoinBase(rpcTx *rpc.Transaction) bool {
	return rpcTx.Vin[0].Coinbase != ""
}

func (qv *QitmeerVerify) isDuplicateCoinBase(block *rpc.Block) bool {
	for _, tx := range block.Transactions {
		if len(tx.Vin) == 1 {
			if tx.Vin[0].Coinbase != "" && tx.Duplicate {
				return true
			}
		}
	}
	return false
}

func (qv *QitmeerVerify) verifyAllAccount(utxo uint64, count int64) (bool, error) {
	should := (uint64(count)-1)*BlockReward + GenesisUTXO
	if should != utxo {
		return false, fmt.Errorf("all account %d is inconsistent with %d", utxo, should)
	}
	return true, nil
}

func (qv *QitmeerVerify) verifyFees(block *rpc.Block) (bool, error) {
	var coinbaseFees uint64
	var totalIn, totalOut uint64

	if !block.Txsvalid {
		return true, nil
	}
	for _, tx := range block.Transactions {
		if !tx.Txsvalid {
			continue
		}
		if qv.IsCoinBase(&tx) {
			coinbaseFees = tx.Vout[0].Amount - BlockReward
		} else if !tx.Duplicate {
			for _, vin := range tx.Vin {
				utxo, err := qv.db.GetVout(vin.Txid, vin.Vout)
				if err != nil {
					return false, err
				}
				totalIn += utxo.Amount
			}
			for _, vout := range tx.Vout {
				totalOut += vout.Amount
			}
		}
	}
	fees := totalIn - totalOut
	if coinbaseFees != fees {
		return false, fmt.Errorf("verify block %d coinbase fees is %d is inconsistent with %d", block.Order, coinbaseFees, fees)
	}
	return true, nil
}

func (qv *QitmeerVerify) VerifyQitmeer(rpcBlock *rpc.Block) (bool, error) {
	if rpcBlock.Order == 0 {
		return true, nil
	}
	if rpcBlock.Order%qv.conf.Interval != 0 {
		return true, nil
	}
	if qv.conf.UTXO {
		utxo, count, err := qv.db.GetAllUtxoAndBlockCount()
		if err != nil {
			return false, err
		}
		if ok, err := qv.verifyAllAccount(uint64(utxo), count); !ok {
			return false, err
		}
	}
	if qv.conf.Fees {
		if ok, err := qv.verifyFees(rpcBlock); !ok {
			return false, err
		}
	}
	return true, nil
}
