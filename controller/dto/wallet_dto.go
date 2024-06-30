package dto

import (
	"encoding/hex"
	"errors"
	"strings"
)

type GenerateWalletData struct {
	SeedPhrase string `json:"seed_phrase"`
}

type ImportAccountData struct {
	PrivateKey string `json:"private_key" required:"true"`
}

func (d *ImportAccountData) Validate() error {
	privateKey := d.PrivateKey

	// Remove '0x' prefix if present
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}

	// Check if the private key is a valid hex string
	if _, err := hex.DecodeString(privateKey); err != nil {
		return errors.New("invalid private key: not a valid hex string")
	}

	// Check if the private key length is 64 characters (32 bytes)
	if len(privateKey) != 64 {
		return errors.New("invalid private key: must be 64 characters long")
	}

	// Additional checks could be added here, such as validating that the key is within the Ethereum private key range

	return nil
}
