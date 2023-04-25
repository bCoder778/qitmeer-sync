package storage

import (
	"fmt"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/utils"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"strconv"
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

func (s *Storage) Set10GenesisUTXO(rpcBlock *rpc.Block) error {
	txData, err := s.createTransactions(rpcBlock.Transactions, rpcBlock.Timestamp.Unix(), rpcBlock.Order, rpcBlock.Height, rpcBlock.IsBlue, rpcBlock.Hash, false)
	if err != nil {
		return err
	}
	if config.Setting.Verify.Version == "0.10" && rpcBlock.Order == 0 && rpcBlock.Height == 0 {
		coinMap := parseVoutCoinAmount(txData.Vouts)
		s.verify.Set10GenesisUTXO(coinMap)
	}
	return nil
}

func (s *Storage) SaveBlock(rpcBlock *rpc.Block) error {
	block := s.crateBlock(rpcBlock)
	txData, err := s.createTransactions(rpcBlock.Transactions, rpcBlock.Timestamp.Unix(), rpcBlock.Order, rpcBlock.Height, rpcBlock.IsBlue, rpcBlock.Hash, false)
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

func (s *Storage) UpdateBlock(rpcBlock *rpc.Block) error {
	block := s.crateBlock(rpcBlock)
	txData, err := s.createTransactions([]rpc.Transaction{rpcBlock.Transactions[0]}, rpcBlock.Timestamp.Unix(), rpcBlock.Order, rpcBlock.Height, rpcBlock.IsBlue, rpcBlock.Hash, false)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateBlockDatas(block, txData.Transactions, txData.Vins, txData.Vouts, txData.SpentedVouts, txData.Transfers)
}

func (s *Storage) SaveTransaction(rpcTx *rpc.Transaction, order, height uint64, color int) error {
	txData, err := s.createTransactions([]rpc.Transaction{*rpcTx}, 0, order, height, color, rpcTx.BlockHash, true)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateTransactionDatas(txData.Transactions, txData.Vins, txData.Vouts, txData.SpentedVouts, txData.Transfers)
}

func (s *Storage) UpdateTransactions(rpcTxs []rpc.Transaction, order uint64, hash string, height uint64, color int) error {
	txData, err := s.createTransactions(rpcTxs, 0, order, height, color, hash, true)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateTransactionDatas(txData.Transactions, txData.Vins, txData.Vouts, txData.SpentedVouts, txData.Transfers)
}

func (s *Storage) UpdateTransactionStat(txId string, confirmations uint64, stat stat.TxStat) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.UpdateTransactionStat(txId, confirmations, stat)
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
		EvmHeight:     rpcBlock.EVMHeight,
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
		Nonce:         strconv.FormatUint(rpcBlock.Pow.Nonce, 10),
		Address:       miner.Address,
		PeerId:        rpcBlock.PeerId(),
		Amount:        miner.Amount,
	}
	if rpcBlock.Pow.ProofData != nil {
		block.EdgeBits = rpcBlock.Pow.ProofData.EdgeBits
		block.CircleNonces = rpcBlock.Pow.ProofData.CircleNonces
	}
	return block
}

