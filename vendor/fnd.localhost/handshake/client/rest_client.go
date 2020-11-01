package client

import (
	"errors"
	"fmt"
)

func (c *Client) GetInfo() (*NodeInfo, error) {
	info := new(NodeInfo)
	if err := c.getJSON("", info); err != nil {
		return nil, err
	}
	return info, nil
}

func (c *Client) GetMempoolSnapshot() ([]string, error) {
	var out []string
	if err := c.getJSON("mempool", out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetMempoolRejectsFilter() (*MempoolRejectsFilterInfo, error) {
	info := new(MempoolRejectsFilterInfo)
	if err := c.getJSON("mempool/invalid", info); err != nil {
		return nil, err
	}
	return info, nil
}

func (c *Client) TestMempoolRejectsFilter(hash string) (bool, error) {
	info := new(struct {
		Invalid bool `json:"invalid"`
	})
	if err := c.getJSON(fmt.Sprintf("mempool/invalid/%s", hash), info); err != nil {
		return false, err
	}
	return info.Invalid, nil
}

func (c *Client) GetBlockByHash(hash string) (*RESTBlock, error) {
	block := new(RESTBlock)
	if err := c.getJSON(fmt.Sprintf("block/%s", hash), block); err != nil {
		return nil, err
	}
	return block, nil
}

func (c *Client) GetBlockByHeight(height int) (*RESTBlock, error) {
	if height < 0 {
		return nil, errors.New("cannot set a negative height")
	}
	block := new(RESTBlock)
	if err := c.getJSON(fmt.Sprintf("block/%d", height), block); err != nil {
		return nil, err
	}
	return block, nil
}

func (c *Client) BroadcastTransaction(tx string) error {
	body := struct {
		Tx string `json:"tx"`
	}{
		tx,
	}
	res := new(struct {
		Success bool `json:"success"`
	})
	if err := c.postJSON("broadcast", body, res); err != nil {
		return err
	}
	if !res.Success {
		return errors.New("error broadcasting transaction, check HSD logs")
	}
	return nil
}

func (c *Client) BroadcastClaim(claim string) error {
	body := struct {
		Claim string `json:"claim"`
	}{
		claim,
	}
	res := new(struct {
		Success bool `json:"success"`
	})
	if err := c.postJSON("claim", body, res); err != nil {
		return err
	}
	if !res.Success {
		return errors.New("error broadcasting claim, check HSD logs")
	}
	return nil
}

func (c *Client) EstimateFee(blocks int) (uint64, error) {
	if blocks < 0 {
		return 0, errors.New("blocks cannot be negative")
	}
	res := new(struct {
		Rate uint64 `json:"rate"`
	})
	if err := c.getJSON(fmt.Sprintf("fee?blocks=%d", blocks), res); err != nil {
		return 0, err
	}
	return res.Rate, nil
}

func (c *Client) ResetBlockchain(height int) error {
	if height < 0 {
		return errors.New("cannot set a zero height")
	}
	body := struct {
		Height int `json:"height"`
	}{
		height,
	}
	res := new(struct {
		Success bool `json:"success"`
	})
	if err := c.postJSON("reset", body, res); err != nil {
		return err
	}
	if !res.Success {
		return errors.New("error resetting chain, check HSD logs")
	}
	return nil
}

func (c *Client) GetCoinByOutpoint(hash string, index int) (*Coin, error) {
	res := new(Coin)
	if err := c.getJSON(fmt.Sprintf("coin/%s/%d", hash, index), res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetCoinsByAddress(address string) ([]*Coin, error) {
	var res []*Coin
	if err := c.getJSON(fmt.Sprintf("coins/address/%s", address), res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetTransactionByHash(hash string) (*Transaction, error) {
	res := new(Transaction)
	if err := c.getJSON(fmt.Sprintf("tx/%s", hash), res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) GetTransactionsByAddress(address string) ([]*Transaction, error) {
	var res []*Transaction
	if err := c.getJSON(fmt.Sprintf("tx/address/%s", address), res); err != nil {
		return nil, err
	}
	return res, nil
}
