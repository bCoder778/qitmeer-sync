package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer-sync/config"
	"github.com/bCoder778/qitmeer-sync/storage/types"
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	rpcAuth []*config.Rpc
	main *config.Rpc
}

func NewClient(auth []*config.Rpc) *Client {
	return &Client{rpcAuth: auth, main: auth[0]}
}

func (c *Client)TransactionStat(txid string, timestamp int64)(stat.TxStat){
	exist := false
	notConfirmed := false
	inBlock := false
	for _, auth := range c.rpcAuth{
		tx, err := c.getTransaction(txid, auth)
		if err != nil {
			notConfirmed = true
			if isNotExist(err){
				log.Warnf("%s, %s", auth.Host, err.Error())
				continue
			}
		}else{
			log.Debugf("%s, %s confirmations = %d", auth.Host,tx.Txid, tx.Confirmations)
			exist = true
			if tx.Confirmations >= stat.Tx_Confirmed_Value{
				return  stat.TX_Confirmed
			}
			if tx.Confirmations < 1{
				notConfirmed = true
			}
			if tx.BlockHash != ""{
				inBlock = true
			}
		}
	}
	if !exist{
		if time.Now().Unix() - timestamp > 60 * 60 {
			return stat.TX_Failed
		}
		return stat.TX_Unconfirmed
	}else{
		if notConfirmed{
			return stat.TX_Unconfirmed
		}else if inBlock{
			return stat.TX_Confirmed
		}else{
			return stat.TX_Memry
		}
	}
}


func isNotExist(err error) bool {
	if strings.Contains(err.Error(), "No information available about transaction") {
		return true
	}
	return false
}

func (c *Client) GetBlock(h uint64) (*Block, error) {
	return c.getBlock(h, c.main)
}

