package shortener

import (
	"crypto/sha256"
	"fmt"
	"github.com/itchyny/base58-go"
	"math/big"
	"os"
)

func sha2560f(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return string(encoded)

}

func GenerateShortLink(initialLink string, userId string) string {
	//Generate a unique hash for the link
	urlHashBytes := sha2560f(initialLink + userId)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	//Encode the hash to base58
	shortLink := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	return shortLink[:8]
}
