package util

import (
	"crypto/ecdsa"
	"crypto/sha512"
	"golang.org/x/crypto/pbkdf2"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

type KeyPair struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Address    string `json:"address"`
}

func GenerateKeyPair(seedPhrase string) (KeyPair, error) {

	var privateKey *ecdsa.PrivateKey
	if seedPhrase == "" {
		privateKey, _ = crypto.GenerateKey()
	} else {
		// Validate and convert the seed phrase to a seed
		seed := bip39.NewSeed(seedPhrase, "")

		// Generate the private key from the seed
		var err error
		privateKey, err = crypto.ToECDSA(pbkdf2.Key(seed, []byte("Ethereum seed"), 2048, 32, sha512.New))
		if err != nil {
			log.Fatalf("failed to generate private key: %v", err)
		}
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

func GetKeypairFromPrivateKey(privateKey string) (KeyPair, error) {
	privateKeyBytes, err := hexutil.Decode(privateKey)
	if err != nil {
		return KeyPair{}, err
	}

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return KeyPair{}, err
	}

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return KeyPair{
		PrivateKey: privateKey,
		PublicKey:  hexutil.Encode(publicKeyBytes)[4:],
		Address:    address,
	}, nil
}
