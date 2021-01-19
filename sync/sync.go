package sync

import (
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/db"
	"github.com/bCoder778/qitmeer-sync/rpc"
	"github.com/bCoder778/qitmeer-sync/storage"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"strings"
	"sync"
	"time"
)

const (
	waitBlockTime = 30
)

type QitmeerSync struct {
	storage          IStorage
	rpc              *rpc.Client
	mutex            sync.RWMutex
	reBlockSync      chan struct{}
	reUncfmBlockSync chan struct{}
	reUncfmTxSync    chan struct{}
	blockCh          chan *rpc.Block
	uncfmBlockCh     chan *rpc.Block
	uncfmTxBlockCh   chan *rpc.Block
	interupt         chan struct{}
	wg               sync.WaitGroup
	verifyFiledCount int
}

func NewQitmeerSync() (*QitmeerSync, error) {
	db, err := db.ConnectDB(config.Setting)
	if err != nil {
		return nil, err
	}
	ve := verify.NewQitmeerVerfiy(config.Setting.Verify, db)
	return &QitmeerSync{
		storage:          storage.NewStorage(db, ve),
		rpc:              rpc.NewClient(config.Setting.Rpc),
		reBlockSync:      make(chan struct{}, 1),
		reUncfmBlockSync: make(chan struct{}, 1),
		reUncfmTxSync:    make(chan struct{}, 1),
		interupt:         make(chan struct{}, 1),
		wg:               sync.WaitGroup{},
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
	go qs.dealFailedTransaction()

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
	ticker := time.NewTicker(time.Second * 60 * 5)
	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()

	wg := sync.WaitGroup{}
	for {
		select {
		case <-ticker.C:
			log.Info("Start sync unconfirmed block")
			qs.initUncfmBlockCh()

			wg.Add(1)
			go qs.requestUnconfirmedBlock(&wg)

			wg.Add(1)
			go qs.saveUnconfirmedBlock(&wg)

			wg.Wait()
		case <-qs.interupt:
			log.Info("Shutdown update unconfirmed block")
			return
		}
	}

}

func (qs *QitmeerSync) updateUnconfirmedTransaction() {
	ticker := time.NewTicker(time.Second * 30)
	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()

	wg := sync.WaitGroup{}
	for {
		select {
		case <-ticker.C:
			log.Info("Start sync unconfirmed transaction")
			qs.initUncfmTransactionCh()

			wg.Add(1)
			go qs.requestUnconfirmedTransaction(&wg)

			wg.Add(1)
			go qs.saveUnconfirmedTransaction(&wg)

			wg.Wait()
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
					tx, err := qs.rpc.GetTransaction(txId)
					if err != nil {
						log.Warnf("Request getTransaction rpc failed! err:%v", err)
						continue
					}
					if err := qs.storage.SaveTransaction(tx, 0, 1); err != nil {
						//log.Mailf(config.Setting.Email.Title, "Sync tx pool to save transaction %v failed! err:%v", tx, err)
						continue
					}
				}
			}
		}
	}
}

func (qs *QitmeerSync) dealFailedTransaction() {
	ticker := time.NewTicker(time.Second * 60 * 10)

	defer func() {
		ticker.Stop()
		qs.wg.Done()
	}()

	for {
		select {
		case <-qs.interupt:
			log.Info("Shutdown deal failed transaction")
			return
		case <-ticker.C:
			log.Info("Start deal failed transaction")
			memTxs := qs.storage.QueryMemTransaction()
			for _, tx := range memTxs {
				select {
				case <-qs.interupt:
					log.Info("Shutdown deal failed transaction when get transaction")
					return
				default:
					qs.dealTransaction(tx)
				}
			}
		}
	}
}

func (qs *QitmeerSync) dealTransaction(tx types.Transaction) {
	_, err := qs.rpc.GetTransaction(tx.TxId)
	if err != nil {
		if isExist(err) {
			if err := qs.storage.UpdateTransactionStat(tx.TxId, stat.TX_Failed); err != nil {
				log.Mailf("Failed to update transaction %s stat to failed!err %v", tx.TxId, err)
			}
		}
	}
}

func (qs *QitmeerSync) requestBlock(group *sync.WaitGroup) {
	defer group.Done()

	start := qs.storage.LastOrder()
	if start <= 5 {
		start = 0
	} else {
		start -= 5
	}
	for {
		select {
		case <-qs.reBlockSync:
			log.Info("Stop and restart request block")
			return
		case <-qs.interupt:
			log.Info("Shutdown request block")
			return
		default:
			block, err := qs.rpc.GetBlock(start)
			if err != nil {
				log.Debugf("Request block %d failed! %s", start, err.Error())
				time.Sleep(time.Second * waitBlockTime)
				continue
			}
			color, err := qs.rpc.IsBlue(block.Hash)
			if err != nil {
				log.Debugf("Request block %d isBlue failed! %s", start, err.Error())
				time.Sleep(time.Second * waitBlockTime)
				continue
			}
			block.IsBlue = color
			start++
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
				log.Mailf(config.Setting.Email.Title, "Failed to save block %d %s, err:%v", block.Order, block.Hash, err)
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
					log.Mailf(config.Setting.Email.Title, "Failed to verify block %d 10 times %s, err:%v", block.Order, block.Hash, err)
					qs.verifyFiledCount = 0
				}
			}
		case <-qs.interupt:
			log.Info("Shutdown save block")
			return
		}
	}
}

