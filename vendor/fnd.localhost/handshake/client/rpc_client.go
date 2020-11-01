package client

import (
	"encoding/json"
	"strconv"
)

func (c *Client) RPCStop() error {
	var res json.RawMessage
	if err := c.executeRPC("stop", res); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCGetInfo() (*GetInfoResult, error) {
	res := new(GetInfoResult)
	if err := c.executeRPC("getinfo", res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetMemoryInfo() (*GetMemoryInfoResult, error) {
	res := new(GetMemoryInfoResult)
	if err := c.executeRPC("getmemoryinfo", res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCSetLogLevel(level string) error {
	if err := c.executeRPC("setloglevel", nil, level); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCValidateAddress(address string) (*ValidateAddressResult, error) {
	res := new(ValidateAddressResult)
	if err := c.executeRPC("validateaddress", res, address); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCCreateMultisig(numRequired int, keys []string) (*CreateMultisigResult, error) {
	res := new(CreateMultisigResult)
	if err := c.executeRPC("createmultisig", res, numRequired, keys); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCSignMessageWithPrivkey(privKey string, message string) (string, error) {
	var res string
	if err := c.executeRPC("signmessagewithprivkey", &res, privKey, message); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCVerifyMessage(address string, signature string, message string) (bool, error) {
	var res bool
	if err := c.executeRPC("verifymessage", res, address, signature, message); err != nil {
		return false, err
	}
	return res, nil
}

func (c *Client) RPCSetMockTime(timestamp int) error {
	if err := c.executeRPC("setmocktime", nil, strconv.Itoa(timestamp)); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCPruneBlockchain() error {
	if err := c.executeRPC("pruneblockchain", nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCInvalidateBlock(blockHash string) error {
	if err := c.executeRPC("invalidateblock", nil, blockHash); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCReconsiderBlock(blockHash string) error {
	if err := c.executeRPC("reconsiderblock", nil, blockHash); err != nil {
		return err
	}
	return nil
}

func (c *Client) RPCGetBlockchainInfo() (*GetBlockchainInfoResult, error) {
	res := new(GetBlockchainInfoResult)
	if err := c.executeRPC("getblockchaininfo", res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBestBlockHash() (string, error) {
	var res string
	if err := c.executeRPC("getbestblockhash", &res); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCGetBlockCount() (int, error) {
	var res int
	if err := c.executeRPC("getblockcount", &res); err != nil {
		return 0, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockByHashWithoutTxs(blockHash string) (*RPCBlockWithoutTxsResponse, error) {
	res := new(RPCBlockWithoutTxsResponse)
	if err := c.executeRPC("getblock", res, blockHash, true, false); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockByHashWithTxs(blockHash string) (*RPCBlockWithTxsResponse, error) {
	res := new(RPCBlockWithTxsResponse)
	if err := c.executeRPC("getblock", res, blockHash, true, true); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockHexByHash(blockHash string) (string, error) {
	var res string
	if err := c.executeRPC("getblock", &res, blockHash, false, false); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCGetBlockByHeightWithoutTxs(blockHeight int) (*RPCBlockWithoutTxsResponse, error) {
	res := new(RPCBlockWithoutTxsResponse)
	if err := c.executeRPC("getblockbyheight", res, blockHeight, true, false); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockByHeightWithTxs(blockHeight int) (*RPCBlockWithTxsResponse, error) {
	res := new(RPCBlockWithTxsResponse)
	if err := c.executeRPC("getblockbyheight", res, blockHeight, true, true); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockHexByHeight(blockHeight int) (string, error) {
	var res string
	if err := c.executeRPC("getblockbyheight", &res, blockHeight, false, false); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCGetBlockHashByHeight(blockHeight int) (string, error) {
	var res string
	if err := c.executeRPC("getblockhash", &res, blockHeight); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCGetBlockHeaderByHash(blockHash string) (*GetBlockHeaderResult, error) {
	res := new(GetBlockHeaderResult)
	if err := c.executeRPC("getblockheader", res, blockHash, true); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetBlockHeaderHexByHash(blockHash string) (string, error) {
	var res string
	if err := c.executeRPC("getblockheader", &res, blockHash, false); err != nil {
		return "", err
	}
	return res, nil
}

func (c *Client) RPCGetChainTips() ([]*GetChainTipsResult, error) {
	res := make([]*GetChainTipsResult, 0)
	if err := c.executeRPC("getchaintips", res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) RPCGetDifficulty() (float64, error) {
	var res float64
	if err := c.executeRPC("getdifficulty", &res); err != nil {
		return 0, err
	}
	return res, nil
}

func (c *Client) RPCGetNameByHash(hash string) (*string, error) {
	var res *string
	if err := c.executeRPC("getnamebyhash", &res, hash); err != nil {
		return nil, err
	}
	return res, nil
}
