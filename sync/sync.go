package sync

import (
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage"
	"github.com/bCoder778/qitmeer-sync/verify"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"strings"
	"sync"
	"time"
)

const (
	waitBlockTime = 5
)

type QitmeerSync struct {
	storage          IStorage
	rpc              *rpc.Client
	mutex            sync.RWMutex
	reBlockSync      chan struct{}
	blockCh          chan *rpc.Block
	uncfmBlockCh     chan *rpc.Block
	uncfmTxCh        chan []rpc.Transaction
	interupt         chan struct{}
	wg               *sync.WaitGroup
	verifyFiledCount int
}

func NewQitmeerSync() (*QitmeerSync, error) {
	db, err := db.ConnectDB(config.Setting)
	if err != nil {
		return nil, err
	}
	ve := verify.NewQitmeerVerfiy(config.Setting.Verify, db)
	return &QitmeerSync{
		storage:     storage.NewStorage(db, ve),
		rpc:         rpc.NewClient(config.Setting.Rpc),
		reBlockSync: make(chan struct{}, 1),
		interupt:    make(chan struct{}, 1),
		wg:          &sync.WaitGroup{},
	}, nil
}

func (qs *QitmeerSync) Stop() {
	if qs.interupt != nil {
		close(qs.interupt)
	}
}

func (qs *QitmeerSync) Run() {
	qs.wg.Add(1)
	go qs.syncBlock()

	qs.wg.Add(1)
	go qs.syncTxPool()

	qs.wg.Add(1)
	go qs.updateUnconfirmedBlock()

	qs.wg.Add(1)
	go qs.updateUnconfirmedTransaction()

	qs.wg.Add(1)
	go qs.updateCoins()

	qs.wg.Wait()
	if err := qs.storage.Close(); err != nil {
		log.Errorf("Close storage failed! %s", err.Error())
	}
	log.Info("Stop qitmeer sync")
}

func (qs *QitmeerSync) syncBlock() {
	defer qs.wg.Done()

	wg := sync.WaitGroup{}
	for {
		select {
		default:
			log.Info("Start sync block")
			qs.initBlockCh()

			wg.Add(1)
			go qs.requestBlock(&wg)

			wg.Add(1)
			go qs.saveBlock(&wg)
			wg.Wait()
		case <-qs.interupt:
			log.Info("Shutdown sync block")
			return
		}
	}
}

func (qs *QitmeerSync) updateUnconfirmedBlock() {
	ticker1 := time.NewTicker(time.Second * 60 * 1)
	ticker2 := time.NewTicker(time.Second * 60 * 5)
	defer func() {
		ticker1.Stop()
		ticker2.Stop()
		qs.wg.Done()
	}()

	qs.initUncfmBlockCh()
	go qs.saveUnconfirmedBlock()

	for {
		select {
		case <-ticker1.C:
			log.Info("Start request unconfirmed block by count")
			qs.requestUnconfirmedBlockByCount(30)
		case <-ticker2.C:
			log.Info("Start request unconfirmed block")
			qs.requestUnconfirmedBlock()
		case <-qs.interupt:
			log.Info("Shutdown update unconfirmed block")
			return
		}
	}
}

func (qs *QitmeerSync) updateUnconfirmedTransaction() {
	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()
	qs.initUncfmTransactionCh()
	go qs.saveUnconfirmedTransaction()

	for {
		select {
		case <-ticker.C:
			if len(qs.uncfmTxCh) != 0 {
				continue
			}
			log.Info("Start request unconfirmed transaction")
			qs.requestUnconfirmedTransaction()
		case <-qs.interupt:
			log.Info("Shutdown update unconfirmed transaction")
			return
		}
	}
}

func (qs *QitmeerSync) syncTxPool() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()

	for {
		select {
		case <-qs.interupt:
			log.Info("Shutdown sync tx pool")
			return
		case <-ticker.C:
			//log.Info("Start sync tx pool")
			txIds, err := qs.rpc.GetMemoryPool()
			if err != nil {
				log.Warnf("Request getMemoryPool rpc failed! err:%v", err)
				continue
			}
			for _, txId := range txIds {
				select {
				case <-qs.interupt:
					log.Info("Shutdown sync tx pool when get transaction")
					return
				default:
					exist := qs.storage.TransactionExist(txId)
					if exist {
						continue
					}

					tx, err := qs.rpc.GetTransaction(txId)
					if err != nil {
						log.Warnf("Request getTransaction rpc failed! err:%v", err)
						continue
					}

					tx.Stat = int(stat.TX_Memry)
					if err := qs.storage.SaveTransaction(tx, 0, 0, 1); err != nil {
						//log.Mailf(config.Setting.Email.Title, "Sync tx pool to save transaction %v failed! err:%v", tx, err)
						continue
					}
				}
			}
		}
	}
}

