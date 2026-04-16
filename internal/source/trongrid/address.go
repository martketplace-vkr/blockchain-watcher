package trongrid

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

const (
	tronAddressPrefix = byte(0x41)
	addressByteLen    = 20
)

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func normalizeEventAddress(address string) (string, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "", nil
	}

	if !strings.HasPrefix(address, "0x") && !strings.HasPrefix(address, "0X") {
		return address, nil
	}

	raw, err := hex.DecodeString(address[2:])
	if err != nil {
		return "", fmt.Errorf("invalid tron event address %q: %w", address, err)
	}
	if len(raw) != addressByteLen {
		return "", fmt.Errorf("invalid tron event address %q: expected %d bytes, got %d", address, addressByteLen, len(raw))
	}

	payload := append([]byte{tronAddressPrefix}, raw...)
	checksumInput := append([]byte(nil), payload...)
	firstHash := sha256.Sum256(checksumInput)
	secondHash := sha256.Sum256(firstHash[:])
	payload = append(payload, secondHash[:4]...)

	return encodeBase58(payload), nil
}

func encodeBase58(input []byte) string {
	value := new(big.Int).SetBytes(input)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	result := make([]byte, 0, len(input))
	for value.Cmp(zero) > 0 {
		value.DivMod(value, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	for _, b := range input {
		if b != 0 {
			break
		}
		result = append(result, base58Alphabet[0])
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
