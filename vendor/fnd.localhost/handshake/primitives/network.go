package primitives

import "errors"

type Network string

const (
	NetworkMainnet Network = "main"
	NetworkTestnet Network = "testnet"
	NetworkRegtest Network = "regtest"
	NetworkSimnet  Network = "simnet"
)

func (n Network) String() string {
	return string(n)
}

func (n Network) RPCPort() int {
	switch n {
	case NetworkMainnet:
		return 12037
	case NetworkTestnet:
		return 13037
	case NetworkRegtest:
		return 14037
	case NetworkSimnet:
		return 15037
	default:
		panic("invalid network")
	}
}

func (n Network) AddressHRP() string {
	switch n {
	case NetworkMainnet:
		return "hs"
	case NetworkTestnet:
		return "ts"
	case NetworkRegtest:
		return "rs"
	case NetworkSimnet:
		return "ss"
	default:
		panic("invalid network")
	}
}

func NetworkFromString(n string) (Network, error) {
	switch Network(n) {
	case NetworkMainnet:
		return NetworkMainnet, nil
	case NetworkTestnet:
		return NetworkTestnet, nil
	case NetworkRegtest:
		return NetworkRegtest, nil
	case NetworkSimnet:
		return NetworkSimnet, nil
	default:
		return NetworkMainnet, errors.New("invalid network")
	}
}
