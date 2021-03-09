package verify

import (
	"fmt"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/params"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

const (
	MEERID  = "MEER"
	QITID   = "QIT"
	PMEERID = "PMEER"
)

type QitmeerVerify struct {
	conf   *config.Verify
	db     db.IDB
	params *params.Params
}

func NewQitmeerVerfiy(conf *config.Verify, db db.IDB) *QitmeerVerify {
	para := params.Qitmeer9Params
	switch conf.Version {
	case "0.9":
		para = params.Qitmeer9Params
	case "0.10":
		para = params.Qitmeer10Params
	}
	return &QitmeerVerify{conf: conf, db: db, params: &para}
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

func (qv *QitmeerVerify) verifyCoinAllAccount(utxo uint64, count int64, coinId string) (bool, error) {
	switch coinId {
	case MEERID:
		fallthrough
	case PMEERID:
		should := (uint64(count)-1)*qv.params.BlockReward + qv.params.GenesisUTXO[coinId]
		if should != utxo {
			return false, fmt.Errorf("%d account %d is inconsistent with %d", coinId, utxo, should)
		}
		log.Infof("verify success, %s all utxo is %d", coinId, utxo)
		return true, nil
	default:
		should := qv.params.GenesisUTXO[coinId]
		if should != utxo {
			return false, fmt.Errorf("%d account %d is inconsistent with %d", coinId, utxo, should)
		}
		log.Infof("verify success, %s all utxo is %d", coinId, utxo)
		return true, nil
	}

}

func (qv *QitmeerVerify) verifyFees(block *rpc.Block) (bool, error) {
	mapIn := map[string]uint64{}
	mapOut := map[string]uint64{}
	mapFees := map[string]uint64{}
	if !block.Txsvalid {
		return true, nil
	}
	for _, tx := range block.Transactions {
		if !tx.Txsvalid {
			continue
		}
		if qv.IsCoinBase(&tx) {
			if tx.Vout[0].CoinID != MEERID {
				return false, fmt.Errorf("block %d, coinbase transaction %s, vout 0 is not meer", block.Order, tx.Txid)
			}
			for _, vout := range tx.Vout {
				if vout.CoinID == MEERID {
					mapFees[vout.CoinID] = vout.Amount - qv.params.BlockReward
				} else if vout.CoinID == "" {
					mapFees[PMEERID] = vout.Amount - qv.params.BlockReward
				} else {
					mapFees[vout.CoinID] = vout.Amount
				}
			}
		} else if !tx.Duplicate {
			for _, vin := range tx.Vin {
				utxo, err := qv.db.GetVout(vin.Txid, vin.Vout)
				if err != nil {
					return false, err
				}
				in, ok := mapIn[utxo.CoinId]
				if ok {
					mapIn[utxo.CoinId] = in + utxo.Amount
				} else {
					mapIn[utxo.CoinId] = utxo.Amount
				}
			}
			for _, vout := range tx.Vout {
				out, ok := mapOut[vout.CoinID]
				if ok {
					mapOut[vout.CoinID] = out + vout.Amount
				} else {
					mapOut[vout.CoinID] = vout.Amount
				}
			}
		}
	}
	for coinId, fees := range mapFees {
		totalIn := mapIn[coinId]
		totalOut := mapOut[coinId]
		calFees := totalIn - totalOut
		if fees != calFees {
			return false, fmt.Errorf("verify block %d %s fees is %d is inconsistent with %d", block.Order, coinId, fees, calFees)
		}
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
		utxos, count, err := qv.db.GetAllUtxoAndBlockCount()
		if err != nil {
			return false, err
		}
		for coinId, utxo := range utxos {
			if ok, err := qv.verifyCoinAllAccount(uint64(utxo), count, coinId); !ok {
				return false, err
			}
		}
	}
	if qv.conf.Fees {
		if ok, err := qv.verifyFees(rpcBlock); !ok {
			return false, err
		}
	}
	return true, nil
}