func (c *Client) getBlock(h uint64, auth *config.Rpc) (*Block, error) {
	params := []interface{}{h, true}
	resp := NewReqeust(params).SetMethod("getBlockByOrder").call(auth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) GetBlockByHash(hash string) (*Block, error) {
	return c.getBlockByHash(hash, c.main)
}

func (c *Client) getBlockByHash(hash string, auth *config.Rpc) (*Block, error) {
	params := []interface{}{hash, true}
	resp := NewReqeust(params).SetMethod("getBlock").call(auth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) GetBlockCount() (uint64, error) {
	return c.getBlockCount(c.main)
}

func (c *Client) getBlockCount(auth *config.Rpc) (uint64, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getBlockCount").call(auth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	count, err := strconv.ParseUint(string(resp.Result), 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Client) GetMainChainHeight() (uint64, error) {
	return c.getMainChainHeight(c.main)
}

func (c *Client) getMainChainHeight(auth *config.Rpc) (uint64, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getMainChainHeight").call(auth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	height, err := strconv.ParseUint(string(resp.Result), 10, 64)
	if err != nil {
		return 0, err
	}
	return height, nil
}

func (c *Client) SendTransaction(tx string) (string, error) {
	return c.sendTransaction(tx, c.main)
}

func (c *Client) sendTransaction(tx string, auth *config.Rpc) (string, error) {
	params := []interface{}{strings.Trim(tx, "\n"), true}
	resp := NewReqeust(params).SetMethod("sendRawTransaction").call(auth)
	if resp.Error != nil {
		return resp.Error.Message, errors.New(resp.Error.Message)
	}
	txid := ""
	json.Unmarshal(resp.Result, &txid)
	return txid, nil
}

func (c *Client) GetTransaction(txId string) (*Transaction, error) {
	return c.getTransaction(txId, c.main)
}


func (c *Client) getTransaction(txId string, auth *config.Rpc) (*Transaction, error) {
	params := []interface{}{txId, true}
	resp := NewReqeust(params).SetMethod("getRawTransaction").call(auth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs *Transaction
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetTransactionByBlockHash(txId string, hash string) (*Transaction, error) {
	return c.getTransactionByBlockHash(txId, hash, c.main)
}



func (c *Client) getTransactionByBlockHash(txId string, hash string, auth *config.Rpc) (*Transaction, error) {
	if hash == ""{
		tx, err := c.getTransaction(txId, auth)
		if err != nil{
			return nil, err
		}
		hash = tx.BlockHash
	}
	block, err := c.getBlockByHash(hash, auth)
	if err != nil{
		return nil, err
	}
	for _, tx := range block.Transactions{
		if tx.Txid == txId{
			tx.BlockOrder = block.Order
			tx.BlockHeight = block.Height
			return &tx, nil
		}
	}
	return nil, errors.New("not found")
}

func (c *Client) GetMemoryPool() ([]string, error) {
	return c.getMemoryPool(c.main)
}

func (c *Client) getMemoryPool(auth *config.Rpc) ([]string, error) {
	params := []interface{}{"", false}
	resp := NewReqeust(params).SetMethod("getMempool").call(auth)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []string
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) GetBlockById(id uint64) (*Block, error) {
	return c.getBlockById(id, c.main)
}

func (c *Client) getBlockById(id uint64, auth *config.Rpc) (*Block, error) {
	params := []interface{}{id, true}
	resp := NewReqeust(params).SetMethod("getBlockByID").call(auth)
	blk := new(Block)
	if resp.Error != nil {
		return blk, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, blk); err != nil {
		return blk, err
	}
	return blk, nil
}

func (c *Client) IsBlue(hash string) (int, error) {
	return c.isBlue(hash, c.main)
}


func (c *Client) isBlue(hash string, auth *config.Rpc) (int, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("isBlue").call(auth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}
	state, err := strconv.Atoi(string(resp.Result))
	if err != nil {
		return 0, err
	}
	return state, nil
}

func (c *Client) GetFees(hash string) (uint64, error) {
	return c.getFees(hash, c.main)
}

func (c *Client) getFees(hash string, auth *config.Rpc) (uint64, error) {
	params := []interface{}{hash}
	resp := NewReqeust(params).SetMethod("getFees").call(auth)
	if resp.Error != nil {
		return 0, errors.New(resp.Error.Message)
	}

	return strconv.ParseUint(string(resp.Result), 10, 64)
}

func (c *Client) GetPeerInfo() ([]PeerInfo, error) {
	var params []interface{}
	resp := NewReqeust(params).SetMethod("getPeerInfo").call(c.main)
	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}
	var rs []PeerInfo
	if err := json.Unmarshal(resp.Result, &rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func (c *Client) getPeerInfo(auth *config.Rpc) ([]PeerInfo, error) {
	return c.getPeerInfo(auth)
}

func (c *Client) GetCoins() ([]types.Coin, error) {
	return c.getCoins(c.main)
}

func (c *Client) getCoins(auth *config.Rpc) ([]types.Coin, error) {
	resp := NewReqeust(nil).SetMethod("getTokenInfo").call(auth)
	coins := []types.Coin{}
	if resp.Error != nil {
		return coins, errors.New(resp.Error.Message)
	}
	if err := json.Unmarshal(resp.Result, &coins); err != nil {
		return coins, err
	}
	return coins, nil
}

func (req *ClientRequest) call(auth *config.Rpc) *ClientResponse {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	defer client.CloseIdleConnections()

	//convert struct to []byte
	marshaledData, err := json.Marshal(req)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	httpRequest, err :=
		http.NewRequest(http.MethodPost, auth.Host, bytes.NewReader(marshaledData))
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}
	if httpRequest == nil {
		return &ClientResponse{Error: &Error{Message: "the httpRequest is nil"}}
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(auth.Admin, auth.Password)
	//log.Debugf("u:%s;p:%s", cfg.User, cfg.Pwd)

	response, err := client.Do(httpRequest)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	body := response.Body

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	resp := &ClientResponse{}
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	err = response.Body.Close()
	if err != nil {
		return &ClientResponse{Error: &Error{Message: err.Error()}}
	}

	return resp
}
