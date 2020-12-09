package storage

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

type blockData struct {
	block types.Block
	transactionData
}

type transactionData struct {
	Transactions []*types.Transaction
	Vinouts      []*types.Vinout
	SpentedVouts []*types.Vinout
}

func (s *Storage) SaveBlock(rpcBlock *rpc.Block) error {
	block := s.crateBlock(rpcBlock)
	txData, err := s.createTransactions(rpcBlock.Transactions, rpcBlock.Order, rpcBlock.IsBlue)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateBlockDatas(block, txData.Transactions, txData.Vinouts, txData.SpentedVouts)
}

func (s *Storage) SaveTransaction(rpcTx *rpc.Transaction, order uint64, color int) error {
	txData, err := s.createTransactions([]rpc.Transaction{*rpcTx}, order, color)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.db.UpdateTransactionDatas(txData.Transactions, txData.Vinouts, txData.SpentedVouts); err != nil {
		return err
	}
	// 删除Mem交易
	for _, tx := range txData.Transactions {
		if tx.Stat != stat.TX_Memry {
			tx, _ := s.db.GetTransaction(tx.TxId, "")
			if tx.TxId != "" && tx.Stat == stat.TX_Memry {
				s.db.DeleteTransaction(tx)
			}
		}
	}

	// 删除历史余留Mem交易
	memTxs, err := s.db.QueryMemTransaction()
	for _, memTx := range memTxs {
		txs, _ := s.db.QueryTransactions(memTx.TxId)
		if len(txs) > 1 {
			for _, tx := range txs {
				if tx.Stat != stat.TX_Memry {
					s.db.DeleteTransaction(&memTx)
				}
			}
		}
	}
	return nil
}

func (s *Storage) UpdateTransactionStat(txId string, stat stat.TxStat) error {
	txs, err := s.db.QueryTransactions(txId)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, tx := range txs {
		tx.Stat = stat
		s.db.UpdateTransaction(&tx)
	}
	return nil
}

func (s *Storage) crateBlock(rpcBlock *rpc.Block) *types.Block {
	miner := s.BlockMiner(rpcBlock)
	block := &types.Block{
		Hash:          rpcBlock.Hash,
		Txvalid:       rpcBlock.Txsvalid,
		Confirmations: rpcBlock.Confirmations,
		Version:       rpcBlock.Version,
		Weight:        rpcBlock.Weight,
		Height:        rpcBlock.Height,
		TxRoot:        rpcBlock.TxRoot,
		Order:         rpcBlock.Order,
		Transactions:  len(rpcBlock.Transactions),
		StateRoot:     rpcBlock.StateRoot,
		Bits:          rpcBlock.Bits,
		Timestamp:     rpcBlock.Timestamp.Unix(),
		ParentRoot:    rpcBlock.ParentRoot,
		Stat:          s.verify.BlockStat(rpcBlock),
		Parents:       rpcBlock.Parents,
		Children:      rpcBlock.Children,
		Difficulty:    rpcBlock.Difficulty,
		PowName:       rpcBlock.Pow.PowName,
		PowType:       rpcBlock.Pow.PowType,
		Nonce:         rpcBlock.Pow.Nonce,
		Address:       miner.Address,
		Amount:        miner.Amount,
	}
	if rpcBlock.Pow.ProofData != nil {
		block.EdgeBits = rpcBlock.Pow.ProofData.EdgeBits
		block.CircleNonces = rpcBlock.Pow.ProofData.CircleNonces
	}
	return block
}

