package types

import (
	"github.com/bCoder778/qitmeer-sync/verify/stat"
	"time"
)

type Reorg struct {
	Id           uint64         `xorm:"bigint autoincr pk" json:"id"`
	OldOrder     uint64         `xorm:"bigint"`
	NewOrder     uint64         `xorm:"bigint"`
	Hash         string         `xorm:"varchar(64)"`
	OldMiner     string         `xorm:"varchar(65)"`
	NewMiner     string         `xorm:"varchar(65)"`
	OldStat      stat.BlockStat `xorm:"int"`
	NewStat      stat.BlockStat `xorm:"int"`
	Transactions []string       `xorm:"json"`
}

type ReorgV2 struct {
	Id            uint64    `xorm:"bigint autoincr pk" json:"id"`
	Order         uint64    `xorm:"bigint"`
	OldHash       string    `xorm:"varchar(64)"`
	NewHash       string    `xorm:"varchar(64)"`
	Confirmations int       `xorm:"int"`
	OldMiner      string    `xorm:"varchar(65)"`
	NewMiner      string    `xorm:"varchar(65)"`
	EvmHeight     uint64    `xorm:"bigint"`
	Timestamp     time.Time `xorm:"datetime"`
	Stat          int       `xorm:"int"`
}

type Block struct {
	Id            uint64         `xorm:"bigint pk" json:"id"`
	Hash          string         `xorm:"varchar(64) index" json:"hash"`
	Txvalid       bool           `xorm:"bool index" json:"txvalid"`
	Confirmations uint64         `xorm:"bigint index" json:"confirmation"`
	Version       uint32         `xorm:"int" json:"version"`
	Weight        uint64         `xorm:"bigint" json:"weight"`
	Height        uint64         `xorm:"bigint index" json:"height"`
	TxRoot        string         `xorm:"varchar(64)" json:"txroot"`
	Order         uint64         `xorm:"bigint index" json:"order"`
	EvmHeight     uint64         `xorm:"bigint" json:"evmHeight"`
	Transactions  int            `xorm:"int" json:"transactions"`
	StateRoot     string         `xorm:"varchar(64)" json:"stateroot"`
	Bits          string         `xorm:"varchar(64)" json:"bits"`
	Timestamp     int64          `xorm:"bigint index" json:"timestamp"`
	ParentRoot    string         `xorm:"varchar(64)" json:"parantroot"`
	Parents       []string       `xorm:"json" json:"parents"`
	Children      []string       `xorm:"json" json:"children"`
	Difficulty    uint64         `xorm:"bigint" json:"difficulty"`
	PowName       string         `xorm:"varchar(20) index" json:"powname"`
	PowType       int            `xorm:"int" json:"powtype"`
	Nonce         string         `xorm:"varchar(32)" json:"nonce"`
	EdgeBits      int            `xorm:"int" json:"edgebits"`
	CircleNonces  string         `xorm:"Text" json:"circlenonces"`
	Address       string         `xorm:"varchar(65) index" json:"address"`
	Amount        uint64         `xorm:"bigint" json:"amount"`
	PeerId        string         `xorm:"varchar(64)" json:"peerid"`
	Color         stat.Color     `xorm:"int index" json:"color"`
	Stat          stat.BlockStat `xorm:"int index" json:"stat"`
}

type Miner struct {
	Address string `xorm:"varchar(65)"`
	Amount  uint64 `xorm:"bigint"`
}

type Transaction struct {
	Id            uint64      `xorm:"bigint autoincr pk" json:"id"`
	TxId          string      `xorm:"varchar(64) index" json:"txid"`
	BlockHash     string      `xorm:"varchar(64) index" json:"blockhash"`
	BlockOrder    uint64      `xorm:"bigint index" json:"blockorder"`
	TxHash        string      `xorm:"varchar(64)" json:"txhash"`
	Size          int         `xorm:"int" json:"size"`
	Version       uint32      `xorm:"int" json:"version"`
	Locktime      uint64      `xorm:"bigint" json:"locktime"`
	Timestamp     int64       `xorm:"bigint index" json:"timestamp"`
	Expire        uint64      `xorm:"bigint" json:"expire"`
	Confirmations uint64      `xorm:"bigint index" json:"confirmations"`
	Txsvaild      bool        `xorm:"bool index" json:"txsvaild"`
	IsCoinbase    bool        `xorm:"bool index" json:"iscoinbase"`
	VinAmount     uint64      `xorm:"bigint" json:"vinamount"`
	VoutAmount    uint64      `xorm:"bigint" json:"voutamount"`
	VinAddress    string      `xorm:"varchar(65)" json:"vinaddress"`
	VoutAddress   string      `xorm:"varchar(65)" json:"voutaddress"`
	To            string      ``
	Vins          int         `xorm:"int" json:"vin"`
	Vouts         int         `xorm:"int" json:"vout"`
	Fees          uint64      `xorm:"bigint" json:"fees"`
	Duplicate     bool        `xorm:"bool index" json:"duplicate"`
	Stat          stat.TxStat `xorm:"int index" json:"stat"`
}

