package cli

import (
	"github.com/ddrp-org/ddrp/config"
	"github.com/ddrp-org/ddrp/crypto"
)

func GetSigner(homeDir string) (crypto.Signer, error) {
	identity, err := config.ReadNodeIdentity(homeDir)
	if err != nil {
		return nil, err
	}
	return crypto.NewSECP256k1Signer(identity.PrivateKey), nil
}
