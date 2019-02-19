package blockchain

// messages and the wrapper send To block chain.
type MsgToChain struct {
	From    string
	User    string
	Version string
	ChainID int32
	Typ     string

	SignedTx  *string
	RequestTx *MsgContractFunc
	Call      *MsgCall
}

type MsgCall struct {
	ContractAddr string
	ContractFunc MsgContractFunc
	Value        *int64
}

type MsgContractFunc struct {
	Function string
	Inputs   []string
}

// messages and the wrapper send From block chain
type MsgFromChain struct {
	To      string
	User    string
	Version string
	ChainID int32
	Typ     string

	TxReceipt  *MsgTxReceipt
	TxSent     *MsgTxSent
	CallReturn *MsgCallReturn
	TxInfo     *MsgTxInfo
}

type MsgTxReceipt struct {
	Confirmed    bool
	TxHash       string
	GasUsed      uint64
	ContractAddr *string
}

type MsgTxSent struct {
	TxHash       string
	GasPrice     int64
	Nonce        uint64
	GasEstimated uint64
}

type MsgCallReturn struct {
	ContractAddr string
	Function     string
	Output       string
}

type MsgTxInfo struct {
	Function string
	GasPrice int64
	GasLimit uint64
	Nonce    uint64
	Data     []byte
}

type TxPending struct {
	From    string
	User    string
	Version string
	ChainID int32

	TxHash string
}