type Transfer struct {
	Id            uint64      `xorm:"bigint autoincr pk" json:"id"`
	TxId          string      `xorm:"varchar(64) index" json:"txid"`
	BlockHash     string      `xorm:"varchar(64) index" json:"block_hash"`
	Address       string      `xorm:"varchar(65) index" json:"address"`
	CoinId        string      `xorm:"varchar(30) index" json:"conid"`
	Confirmations uint64      `xorm:"bigint index" json:"confirmations"`
	Txsvaild      bool        `xorm:"bool index" json:"txsvaild"`
	IsCoinbase    bool        `xorm:"bool index" json:"iscoinbase"`
	IsBlue        bool        `xorm:"bool index" json:"isblue"`
	Change        int64       `xorm:"bigint index" json:"change"`
	Fees          uint64      `xorm:"bigint" json:"fees"`
	Timestamp     int64       `xorm:"bigint index" json:"timestamp"`
	Stat          stat.TxStat `xorm:"int index" json:"stat"`
	Duplicate     bool
}

type Vin struct {
	Id            uint64      `xorm:"bigint autoincr pk" json:"id"`
	TxId          string      `xorm:"varchar(64) index(queryvin)" json:"txid"`
	Number        int         `xorm:"int index" json:"number"`
	Order         uint64      `xorm:"bigint index" json:"order"`
	Timestamp     int64       `xorm:"bigint index" json:"timestamp"`
	Address       string      `xorm:"varchar(65) index" json:"address"`
	CoinId        string      `xorm:"varchar(5) index" json:"conid"`
	Amount        uint64      `xorm:"bigint" json:"amount"`
	SpentedTx     string      `xorm:"varchar(64)" json:"spentedtx"`
	Vout          int         `xorm:"int index(queryvin)" json:"vout"`
	Confirmations uint64      `xorm:"bigint index" json:"confirmations"`
	Sequence      uint64      `xorm:"bigint" json:"sequence"`
	ScriptSig     *ScriptSig  `xorm:"json" json:"scriptsig"`
	Stat          stat.TxStat `xorm:"int index" json:"stat"`
	Duplicate     bool
}

type Vout struct {
	Id            uint64        `xorm:"bigint autoincr pk" json:"id"`
	TxId          string        `xorm:"varchar(64) index" json:"txid"`
	Number        int           `xorm:"int index" json:"number"`
	Order         uint64        `xorm:"bigint index" json:"order"`
	Height        uint64        `xorm:"bigint index" json:"height"`
	Timestamp     int64         `xorm:"bigint index" json:"timestamp"`
	Address       string        `xorm:"varchar(65) index" json:"address"`
	PKAddress     string        `xorm:"varchar(65) index pk_address" json:"pkAddress"`
	EVMAddress    string        `xorm:"varchar(65) index evm_address" json:"evmAddress"`
	Amount        uint64        `xorm:"bigint index" json:"amount"`
	CoinId        string        `xorm:"varchar(30) index" json:"coinid"`
	IsCoinbase    bool          `xorm:"bool index" json:"iscoinbase"`
	IsBlue        bool          `xorm:"bool index" json:"isblue"`
	ScriptPubKey  *ScriptPubKey `xorm:"json" json:"scriptpubkey"`
	SpentTx       string        `xorm:"varchar(64) index" json:"spenttx"`
	Spented       bool          `xorm:"bool spented" json:"spented"`
	Confirmations uint64        `xorm:"bigint " json:"confirmations"`
	Stat          stat.TxStat   `xorm:"int " json:"stat"`
	Lock          uint64        `xorm:"bigint " json:"lock"`
	Duplicate     bool
}

type ScriptPubKey struct {
	Asm     string `json:"asm"`
	Hex     string `json:"hex"`
	ReqSigs int    `json:"reqsigs"`
	Type    string `json:"type"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Coin struct {
	CoinId   uint16 `xorm:"bigint pk id" json:"coinId"`
	CoinName string `xorm:"varchar(35)" json:"coinName"`
}
