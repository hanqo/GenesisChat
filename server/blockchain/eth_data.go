package blockchain

// messages and the wrapper send to block chain.
type MsgToChain struct {
	from 		string
	user  		string
	version 	string
	chainID 	int32
	typ			string

	signedTx 	    *string
    requestTx		*MsgContractFunc
	call            *MsgCall
}

type MsgCall struct {
	contractAddr 	string
	contractFunc 	MsgContractFunc
	value           *int64
}

type MsgContractFunc struct{
	function 	 	 string
	inputs 		    []string
}

// messages and the wrapper send from block chain
type MsgFromChain struct{
	to 			string
	user  	    string
	version 	string
	chainID 	int32
	typ         string

	txReceipt   	*MsgTxReceipt
	txSent      	*MsgTxSent
	callReturn      *MsgCallReturn
	txInfo    		*MsgTxInfo
}

type MsgTxReceipt struct{
	confirmed   	bool
	txHash 			string
	gasUsed     	uint64
	contractAddr 	*string
}

type MsgTxSent struct{
	txHash 		 string
	gasPrice     int64
	nonce        uint64
	gasEstimated uint64
}

type MsgCallReturn struct {
	contractAddr    string
	function        string
	output          string
}

type MsgTxInfo	struct{
	function   			string
	gasPrice 			int64
	gasLimit    		uint64
	nonce  	 			uint64
	data     			[]byte
}

type TxPending struct{
	from 		string
	user  		string
	version 	string
	chainID 	int32

	txHash      string
}