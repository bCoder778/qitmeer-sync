package storage

import (
	"fmt"
	"qitmeer-sync/rpc"
	"qitmeer-sync/storage/types"
	"qitmeer-sync/verify"
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

	return s.db.UpdateTransactionDatas(txData.Transactions, txData.Vinouts, txData.SpentedVouts)
}

func (s *Storage) UpdateTxFailed(txId string) error {
	tx, err := s.db.GetTransaction(txId)
	if err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	tx.Stat = verify.TX_Failed
	return s.db.UpdateTransaction(tx)
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
		stat := s.verify.TransactionStat(&rpcTx, color)
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
				// 添加需要更新的被花费vout
				if stat == verify.TX_Confirmed {
					vout.SpentTx = rpcTx.Txid
					vout.SpentNumber = index
				} else {
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
					Type:      verify.TX_Vin,
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
					Stat:                   stat,
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
				Type:      verify.TX_Vout,
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
				Stat:                   stat,
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
			Stat:          stat,
		}
		txs = append(txs, tx)
	}

	return &transactionData{txs, vinouts, spentedVouts}, nil
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
