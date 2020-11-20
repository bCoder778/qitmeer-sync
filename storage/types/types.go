package types

import "qitmeer-sync/verify"

type Block struct {
	Id            uint64           `xorm:"bigint pk autoincr"`
	Hash          string           `xorm:"varchar(64) index"`
	Txvalid       bool             `xorm:"bool"`
	Confirmations uint64           `xorm:"bigint"`
	Version       uint32           `xorm:"int"`
	Weight        uint64           `xorm:"bigint"`
	Height        uint64           `xorm:"bigint"`
	TxRoot        string           `xorm:"varchar(64)"`
	Order         uint64           `xorm:"bigint"`
	Transactions  int              `xorm:"int"`
	StateRoot     string           `xorm:"varchar(64)"`
	Bits          string           `xorm:"varchar(64)"`
	Timestamp     int64            `xorm:"bigint"`
	ParentRoot    string           `xorm:"varchar(64)"`
	Parents       []string         `xorm:"json"`
	Children      []string         `xorm:"json"`
	Difficulty    uint64           `xorm:"bigint"`
	PowName       string           `xorm:"varchar(20)"`
	PowType       int              `xorm:"int"`
	Nonce         uint64           `xorm:"bigint"`
	EdgeBits      int              `xorm:"int"`
	CircleNonces  string           `xorm:"Text"`
	Address       string           `xorm:"varchar(40)"`
	Amount        uint64           `xorm:"bigint"`
	Stat          verify.BlockStat `xorm:"int"`
	UpdateVersion int              `xorm:"version"`
}

type Miner struct {
	Address string `xorm:"varchar(40)"`
	Amount  uint64 `xorm:"bigint"`
}

type Transaction struct {
	Id            uint64        `xorm:"bigint autoincr pk"`
	TxId          string        `xorm:"varchar(64) index"`
	BlockHash     string        `xorm:"varchar(64) index"`
	BlockOrder    uint64        `xorm:"bigint"`
	TxHash        string        `xorm:"varchar(64)"`
	Size          int           `xorm:"int"`
	Version       uint32        `xorm:"int"`
	Locktime      uint64        `xorm:"bigint"`
	Timestamp     int64         `xorm:"bigint"`
	Expire        uint64        `xorm:"bigint"`
	Confirmations uint64        `xorm:"bigint"`
	Txsvaild      bool          `xorm:"bool"`
	IsCoinbase    bool          `xorm:"bool"`
	Vins          int           `xorm:"int"`
	Vouts         int           `xorm:"int"`
	TotalVin      uint64        `xorm:"bigint"`
	TotalVout     uint64        `xorm:"bigint"`
	Fees          uint64        `xorm:"bigint"`
	Duplicate     bool          `xorm:"bool"`
	UpdateVersion int           `xorm:"version"`
	Stat          verify.TxStat `xorm:"int"`
}

type Vinout struct {
	Id                     uint64        `xorm:"bigint autoincr pk"`
	TxId                   string        `xorm:"varchar(64) index"`
	Type                   verify.TxType `xorm:"int index"`
	Number                 int           `xorm:"int index"`
	Order                  uint64        `xorm:"bigint"`
	Timestamp              int64         `xorm:"bigint"`
	Address                string        `xorm:"varchar(35) index"`
	Amount                 uint64        `xorm:"bigint"`
	ScriptPubKey           *ScriptPubKey `xorm:"json"`
	SpentTx                string        `xorm:"varchar(64)"`
	SpentNumber            int           `xorm:"int"`
	UnconfirmedSpentTx     string        `xorm:"varchar(64)"`
	UnconfirmedSpentNumber int           `xorm:"int"`
	SpentedTx              string        `xorm:"varchar(64) index(queryvin)"`
	Vout                   int           `xorm:"int index(queryvin)"`
	Sequence               uint64        `xorm:"bigint"`
	ScriptSig              *ScriptSig    `xorm:"json"`
	Stat                   verify.TxStat `xorm:"stat"`

	UpdateVersion int `xorm:"version"`
}

type ScriptPubKey struct {
	Asm     string
	Hex     string
	ReqSigs int
	Type    string
}

type ScriptSig struct {
	Asm string
	Hex string
}
