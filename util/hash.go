package util

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"strings"
)

type HashAlgorithm string

const (
	SHA256    HashAlgorithm = "SHA256"
	SHA512    HashAlgorithm = "SHA512"
	Keccak256 HashAlgorithm = "Keccak256"
)

func Hash(data []byte, algorithm HashAlgorithm) []byte {
	switch algorithm {
	case SHA256:
		hash := sha256.New()
		hash.Write(data)
		return hash.Sum(nil)
	case SHA512:
		hash := sha512.New()
		hash.Write(data)
		return hash.Sum(nil)
	case Keccak256:
		return CryptoHash(data).Bytes()
	default:
		return nil
	}
}

func CryptoHash(data []byte) common.Hash {
	return crypto.Keccak256Hash(data)
}

func Sign(data []byte, privateKey string) (string, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	signature, err := crypto.Sign(data, privateKeyECDSA)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}

func VerifySignature(publicKey string, data []byte, signature string) bool {
	publicKeyBytes, err := hex.DecodeString("04" + publicKey)
	if err != nil {
		return false
	}

	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return false
	}

	signatureBytes = signatureBytes[:64]

	return crypto.VerifySignature(publicKeyBytes, data, signatureBytes)
}

func HexToBin(hexString string) (string, error) {
	if strings.HasPrefix(hexString, "0x") {
		hexString = hexString[2:]
	}

	decoded, err := hex.DecodeString(hexString)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	// Convert decoded bytes to binary string
	var binaryStrings []string
	for _, b := range decoded {
		binaryStrings = append(binaryStrings, fmt.Sprintf("%08b", b))
	}

	return strings.Join(binaryStrings, ""), nil
}
