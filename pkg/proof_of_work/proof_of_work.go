package proofofwork

import (
	"crypto/sha256"
	block "golang_blockchain/pkg/block"
	"log"
	"math/big"
)

const (
	maxNonce   = 65535
	targetBits = 12
)

type SelfProofOfWork struct {
	target *big.Int
}

func NewProofOfWork() *SelfProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &SelfProofOfWork{target}
}

/*
Реализует интерфейс проверки работы, в данной
реализации выполняет всю работу сам
*/
func (spw *SelfProofOfWork) PWExecute(block *block.Block) (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	counter := 0

	log.Printf("Mining the block containing \n%s\n", block.Data)
	for counter < maxNonce {
		block.ProofOfWorkValue = counter
		hash = [32]byte(spw.HashCalculate(block.BlockToBytes()))
		log.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(spw.target) == -1 {
			break
		} else {
			counter += 1
		}
	}
	log.Printf("Counter result value:\n%v\n", counter)

	return counter, hash[:]
}

/*
Реализует интерфейс хэш-калькулятора
*/
func (spw *SelfProofOfWork) HashCalculate(data []byte) []byte {
	// В данной реализации пусть будет SHA256
	hash := sha256.Sum256(data)
	return hash[:]
}
