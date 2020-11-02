package client

import "encoding/json"

type GetInfoResult struct {
	Version         string  `json:"version"`
	Protocolversion int     `json:"protocolversion"`
	Walletversion   int     `json:"walletversion"`
	Balance         int     `json:"balance"`
	Blocks          int     `json:"blocks"`
	Timeoffset      int     `json:"timeoffset"`
	Connections     int     `json:"connections"`
	Proxy           string  `json:"proxy"`
	Difficulty      float64 `json:"difficulty"`
	Testnet         bool    `json:"testnet"`
	Keypoololdest   int     `json:"keypoololdest"`
	Keypoolsize     int     `json:"keypoolsize"`
	UnlockedUntil   int     `json:"unlocked_until"`
	Paytxfee        float64 `json:"paytxfee"`
	Relayfee        float64 `json:"relayfee"`
	Errors          string  `json:"errors"`
}

type GetMemoryInfoResult struct {
	Total       int `json:"total"`
	JsHeap      int `json:"jsHeap"`
	JsHeapTotal int `json:"jsHeapTotal"`
	NativeHeap  int `json:"nativeHeap"`
	External    int `json:"external"`
}

type ValidateAddressResult struct {
	IsValid        bool   `json:"isvalid"`
	Address        string `json:"address"`
	IsScript       bool   `json:"isscript"`
	IsSpendable    bool   `json:"isspendable"`
	WitnessVersion int    `json:"witness_version"`
	WitnessProgram string `json:"witness_program"`
}

type CreateMultisigResult struct {
	Address      string `json:"address"`
	RedeemScript string `json:"redeemScript"`
}

type GetBlockchainInfoResult struct {
	Chain                string                     `json:"chain"`
	Blocks               int                        `json:"blocks"`
	Headers              int                        `json:"headers"`
	BestBlockHash        string                     `json:"bestblockhash"`
	TreeRoot             string                     `json:"treeroot"`
	Difficulty           float64                    `json:"difficulty"`
	MedianTime           int                        `json:"mediantime"`
	VerificationProgress float64                    `json:"verificationprogress"`
	ChainWork            string                     `json:"chainwork"`
	Pruned               bool                       `json:"pruned"`
	SoftForks            map[string]json.RawMessage `json:"softforks"`
	Pruneheight          *int                       `json:"pruneheight"`
}

type RPCBlockWithoutTxsResponse struct {
	Hash              string   `json:"hash"`
	Confirmations     int      `json:"confirmations"`
	StrippedSize      int      `json:"strippedsize"`
	Size              int      `json:"size"`
	Weight            int      `json:"weight"`
	Height            int      `json:"height"`
	Version           int      `json:"version"`
	VersionHex        string   `json:"versionHex"`
	MerkleRoot        string   `json:"merkleroot"`
	WitnessRoot       string   `json:"witnessroot"`
	TreeRoot          string   `json:"treeroot"`
	ReservedRoot      string   `json:"reservedroot"`
	Mask              string   `json:"mask"`
	Coinbase          []string `json:"coinbase"`
	TxHashes          []string `json:"tx"`
	Time              int      `json:"time"`
	MedianTime        int      `json:"mediantime"`
	Bits              int      `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	ChainWork         string   `json:"chainwork"`
	PreviousBlockHash string   `json:"previousblockhash"`
	NextBlockHash     *string  `json:"nextblockhash"`
}

type RPCBlockWithTxsResponse struct {
	Hash              string  `json:"hash"`
	Confirmations     int     `json:"confirmations"`
	StrippedSize      int     `json:"strippedsize"`
	Size              int     `json:"size"`
	Weight            int     `json:"weight"`
	Height            int     `json:"height"`
	Version           int     `json:"version"`
	VersionHex        string  `json:"versionHex"`
	MerkleRoot        string  `json:"merkleroot"`
	WitnessRoot       string  `json:"witnessroot"`
	TreeRoot          string  `json:"treeroot"`
	ReservedRoot      string  `json:"reservedroot"`
	Mask              string  `json:"mask"`
	Txs               []RPCTx `json:"tx"`
	Time              int     `json:"time"`
	Mediantime        int     `json:"mediantime"`
	Bits              int     `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	PreviousBlockHash *string `json:"previousblockhash"`
	NextBlockHash     *string `json:"nextblockhash"`
}

type RPCTx struct {
	ID            string    `json:"txid"`
	Hash          string    `json:"hash"`
	Size          int       `json:"size"`
	VSize         int       `json:"vsize"`
	Version       int       `json:"version"`
	LockTime      int       `json:"locktime"`
	Vin           []RPCVin  `json:"vin"`
	VOut          []RPCVOut `json:"vout"`
	BlockHash     string    `json:"blockhash"`
	Confirmations int       `json:"confirmations"`
	Time          int       `json:"time"`
	BlockTime     int       `json:"blocktime"`
}

type RPCAddress struct {
	Version int    `json:"version"`
	Hash    string `json:"hash"`
}

type RPCVin struct {
	Coinbase    bool     `json:"coinbase"`
	Txid        string   `json:"txid"`
	Vout        int64    `json:"vout"`
	Txinwitness []string `json:"txinwitness"`
	Sequence    int      `json:"sequence"`
}

type RPCVOut struct {
	Value    float64    `json:"value"`
	N        int        `json:"n"`
	Address  RPCAddress `json:"address"`
	Covenant Covenant   `json:"covenant"`
}

type GetBlockHeaderResult struct {
	Hash              string  `json:"hash"`
	Confirmations     int     `json:"confirmations"`
	Height            int     `json:"height"`
	Version           int     `json:"version"`
	VersionHex        string  `json:"versionHex"`
	MerkleRoot        string  `json:"merkleroot"`
	WitnessRoot       string  `json:"witnessroot"`
	TreeRoot          string  `json:"treeroot"`
	ReservedRoot      string  `json:"reservedroot"`
	Mask              string  `json:"mask"`
	Time              int     `json:"time"`
	MedianTime        int     `json:"mediantime"`
	Bits              int     `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	ChainWork         string  `json:"chainwork"`
	PreviousBlockHash *string `json:"previousblockhash"`
	NextBlockHash     *string `json:"nextblockhash"`
}

type GetChainTipsResult struct {
	Height    int    `json:"height"`
	Hash      string `json:"hash"`
	Branchlen int    `json:"branchlen"`
	Status    string `json:"status"`
}
