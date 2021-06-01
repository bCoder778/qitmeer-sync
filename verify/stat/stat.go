package stat

type TxStat int
type BlockStat int
type Color int

const (
	TX_Confirmed   TxStat = 0 // 已确认
	TX_Unconfirmed TxStat = 1 // 未确认
	TX_Memry       TxStat = 2 // 交易池
	TX_Failed      TxStat = 3 // 失败
)

const (
	Block_Confirmed   BlockStat = 0 // 已确认
	Block_Unconfirmed BlockStat = 1 // 未确认
	Block_InValid     BlockStat = 2 // 无效
	Block_Red         BlockStat = 3 // 红色
	Block_Duplicate   BlockStat = 4 // 重复
)

const (
	Block_Color_Red   BlockStat = 0 // 已确认
	Block_Color_Blue BlockStat = 1 // 未确认
	Block_Color_Un BlockStat = 2 // 未知
)


const (
	Block_Confirmed_Value = 720
	Tx_Confirmed_Value    = 10
)
