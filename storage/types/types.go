package types

import (
	"github.com/bCoder778/qitmeer-sync/verify/stat"
)

type Block struct {
	Id            uint64         `xorm:"bigint pk autoincr" json:"id"`
	Hash          string         `xorm:"varchar(64) index" json:"hash"`
	Txvalid       bool           `xorm:"bool" json:"txvalid"`
	Confirmations uint64         `xorm:"bigint" json:"confirmation"`
	Version       uint32         `xorm:"int" json:"version"`
	Weight        uint64         `xorm:"bigint" json:"weight"`
	Height        uint64         `xorm:"bigint" json:"height"`
	TxRoot        string         `xorm:"varchar(64)" json:"txroot"`
	Order         uint64         `xorm:"bigint" json:"order"`
	Transactions  int            `xorm:"int" json:"transactions"`
	StateRoot     string         `xorm:"varchar(64)" json:"stateroot"`
	Bits          string         `xorm:"varchar(64)" json:"bits"`
	Timestamp     int64          `xorm:"bigint" json:"timestamp"`
	ParentRoot    string         `xorm:"varchar(64)" json:"parantroot"`
	Parents       []string       `xorm:"json" json:"parents"`
	Children      []string       `xorm:"json" json:"children"`
	Difficulty    uint64         `xorm:"bigint" json:"difficulty"`
	PowName       string         `xorm:"varchar(20)" json:"powname"`
	PowType       int            `xorm:"int" json:"powtype"`
	Nonce         uint64         `xorm:"bigint" json:"nonce"`
	EdgeBits      int            `xorm:"int" json:"edgebits"`
	CircleNonces  string         `xorm:"Text" json:"circlenonces"`
	Address       string         `xorm:"varchar(40)" json:"address"`
	Amount        uint64         `xorm:"bigint" json:"amount"`
	Stat          stat.BlockStat `xorm:"int" json:"stat"`
}

type Miner struct {
	Address string `xorm:"varchar(40)"`
	Amount  uint64 `xorm:"bigint"`
}

type Transaction struct {
	Id            uint64      `xorm:"bigint autoincr pk" json:"id"`
	TxId          string      `xorm:"varchar(64) index" json:"txid"`
	BlockHash     string      `xorm:"varchar(64) index" json:"blockhash"`
	BlockOrder    uint64      `xorm:"bigint" json:"blockorder"`
	TxHash        string      `xorm:"varchar(64)" json:"txhash"`
	Size          int         `xorm:"int" json:"size"`
	Version       uint32      `xorm:"int" json:"version"`
	Locktime      uint64      `xorm:"bigint" json:"locktime"`
	Timestamp     int64       `xorm:"bigint" json:"timestamp"`
	Expire        uint64      `xorm:"bigint" json:"expire"`
	Confirmations uint64      `xorm:"bigint" json:"confirmations"`
	Txsvaild      bool        `xorm:"bool" json:"txsvaild"`
	IsCoinbase    bool        `xorm:"bool" json:"iscoinbase"`
	Vins          int         `xorm:"int" json:"vin"`
	Vouts         int         `xorm:"int" json:"vout"`
	TotalVin      uint64      `xorm:"bigint" json:"totalvin"`
	TotalVout     uint64      `xorm:"bigint" json:"totalvout"`
	Fees          uint64      `xorm:"bigint" json:"fees"`
	Duplicate     bool        `xorm:"bool" json:"duplicate"`
	Stat          stat.TxStat `xorm:"int" json:"stat"`
}

type Vinout struct {
	Id                     uint64        `xorm:"bigint autoincr pk" json:"id"`
	TxId                   string        `xorm:"varchar(64) index" json:"txid"`
	Type                   stat.TxType   `xorm:"int index" json:"type"`
	Number                 int           `xorm:"int index" json:"number"`
	Order                  uint64        `xorm:"bigint" json:"order"`
	Timestamp              int64         `xorm:"bigint" json:"timestamp"`
	Address                string        `xorm:"varchar(35) index" json:"address"`
	Amount                 uint64        `xorm:"bigint" json:"amount"`
	ScriptPubKey           *ScriptPubKey `xorm:"json" json:"scriptpubkey"`
	SpentTx                string        `xorm:"varchar(64)" json:"spenttx"`
	SpentNumber            int           `xorm:"int" json:"spentnumber"`
	UnconfirmedSpentTx     string        `xorm:"varchar(64)" json:"unconfirmedspenttx"`
	UnconfirmedSpentNumber int           `xorm:"int" json:"unconfirmedspentnumber"`
	SpentedTx              string        `xorm:"varchar(64) index(queryvin)" json:"spentedtx"`
	Vout                   int           `xorm:"int index(queryvin)" json:"vout"`
	Sequence               uint64        `xorm:"bigint" json:"sequence"`
	ScriptSig              *ScriptSig    `xorm:"json" json:"scriptsig"`
	Stat                   stat.TxStat   `xorm:"stat" json:"stat"`
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
