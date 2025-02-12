package hashcalulator

import (
	"crypto/sha256"
)

type HashCalculator struct {}

func NewHashCalculator() *HashCalculator {
	return &HashCalculator{}
}

func (hc *HashCalculator) HashCalculate(data []byte) []byte {
	// В данной реализации пусть будет SHA256
	hash := sha256.Sum256(data)
	return hash[:]
}