package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func randomString(n int) (string, error) {
	const letters = "0123456789"
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %v", err)
		}
		res[i] = letters[num.Int64()]
	}

	return string(res), nil
}