func (qs *QitmeerSync) updateCoins() {
	ticker := time.NewTicker(time.Second * 60 * 10)

	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()

	for {
		select {
		case <-qs.interupt:
			log.Info("Shutdown update coin")
			return
		case <-ticker.C:
			log.Info("Start update coin")
			coins, _ := qs.rpc.GetCoins()
			qs.storage.UpdateCoins(coins)
		}
	}
}

func (qs *QitmeerSync) requestBlock(group *sync.WaitGroup) {
	defer group.Done()

	for {
		block0, err := qs.getBlockById(0)
		if err != nil {
			time.Sleep(time.Second * waitBlockTime)
			continue
		}
		qs.storage.Set10GenesisUTXO(block0)
		break
	}
	start := qs.storage.LastId()
	if start <= 2 {
		start = 0
	} else {
		start -= 2
	}
	var lastOrder uint64 = 0
	for {
		select {
		case <-qs.reBlockSync:
			log.Info("Stop and restart request block")
			return
		case <-qs.interupt:
			log.Info("Shutdown request block")
			return
		default:
			block, err := qs.getBlockById(start)
			if err != nil {
				if strings.Contains(err.Error(), "no node") {
					log.Debugf("no sys %d", start)
					start++
					continue
				} else if strings.Contains(err.Error(), "no block") {
					_, err = qs.rpc.GetBlock(lastOrder + 10)
					if err == nil {
						start++
					} else {
						log.Debugf("Request block order %d failed! %s", lastOrder+1, err.Error())
					}
					log.Debugf("Request block id %d failed! %s", start, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				} else {
					log.Debugf("Request block id %d failed! %s", start, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
			}
			start++
			if block.Order != 0 && block.Order > lastOrder {
				lastOrder = block.Order
			}
			qs.blockCh <- block
		}
	}
}

func (qs *QitmeerSync) saveBlock(group *sync.WaitGroup) {
	defer group.Done()

	for {
		select {
		case block := <-qs.blockCh:
			if err := qs.storage.SaveBlock(block); err != nil {
				log.Mailf(config.Setting.Email.Title, "Failed to save block %d %s, err:%s", block.Order, block.Hash, err.Error())
				qs.reBlockSync <- struct{}{}
				return
			}
			log.Infof("Save block %d", block.Order)
			if _, err := qs.storage.VerifyQitmeer(block); err != nil {
				//log.Mailf(config.Setting.Email.Title, "Failed to verify block %d %s, err:%v", block.Order, block.Hash, err)
				// 验证失败
				qs.verifyFiledCount++
				// 由于交易池中的交易会造成暂时的验证失败，所以当多次一直验证失败，才发送邮件
				if qs.verifyFiledCount >= 10 {
					log.Mailf(config.Setting.Email.Title, "Failed to verify block %d 10 times %s, err:%s", block.Order, block.Hash, err.Error())
					qs.verifyFiledCount = 0
				}
			}
		case <-qs.interupt:
			log.Info("Shutdown save block")
			return
		}
	}
}

func (qs *QitmeerSync) requestUnconfirmedBlockByCount(count int) {
	ids := qs.storage.UnconfirmedIdsByCount(count)
	for _, id := range ids {
		if id != 0 {
			select {
			case <-qs.interupt:
				log.Info("Shutdown request unconfirmed block")
				return
			default:
				block, err := qs.getBlockById(id)
				if err != nil {
					if strings.Contains(err.Error(), "no node") || strings.Contains(err.Error(), "no block") {
						log.Debugf("Request no node block id %d failed! %s", id, err.Error())
						continue
					}
					log.Debugf("Request block id %d failed! %s", id, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				qs.uncfmBlockCh <- block
			}
		}
	}
	log.Infof("Request unconfirmed block end")
}

func (qs *QitmeerSync) requestUnconfirmedBlock() {
	ids := qs.storage.UnconfirmedIds()
	for _, id := range ids {
		if id != 0 {
			select {
			case <-qs.interupt:
				log.Info("Shutdown request unconfirmed block")
				return
			default:
				block, err := qs.getBlockById(id)
				if err != nil {
					if strings.Contains(err.Error(), "no node") || strings.Contains(err.Error(), "no block") {
						log.Debugf("Request no node block id %d failed! %s", id, err.Error())
						continue
					}
					log.Debugf("Request block id %d failed! %s", id, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				qs.uncfmBlockCh <- block
			}
		}
	}
	log.Infof("Request unconfirmed block end")
}

func (qs *QitmeerSync) saveUnconfirmedBlock() {
	for block := range qs.uncfmBlockCh {
		select {
		case <-qs.interupt:
			log.Info("Shutdown save unconfirmed block")
			return
		default:
			if err := qs.storage.UpdateBlock(block); err != nil {
				log.Errorf(config.Setting.Email.Title, "Failed to save unconfirmed block %d %s, err:%v", block.Order, block.Hash, err)
			} else {
				log.Infof("Save unconfirmed block %d", block.Order)
			}
		}
	}
	log.Infof("Save unconfirmed block end")
}

func (qs *QitmeerSync) requestUnconfirmedTransaction() {
	hashMap := map[string]bool{}
	txs := qs.storage.QueryUnconfirmedTranslateTransaction()
	for _, tx := range txs {
		select {
		case <-qs.interupt:
			log.Info("Shutdown request unconfirmed transaction")
			return
		default:
			log.Infof("Get unconfirmed transaction %s", tx.TxId)
			/*stat, _, confirmations := qs.rpc.TransactionStat(tx.TxId, tx.Timestamp)
			qs.storage.UpdateTransactionStat(tx.TxId, confirmations, stat )*/
			if tx.Stat != stat.TX_Confirmed {
				txStat := qs.rpc.TransactionStat(tx.TxId, tx.Timestamp)
				if txStat == stat.TX_Failed {
					log.Infof("update tx  %s failed", tx.TxId)
					if err := qs.storage.UpdateTransactionStat(tx.TxId, 0, stat.TX_Failed); err != nil {
						log.Mailf("Failed to update transaction %s stat to failed!err %v", tx.TxId, err)
					}
					continue
				}
				tx.Stat = txStat
			}
			exist := hashMap[tx.BlockHash]
			if exist {
				continue
			}
			block, err := qs.rpc.GetBlockByHash(tx.BlockHash)
			if err != nil {
				continue
			}
			var rpcTxs []rpc.Transaction
			for _, rpcTx := range block.Transactions {
				if rpcTx.IsCoinbase() {
					continue
				}
				rpcTx.Stat = int(tx.Stat)
				rpcTx.BlockOrder = block.Order
				rpcTx.BlockHeight = block.Height
				rpcTxs = append(rpcTxs, rpcTx)
			}
			hashMap[tx.BlockHash] = true
			qs.uncfmTxCh <- rpcTxs
		}
	}
	log.Infof("Request unconfirmed transaction end")
}

func (qs *QitmeerSync) saveUnconfirmedTransaction() {
	for txs := range qs.uncfmTxCh {
		select {
		case <-qs.interupt:
			log.Info("Shutdown save unconfirmed transaction")
			return
		default:
			if len(txs) == 0 {
				continue
			}
			log.Infof("Save unconfirmed transaction in block %d %s", txs[0].BlockOrder, txs[0].BlockHash)
			if err := qs.storage.UpdateTransactions(txs, txs[0].BlockOrder, txs[0].BlockHash, txs[0].BlockHeight, 1); err != nil {
				log.Errorf(config.Setting.Email.Title, "Failed to save unconfirmed transaction block %d, err:%v", txs[0].BlockOrder, err)
			} else {
				log.Infof("Save unconfirmed transaction block %d", txs[0].BlockOrder)
			}
		}
	}
	log.Infof("Save unconfirmed transaction end")
}

func (qs *QitmeerSync) initBlockCh() {
	if qs.blockCh != nil {
		close(qs.blockCh)
	}
	qs.blockCh = make(chan *rpc.Block, 100)
}

func (qs *QitmeerSync) initUncfmBlockCh() {
	if qs.uncfmBlockCh != nil {
		close(qs.uncfmBlockCh)
	}
	qs.uncfmBlockCh = make(chan *rpc.Block, 1000)
}

func (qs *QitmeerSync) initUncfmTransactionCh() {
	if qs.uncfmTxCh != nil {
		close(qs.uncfmTxCh)
	}
	qs.uncfmTxCh = make(chan []rpc.Transaction, 100)
}

func isExist(err error) bool {
	if strings.Contains(err.Error(), "No information available about transaction") {
		return true
	}
	return false
}

func (qs *QitmeerSync) getBlockById(id uint64) (*rpc.Block, error) {
	block, err := qs.rpc.GetBlockById(id)
	if err != nil {
		return nil, err
	}
	color, err := qs.rpc.IsBlue(block.Hash)
	if err != nil {
		return nil, err
	}
	block.IsBlue = color
	block.Id = id
	return block, err
}

func (qs *QitmeerSync) getBlockByHash(hash string) (*rpc.Block, error) {
	block, err := qs.rpc.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	color, err := qs.rpc.IsBlue(block.Hash)
	if err != nil {
		return nil, err
	}
	block.IsBlue = color
	return block, err
}
