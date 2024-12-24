package proofofwork

import (
	"crypto/sha256"
)

type SelfProofOfWork struct{}

func (spw *SelfProofOfWork) PWExecute(bytesBlock []byte) {
	
}

func (spw *SelfProofOfWork) HashCalculate(data []byte) []byte {
	// В данной реализации пусть будет SHA256
	hash := sha256.Sum256(data)
	// Sum256 возвращает массив, преобразуем его в слайл для удобства
	return hash[:]
}
