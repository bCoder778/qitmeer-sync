package verify

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/rpc"
)

type TxStat int
type BlockStat int
type TxType int

const (
	TX_Confirmed   TxStat = 0 // 已确认
	TX_Unconfirmed TxStat = 1 // 未确认
	TX_Memry       TxStat = 2 // 交易池
	TX_Failed      TxStat = 3 // 失败
)

const (
	Block_Confirmed   BlockStat = 0 // 已确认
	Block_Unconfirmed BlockStat = 1 // 未确认
	Block_InValid     BlockStat = 2 // 无效
	Block_Red         BlockStat = 3 // 红色
	Block_Duplicate   BlockStat = 4 // 红色
)

const (
	TX_Vin  TxType = 0
	TX_Vout TxType = 1
)

const (
	Block_Confirmed_Value = 720
	Tx_Confirmed_Value    = 10
)

const (
	BlockReward = 12000000000
	GenesisUTXO = 6524293004366634
)

type QitmeerVerify struct {
}

func NewQitmeerVerfiy() *QitmeerVerify {
	return &QitmeerVerify{}
}

func (qv *QitmeerVerify) BlockStat(block *rpc.Block) BlockStat {
	if block.Confirmations <= Block_Confirmed_Value {
		return Block_Unconfirmed
	}
	if !block.Txsvalid {
		return Block_InValid
	}

	switch block.IsBlue {
	case 0:
		return Block_Red
	case 1:
		// coinbase 是重复交易的情况
		if qv.isDuplicateCoinBase(block) {
			return Block_Duplicate
		}
		return Block_Confirmed
	case 2:
		return Block_Unconfirmed
	}
	return Block_Unconfirmed
}

func (qv *QitmeerVerify) TransactionStat(tx *rpc.Transaction, color int) TxStat {
	if qv.IsCoinBase(tx) {
		if tx.Confirmations <= Block_Confirmed_Value {
			return TX_Unconfirmed
		}
		if !tx.Txsvalid {
			return TX_Failed
		}

		switch color {
		case 0:
			return TX_Failed
		case 1:
			return TX_Confirmed
		case 2:
			return TX_Unconfirmed
		default:
			return TX_Unconfirmed
		}
	} else {
		if tx.BlockHash == "" {
			return TX_Memry
		}
		if tx.Confirmations <= Tx_Confirmed_Value {
			return TX_Unconfirmed
		} else {
			if !tx.Txsvalid {
				return TX_Failed
			}
			return TX_Confirmed
		}
	}
	return TX_Unconfirmed
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

func (qv *QitmeerVerify) VerifyAllAccount(utxo uint64, count int64) (bool, error) {
	should := (uint64(count)-1)*BlockReward + GenesisUTXO
	if should != utxo {
		return false, fmt.Errorf("all account %d is inconsistent with %d", utxo, should)
	}
	return true, nil
}
