package shortener

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/itchyny/base58-go"
)

func sha256Of(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func base58Encoded(bytes []byte) (string, error) {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		return "", fmt.Errorf("base58 encoding failed: %w", err)
	}
	return string(encoded), nil
}

// GenerateShortLink returns an 8-character Base58-encoded short code
// derived from a SHA-256 hash of the combination of initialLink and userId.
func GenerateShortLink(initialLink string, userId string) (string, error) {
	urlHashBytes := sha256Of(initialLink + userId)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	shortLink, err := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	if err != nil {
		return "", err
	}
	return shortLink[:8], nil
}
