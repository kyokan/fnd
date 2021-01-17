package cli

import (
	"fnd/config"
	"fnd/crypto"
)

func GetSigner(homeDir string) (crypto.Signer, error) {
	identity, err := config.ReadNodeIdentity(homeDir)
	if err != nil {
		return nil, err
	}
	return crypto.NewSECP256k1Signer(identity.PrivateKey), nil
}
