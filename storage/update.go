package storage

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/utils"
	"github.com/bCoder778/qitmeer-sync/verify"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"strings"
)

type blockData struct {
	block types.Block
	transactionData
}

type transactionData struct {
	Transactions []*types.Transaction
	Vins         []*types.Vin
	Vouts        []*types.Vout
	SpentedVouts []*types.Vout
	Transfers    []*types.Transfer
}

func (s *Storage) SaveBlock(rpcBlock *rpc.Block) error {
	block := s.crateBlock(rpcBlock)
	txData, err := s.createTransactions(rpcBlock.Transactions, rpcBlock.Order, rpcBlock.IsBlue)
	if err != nil {
		return err
	}
	if config.Setting.Verify.Version == "0.10" && rpcBlock.Order == 0 && rpcBlock.Height == 0 {
		coinMap := parseVoutCoinAmount(txData.Vouts)
		s.verify.Set10GenesisUTXO(coinMap)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateBlockDatas(block, txData.Transactions, txData.Vins, txData.Vouts, txData.SpentedVouts, txData.Transfers)
}

func (s *Storage) SaveTransaction(rpcTx *rpc.Transaction, order uint64, color int) error {
	txData, err := s.createTransactions([]rpc.Transaction{*rpcTx}, order, color)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if err := s.db.UpdateTransactionDatas(txData.Transactions, txData.Vins, txData.Vouts, txData.SpentedVouts, txData.Transfers); err != nil {
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
		s.db.UpdateTransactionStat(&tx)
	}
	return nil
}

func (s *Storage) crateBlock(rpcBlock *rpc.Block) *types.Block {
	miner := s.BlockMiner(rpcBlock)
	block := &types.Block{
		Id:            rpcBlock.Id,
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
		Color:         stat.Color(rpcBlock.IsBlue),
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
	vins := []*types.Vin{}
	vouts := []*types.Vout{}
	spentedVouts := []*types.Vout{}
	transfers := []*types.Transfer{}
	for _, rpcTx := range rpcTxs {
		addressInOut := NewAddressInOutMap()
		status := s.verify.TransactionStat(&rpcTx, color)
		var totalVin, totalVout, fees uint64
		for index, vin := range rpcTx.Vin {
			var (
				address string
				amount  uint64
			)
			if vin.Coinbase != "" {
				address = "coinbase"
			} else if vin.Txid != "0000000000000000000000000000000000000000000000000000000000000000" {
				vout, err := s.db.GetVout(vin.Txid, vin.Vout)
				if err != nil {
					return nil, fmt.Errorf("query txid %s, vout=%d failed!", vin.Txid, vin.Vout)
				}

				// 可能引用同一区块vout
				if vout.TxId == "" {
					if vout, err = s.finVout(vin.Txid, vin.Vout, vouts); err != nil {
						return nil, fmt.Errorf("query txid %s, vout=%d failed!", vin.Txid, vin.Vout)
					}
				}
				// 添加需要更新的被花费vout
				if status != stat.TX_Failed {
					vout.SpentTx = rpcTx.Txid
					spentedVouts = append(spentedVouts, vout)
				}

				// 添加新的vin
				address = vout.Address
				amount = vout.Amount
				totalVin += amount
				//fmt.Printf("address %s vin %d\n", address, amount)
				addressInOut.AddAddressIn(address, vout.CoinId, int64(amount))
				newVin := &types.Vin{
					TxId:      rpcTx.Txid,
					SpentedTx: vout.TxId,
					Order:     order,
					Address:   address,
					Vout:      vin.Vout,
					CoinId:    vout.CoinId,
					Amount:    amount,
					Number:    index,
					Sequence:  vin.Sequence,
					ScriptSig: &types.ScriptSig{
						Hex: vin.ScriptSig.Hex,
						Asm: vin.ScriptSig.Asm,
					},
					Confirmations: rpcTx.Confirmations,
					Stat:          status,
					Timestamp:     rpcTx.Timestamp.Unix(),
				}
				vins = append(vins, newVin)
			}
		}

		var err error
		// 添加新的vout
		for index, vout := range rpcTx.Vout {
			var lock uint64
			switch vout.ScriptPubKey.Type {
			case "pubkeyhash":

			case "cltvpubkeyhash":
				codes := strings.Split(vout.ScriptPubKey.Asm, " ")
				if len(codes) == 0 {
					return nil, fmt.Errorf("cltvpubkeyhash vout error,  %s", vout.ScriptPubKey.Asm)
				}
				lock, err = utils.LittleHexToUint64(codes[0])
				if err != nil {
					return nil, fmt.Errorf("little hex %s to uint64 error, %s", codes[0], err.Error())
				}
			default:
				continue
			}

			if vout.CoinID == "" {
				// 0.9的网络
				vout.CoinID = verify.PMEERID
			}
			newVout := &types.Vout{
				TxId:    rpcTx.Txid,
				Order:   order,
				Address: vout.ScriptPubKey.Addresses[0],
				Amount:  vout.Amount,
				CoinId:  vout.CoinID,
				Number:  index,
				ScriptPubKey: &types.ScriptPubKey{
					Asm:     vout.ScriptPubKey.Asm,
					Hex:     vout.ScriptPubKey.Hex,
					ReqSigs: vout.ScriptPubKey.ReqSigs,
					Type:    vout.ScriptPubKey.Type,
				},
				SpentTx:       "",
				Confirmations: rpcTx.Confirmations,
				Stat:          status,
				Timestamp:     rpcTx.Timestamp.Unix(),
				Lock:          lock,
			}
			totalVout += vout.Amount
			vouts = append(vouts, newVout)
			//fmt.Printf("address %s vout %d\n", newVout.Address, int64(newVout.Amount))
			addressInOut.AddAddressOut(newVout.Address, vout.CoinID, int64(newVout.Amount))
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
			Fees:          fees,
			Duplicate:     rpcTx.Duplicate,
			Stat:          status,
		}
		txs = append(txs, tx)

		// 创建地址交易信息

		addrChanges := addressInOut.AddressChange()

		for _, change := range addrChanges {
			transfers = append(transfers, &types.Transfer{
				TxId:          tx.TxId,
				Address:       change.Address,
				Confirmations: tx.Confirmations,
				CoinId:        change.CoinID,
				Txsvaild:      tx.Txsvaild,
				IsCoinbase:    tx.IsCoinbase,
				Change:        change.Change,
				Timestamp:     tx.Timestamp,
				Fees:          tx.Fees,
				Stat:          tx.Stat,
			})
		}
	}

	return &transactionData{txs, vins, vouts, spentedVouts, transfers}, nil
}

func (s *Storage) finVout(txId string, number int, vouts []*types.Vout) (*types.Vout, error) {
	for _, v := range vouts {
		if v.TxId == txId && v.Number == number {
			return v, nil
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

type AddressInOut struct {
	Address  string
	TotalIn  int64
	TotalOut int64
}

type Key struct {
	Address string
	CoinID  string
}

type AddressChange struct {
	Address string
	Change  int64
	CoinID  string
}

type AddressInOutMap struct {
	addrMap map[Key]*AddressInOut
}

func NewAddressInOutMap() *AddressInOutMap {
	return &AddressInOutMap{
		addrMap: make(map[Key]*AddressInOut),
	}
}

func (a *AddressInOutMap) AddAddressIn(address, coinID string, inAmount int64) {
	key := Key{address, coinID}
	inOut, ok := a.addrMap[key]
	if ok {
		inOut.TotalIn += inAmount
	} else {
		a.addrMap[key] = &AddressInOut{
			Address:  address,
			TotalIn:  inAmount,
			TotalOut: 0,
		}
	}
}

func (a *AddressInOutMap) AddAddressOut(address, coinID string, outAmount int64) {
	key := Key{address, coinID}
	inOut, ok := a.addrMap[key]
	if ok {
		inOut.TotalOut += outAmount
	} else {
		a.addrMap[key] = &AddressInOut{
			Address:  address,
			TotalIn:  0,
			TotalOut: outAmount,
		}
	}
}

func (a *AddressInOutMap) AddressChange() []*AddressChange {
	addrChanges := []*AddressChange{}
	for key, inOut := range a.addrMap {
		addrChanges = append(addrChanges, &AddressChange{
			Address: key.Address,
			CoinID:  key.CoinID,
			Change:  inOut.TotalOut - inOut.TotalIn,
		})
	}
	return addrChanges
}

func parseVoutCoinAmount(vouts []*types.Vout) map[string]uint64 {
	coinMap := map[string]uint64{}
	for _, vout := range vouts {
		if _, ok := coinMap[vout.CoinId]; ok {
			coinMap[vout.CoinId] += vout.Amount
		} else {
			coinMap[vout.CoinId] = vout.Amount
		}
	}
	return coinMap
}