func (qs *QitmeerSync) requestUnconfirmedBlock(group *sync.WaitGroup) {
	defer group.Done()

	unOrder := qs.storage.LastUnconfirmedOrder()
	if unOrder != 0 {
		order := qs.storage.LastOrder()
		for ; unOrder <= order; unOrder++ {
			select {
			case <-qs.reUncfmBlockSync:
				log.Info("Stop and restart request unconfirmed block")
				return
			case <-qs.interupt:
				log.Info("Shutdown request unconfirmed block")
				return
			default:
				block, err := qs.rpc.GetBlock(unOrder)
				if err != nil {
					log.Debugf("Request block %d failed! %s", unOrder, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				color, err := qs.rpc.IsBlue(block.Hash)
				if err != nil {
					log.Debugf("Request block isBlue %d failed! %s", unOrder, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				block.IsBlue = color
				qs.uncfmBlockCh <- block
			}
		}
	}
	qs.reUncfmBlockSync <- struct{}{}
}

func (qs *QitmeerSync) saveUnconfirmedBlock(group *sync.WaitGroup) {
	defer group.Done()

	for {
		select {
		case block := <-qs.uncfmBlockCh:
			if err := qs.storage.SaveBlock(block); err != nil {
				log.Mailf(config.Setting.Email.Title, "Failed to save unconfirmed block %d %s, err:%v", block.Order, block.Hash, err)
				qs.reUncfmBlockSync <- struct{}{}
				return
			}
			log.Infof("Save unconfirmed block %d", block.Order)
		case <-qs.reUncfmBlockSync:
			log.Info("Stop and restart save unconfirmed block")
			return
		case <-qs.interupt:
			log.Info("Shutdown save unconfirmed block")
			return
		}
	}
}

func (qs *QitmeerSync) requestUnconfirmedTransaction(group *sync.WaitGroup) {
	defer group.Done()

	txs := qs.storage.QueryUnconfirmedTranslateTransaction()
	for _, tx := range txs {
		select {
		case <-qs.reUncfmTxSync:
			log.Info("Stop and restart request unconfirmed transaction")
			return
		case <-qs.interupt:
			log.Info("Shutdown request unconfirmed transaction")
			return
		default:
			rpcTx, err := qs.rpc.GetTransaction(tx.TxId)
			if err != nil {
				log.Debugf("Request getTransaction %d rpc failed! err:%v", tx.BlockOrder, err)
				time.Sleep(time.Second * waitBlockTime)
				continue
			}
			if rpcTx.BlockHash != "" {
				block, err := qs.rpc.GetBlockByHash(rpcTx.BlockHash)
				if err != nil {
					log.Debugf("Request getBlock %d rpc failed! err:%v", tx.BlockOrder, err)
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				color, err := qs.rpc.IsBlue(rpcTx.BlockHash)
				if err != nil {
					log.Debugf("Request block isBlue %d failed! %s", rpcTx.BlockOrder, err.Error())
					time.Sleep(time.Second * waitBlockTime)
					continue
				}
				// 重新更新整个块而不是单个交易，以免verify block 失败
				if block.Order < qs.storage.LastOrder() {
					block.IsBlue = color
					qs.uncfmTxBlockCh <- block
				}
			}
		}
	}
	qs.reUncfmTxSync <- struct{}{}
}

func (qs *QitmeerSync) saveUnconfirmedTransaction(group *sync.WaitGroup) {
	defer group.Done()

	for {
		select {
		case block := <-qs.uncfmTxBlockCh:
			if err := qs.storage.SaveBlock(block); err != nil {
				log.Mailf(config.Setting.Email.Title, "Failed to save unconfirmed transaction block %d, err:%v", block.Order, err)
				qs.reUncfmTxSync <- struct{}{}
				return
			}
		case <-qs.reUncfmTxSync:
			log.Info("Stop and restart save unconfirmed transaction")
			return
		case <-qs.interupt:
			log.Info("Shutdown save unconfirmed transaction")
			return
		}
	}
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
	qs.uncfmBlockCh = make(chan *rpc.Block, 100)
}

func (qs *QitmeerSync) initUncfmTransactionCh() {
	if qs.uncfmTxBlockCh != nil {
		close(qs.uncfmTxBlockCh)
	}
	qs.uncfmTxBlockCh = make(chan *rpc.Block, 100)
}

func isExist(err error) bool {
	if strings.Contains(err.Error(), "No information available about transaction") {
		return true
	}
	return false
}
