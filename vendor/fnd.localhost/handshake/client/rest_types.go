package client

type NodeInfo struct {
	Version string      `json:"version"`
	Network string      `json:"network"`
	Chain   ChainInfo   `json:"chain"`
	Pool    PoolInfo    `json:"pool"`
	Mempool MempoolInfo `json:"mempool"`
	Time    TimeInfo    `json:"time"`
	Memory  MemoryInfo  `json:"memory"`
}

type StateInfo struct {
	Tx     int   `json:"tx"`
	Coin   int   `json:"coin"`
	Value  int64 `json:"value"`
	Burned int64 `json:"burned"`
}

type ChainInfo struct {
	Height   int       `json:"height"`
	Tip      string    `json:"tip"`
	TreeRoot string    `json:"treeRoot"`
	Progress int       `json:"progress"`
	State    StateInfo `json:"state"`
}

type PoolInfo struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	BrontideHost string `json:"brontideHost"`
	BrontidePort int    `json:"brontidePort"`
	Identitykey  string `json:"identitykey"`
	Agent        string `json:"agent"`
	Services     string `json:"services"`
	Outbound     int    `json:"outbound"`
	Inbound      int    `json:"inbound"`
}

type MempoolInfo struct {
	Tx       int `json:"tx"`
	Size     int `json:"size"`
	Claims   int `json:"claims"`
	Airdrops int `json:"airdrops"`
	Orphans  int `json:"orphans"`
}

type TimeInfo struct {
	Uptime   int `json:"uptime"`
	System   int `json:"system"`
	Adjusted int `json:"adjusted"`
	Offset   int `json:"offset"`
}

type MemoryInfo struct {
	Total       int `json:"total"`
	JsHeap      int `json:"jsHeap"`
	JsHeapTotal int `json:"jsHeapTotal"`
	NativeHeap  int `json:"nativeHeap"`
	External    int `json:"external"`
}

type MempoolRejectsFilterInfo struct {
	Items   int   `json:"items"`
	Size    int   `json:"size"`
	Entries int   `json:"entries"`
	N       int   `json:"n"`
	Limit   int   `json:"limit"`
	Tweak   int64 `json:"tweak"`
}

type RESTBlock struct {
	Hash         string        `json:"hash"`
	Height       int           `json:"height"`
	Depth        int           `json:"depth"`
	Version      int           `json:"version"`
	PrevBlock    string        `json:"prevBlock"`
	MerkleRoot   string        `json:"merkleRoot"`
	WitnessRoot  string        `json:"witnessRoot"`
	TreeRoot     string        `json:"treeRoot"`
	ReservedRoot string        `json:"reservedRoot"`
	Time         int           `json:"time"`
	Bits         int           `json:"bits"`
	Nonce        int           `json:"nonce"`
	ExtraNonce   string        `json:"extraNonce"`
	Mask         string        `json:"mask"`
	Txs          []Transaction `json:"txs"`
}

type Prevout struct {
	Hash  string `json:"hash"`
	Index int64  `json:"index"`
}

type Input struct {
	Prevout  Prevout  `json:"prevout"`
	Witness  []string `json:"witness"`
	Sequence int      `json:"sequence"`
	Address  *string  `json:"address"`
}

type Covenant struct {
	Type   int      `json:"type"`
	Action string   `json:"action"`
	Items  []string `json:"items"`
}

type Output struct {
	Value    int      `json:"value"`
	Address  string   `json:"address"`
	Covenant Covenant `json:"covenant"`
}

type Transaction struct {
	Hash        string   `json:"hash"`
	WitnessHash string   `json:"witnessHash"`
	Fee         int      `json:"fee"`
	Rate        int      `json:"rate"`
	Mtime       int      `json:"mtime"`
	Index       int      `json:"index"`
	Version     int      `json:"version"`
	Inputs      []Input  `json:"inputs"`
	Outputs     []Output `json:"outputs"`
	Locktime    int      `json:"locktime"`
	Hex         string   `json:"hex"`
}

type Coin struct {
	Version  int      `json:"version"`
	Height   int      `json:"height"`
	Value    int      `json:"value"`
	Address  string   `json:"address"`
	Covenant Covenant `json:"covenant"`
	Coinbase bool     `json:"coinbase"`
	Hash     string   `json:"hash"`
	Index    int      `json:"index"`
}
