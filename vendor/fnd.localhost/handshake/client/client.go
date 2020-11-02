package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fnd.localhost/handshake/primitives"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"
)

type Client struct {
	apiKey   string
	host     string
	port     int
	c        *http.Client
	rpcID    int64
	basePath string
}

type Opt func(c *Client)

type rpcRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int64         `json:"id"`
}

type rpcResponse struct {
	ID     int64 `json:"id"`
	Result json.RawMessage
	Error  *RPCError
}

type RPCError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (r *RPCError) Error() string {
	return r.Message
}

func WithAPIKey(apiKey string) Opt {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

func WithHTTPClient(client *http.Client) Opt {
	return func(c *Client) {
		c.c = client
	}
}

func WithNetwork(n primitives.Network) Opt {
	return func(c *Client) {
		c.port = n.RPCPort()
	}
}

func WithPort(port int) Opt {
	return func(c *Client) {
		c.port = port
	}
}

func WithBasePath(path string) Opt {
	return func(c *Client) {
		c.basePath = strings.Trim(path, "/")
	}
}

func NewClient(host string, opts ...Opt) *Client {
	c := &Client{
		host: host,
		port: primitives.NetworkMainnet.RPCPort(),
		c:    http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) getJSON(path string, result interface{}) error {
	req, err := http.NewRequest("GET", c.makeURL(path), nil)
	if err != nil {
		return err
	}
	if c.apiKey != "" {
		req.SetBasicAuth("x", c.apiKey)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}
	return nil
}

func (c *Client) postJSON(path string, body interface{}, result interface{}) error {
	bodyB, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.makeURL(path), bytes.NewReader(bodyB))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.SetBasicAuth("x", c.apiKey)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(resBody, result); err != nil {
		return err
	}
	return nil
}

func (c *Client) executeRPC(method string, resp interface{}, params ...interface{}) error {
	reqBody, err := json.Marshal(&rpcRequest{
		ID:     atomic.AddInt64(&c.rpcID, 1),
		Method: method,
		Params: params,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.makeURL(""), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.SetBasicAuth("x", c.apiKey)
	}
	res, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("non-200 status code: %d", res.StatusCode)
	}
	if resp == nil {
		return nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	resBody := new(rpcResponse)
	if err := json.Unmarshal(body, resBody); err != nil {
		return err
	}
	if resBody.Error != nil {
		return resBody.Error
	}
	if err := json.Unmarshal(resBody.Result, resp); err != nil {
		return err
	}
	return nil
}

func (c *Client) makeURL(path string) string {
	if c.basePath != "" {
		return fmt.Sprintf("%s:%d/%s/%s", c.host, c.port, c.basePath, path)
	}

	return fmt.Sprintf("%s:%d/%s", c.host, c.port, path)
}
