package util

import (
	"crypto/ecdsa"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

func GenerateKeyPair() (KeyPair, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return KeyPair{}, err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return KeyPair{
		PrivateKey: hexutil.Encode(privateKeyBytes)[2:],
		PublicKey:  hexutil.Encode(publicKeyBytes)[4:],
		Address:    address,
	}, nil
}