func (s *Storage) createTransactions(rpcTxs []rpc.Transaction, order uint64, color int) (*transactionData, error) {
	txs := []*types.Transaction{}
	vinouts := []*types.Vinout{}
	spentedVouts := []*types.Vinout{}

	for _, rpcTx := range rpcTxs {
		status := s.verify.TransactionStat(&rpcTx, color)
		var totalVin, totalVout, fees uint64
		for index, vin := range rpcTx.Vin {
			var (
				address string
				amount  uint64
			)
			if vin.Coinbase != "" {
				address = "coinbase"
			} else {
				vout, err := s.db.GetVout(vin.Txid, vin.Vout)
				if err != nil {
					return nil, fmt.Errorf("query txid%s, vout=%d failed!", vin.Txid, vin.Vout)
				}
				// 可能引用同一区块vout
				if vout.TxId == "" {
					if vout, err = s.finVout(vin.Txid, vin.Vout, vinouts); err != nil {
						return nil, fmt.Errorf("query txid%s, vout=%d failed!", vin.Txid, vin.Vout)
					}
				}
				// 添加需要更新的被花费vout
				if status == stat.TX_Confirmed {
					vout.SpentTx = rpcTx.Txid
					vout.SpentNumber = index
				} else if status == stat.TX_Unconfirmed || status == stat.TX_Memry {
					vout.UnconfirmedSpentTx = rpcTx.Txid
					vout.UnconfirmedSpentNumber = index
				}
				spentedVouts = append(spentedVouts, vout)

				// 添加新的vin
				address = vout.Address
				amount = vout.Amount
				totalVin += amount

				vinout := &types.Vinout{
					TxId:      rpcTx.Txid,
					SpentedTx: vout.TxId,
					Order:     order,
					Type:      stat.TX_Vin,
					Address:   address,
					Vout:      vin.Vout,
					Amount:    amount,
					Number:    index,
					Sequence:  vin.Sequence,
					ScriptSig: &types.ScriptSig{
						Hex: vin.ScriptSig.Hex,
						Asm: vin.ScriptSig.Asm,
					},
					ScriptPubKey:           &types.ScriptPubKey{},
					SpentTx:                "",
					SpentNumber:            0,
					UnconfirmedSpentTx:     "",
					UnconfirmedSpentNumber: 0,
					Confirmations:          rpcTx.Confirmations,
					Stat:                   status,
					Timestamp:              rpcTx.Timestamp.Unix(),
				}
				vinouts = append(vinouts, vinout)
			}
		}

		// 添加新的vout
		for index, vout := range rpcTx.Vout {
			vinout := &types.Vinout{
				TxId:      rpcTx.Txid,
				Order:     order,
				Type:      stat.TX_Vout,
				Address:   vout.ScriptPubKey.Addresses[0],
				Vout:      index,
				Amount:    vout.Amount,
				Number:    index,
				Sequence:  0,
				ScriptSig: &types.ScriptSig{},
				ScriptPubKey: &types.ScriptPubKey{
					Asm:     vout.ScriptPubKey.Asm,
					Hex:     vout.ScriptPubKey.Hex,
					ReqSigs: vout.ScriptPubKey.ReqSigs,
					Type:    vout.ScriptPubKey.Type,
				},
				SpentTx:                "",
				SpentNumber:            0,
				UnconfirmedSpentTx:     "",
				UnconfirmedSpentNumber: 0,
				Confirmations:          rpcTx.Confirmations,
				Stat:                   status,
				Timestamp:              rpcTx.Timestamp.Unix(),
			}
			totalVout += vout.Amount
			vinouts = append(vinouts, vinout)
		}
		if totalVin > totalVout {
			fees = totalVin - totalVout
		}
		tx := &types.Transaction{
			TxId:          rpcTx.Txid,
			TxHash:        rpcTx.Txhash,
			Size:          rpcTx.Size,
			Version:       rpcTx.Version,
			Locktime:      rpcTx.Locktime,
			Timestamp:     rpcTx.Timestamp.Unix(),
			Expire:        rpcTx.Expire,
			BlockHash:     rpcTx.BlockHash,
			BlockOrder:    order,
			Confirmations: rpcTx.Confirmations,
			Txsvaild:      rpcTx.Txsvalid,
			IsCoinbase:    s.verify.IsCoinBase(&rpcTx),
			Vins:          len(rpcTx.Vin),
			Vouts:         len(rpcTx.Vout),
			TotalVin:      totalVin,
			TotalVout:     totalVout,
			Fees:          fees,
			Duplicate:     rpcTx.Duplicate,
			Stat:          status,
		}
		txs = append(txs, tx)
	}

	return &transactionData{txs, vinouts, spentedVouts}, nil
}

func (s *Storage) finVout(txId string, vout int, vinouts []*types.Vinout) (*types.Vinout, error) {
	for _, vinout := range vinouts {
		if vinout.Type == stat.TX_Vout && vinout.TxId == txId && vinout.Number == vout {
			return vinout, nil
		}
	}
	return nil, fmt.Errorf("vout is not exist")
}

func (s *Storage) BlockMiner(rpcBlock *rpc.Block) *types.Miner {
	for _, tx := range rpcBlock.Transactions {
		if len(tx.Vin) == 1 {
			for _, vin := range tx.Vin {
				if vin.Coinbase != "" {
					return &types.Miner{Address: tx.Vout[0].ScriptPubKey.Addresses[0], Amount: tx.Vout[0].Amount}
				}
			}
		}
	}
	return &types.Miner{}
}