func (s *Storage) createTransactions(rpcTxs []rpc.Transaction, blockTime int64, order, height uint64, color int, blockHash string, isTransfer bool) (*transactionData, error) {
	txs := []*types.Transaction{}
	vins := []*types.Vin{}
	vouts := []*types.Vout{}
	spentedVouts := []*types.Vout{}
	transfers := []*types.Transfer{}

	for _, rpcTx := range rpcTxs {
		vinAddress := ""
		voutAddress := ""
		voutPKAddress := ""
		voutEVMAddress := ""
		isCoinbase := s.verify.IsCoinBase(&rpcTx)
		addressInOut := NewAddressInOutMap()
		status := s.verify.TransactionStat(&rpcTx, color)
		if isTransfer {
			status = stat.TxStat(rpcTx.Stat)
		}

		var totalVin, totalVout, fees uint64
		var isEVM = false
		for index, vin := range rpcTx.Vin {
			var (
				address string
				amount  uint64
			)
			if vin.Type == "TxTypeCrossChainVM" {
				isEVM = true
			}
			if vin.Coinbase != "" {
				address = "coinbase"
				vinAddress = address
			} else if vin.Txid != "0000000000000000000000000000000000000000000000000000000000000000" && vin.Txid != "" {
				if vin.Type != "" {
					newVin := &types.Vin{
						TxId:    rpcTx.Txid,
						Number:  index,
						Order:   order,
						Address: vin.Type,
						ScriptSig: &types.ScriptSig{
							Hex: vin.ScriptSig.Hex,
							Asm: vin.ScriptSig.Asm,
						},
						Confirmations: rpcTx.Confirmations,
						Stat:          status,
						Timestamp:     rpcTx.Timestamp.Unix(),
						Duplicate:     rpcTx.Duplicate,
					}
					vins = append(vins, newVin)
				} else {
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

					if index == 0 {
						vinAddress = vout.Address
					}
					// 添加需要更新的被花费vout
					if status != stat.TX_Failed {
						vout.SpentTx = rpcTx.Txid
						vout.Spented = true
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
						Duplicate:     rpcTx.Duplicate,
					}
					vins = append(vins, newVin)
				}
			} else {
				vinAddress = "Token"
			}
		}

		var err error
		// 添加新的vout
		if isEVM {
			continue
		}
		for index, vout := range rpcTx.Vout {
			var lock uint64
			switch vout.ScriptPubKey.Type {
			case "pubkeyhash":
			case "pubkey":
			case "cltvpubkeyhash":
				codes := strings.Split(vout.ScriptPubKey.Asm, " ")
				if len(codes) == 0 {
					return nil, fmt.Errorf("cltvpubkeyhash vout error,  %s", vout.ScriptPubKey.Asm)
				}
				lock, err = utils.LittleHexToUint64(codes[0])
				if err != nil {
					return nil, fmt.Errorf("little hex %s to uint64 error, %s", codes[0], err.Error())
				}
			}
			if vout.ScriptPubKey.Addresses == nil {
				continue
			}
			if utils.IsPkAddress(vout.ScriptPubKey.Addresses[0]) {
				voutAddress, err = utils.PkAddressToAddress(vout.ScriptPubKey.Addresses[0])
				if err != nil {
					return nil, fmt.Errorf("wrong address %s, %s", vout.ScriptPubKey.Addresses[0], err.Error())
				}
				voutPKAddress = vout.ScriptPubKey.Addresses[0]
			} else {
				voutAddress = vout.ScriptPubKey.Addresses[0]
			}
			if vout.CoinID == "ETH" || vout.CoinID == "MEER Balance" {
				voutEVMAddress, err = utils.PkAddressToEVMAddress(vout.ScriptPubKey.Addresses[0])
				if err != nil {
					return nil, fmt.Errorf("wrong address %s, %s", vout.ScriptPubKey.Addresses[0], err.Error())
				}
			}
			if vout.CoinID == "MEER Asset" {
				vout.CoinID = "MEER"
			}
			newVout := &types.Vout{
				TxId:       rpcTx.Txid,
				Height:     height,
				Order:      order,
				Address:    voutAddress,
				PKAddress:  voutPKAddress,
				EVMAddress: voutEVMAddress,
				Amount:     vout.Amount,
				CoinId:     vout.CoinID,
				Number:     index,
				IsCoinbase: isCoinbase,
				IsBlue:     color == 1,
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
				Duplicate:     rpcTx.Duplicate,
			}
			totalVout += vout.Amount
			vouts = append(vouts, newVout)
			//fmt.Printf("address %s vout %d\n", newVout.Address, int64(newVout.Amount))
			addressInOut.AddAddressOut(newVout.Address, vout.CoinID, int64(newVout.Amount))
		}
		if totalVin > totalVout {
			fees = totalVin - totalVout
		}
		txTime := rpcTx.Timestamp.Unix()
		if blockTime != 0 && isCoinbase {
			txTime = blockTime
		}
		tx := &types.Transaction{
			TxId:          rpcTx.Txid,
			TxHash:        rpcTx.Txhash,
			Size:          rpcTx.Size,
			Version:       rpcTx.Version,
			Locktime:      rpcTx.Locktime,
			Timestamp:     txTime,
			Expire:        rpcTx.Expire,
			BlockHash:     rpcTx.BlockHash,
			BlockOrder:    order,
			Confirmations: rpcTx.Confirmations,
			Txsvaild:      rpcTx.Txsvalid,
			IsCoinbase:    isCoinbase,
			VinAmount:     totalVin,
			VoutAmount:    totalVout,
			VinAddress:    vinAddress,
			VoutAddress:   voutAddress,
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
				BlockHash:     blockHash,
				Address:       change.Address,
				Confirmations: tx.Confirmations,
				CoinId:        change.CoinID,
				Txsvaild:      tx.Txsvaild,
				IsCoinbase:    isCoinbase,
				IsBlue:        color == 1,
				Change:        change.Change,
				Timestamp:     tx.Timestamp,
				Fees:          tx.Fees,
				Stat:          tx.Stat,
				Duplicate:     rpcTx.Duplicate,
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
				if vin.Coinbase != "" && tx.Vout[0].ScriptPubKey.Addresses != nil {
					return &types.Miner{Address: tx.Vout[0].ScriptPubKey.Addresses[0], Amount: tx.Vout[0].Amount}
				}
			}
		}
	}
	return &types.Miner{}
}

func (s *Storage) UpdateCoins(coins []types.Coin) {
	s.db.UpdateCoin(coins)
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
