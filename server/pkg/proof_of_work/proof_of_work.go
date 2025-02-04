package proofofwork

import (
	"crypto/sha256"
	"fmt"
	"golang_blockchain/pkg/block"
	"log"
	"math/big"
)

const (
	maxNonce   = 256 * 65535
	targetBits = 15
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
func (spw *SelfProofOfWork) PWExecute(block *block.Block, pwValue int) (int, []byte, error) {
	var hashInt big.Int
	var hash [32]byte
	counter := 0

	log.Printf("Mining the block containing: %x\n", block.Data)
	for counter < maxNonce {
		block.ProofOfWorkValue = counter
		bytes, err := block.BlockToBytes()
		if err != nil {
			return 0, nil, fmt.Errorf("Can not calculate hash from block: %v\n", err)
		}
		hash = [32]byte(spw.HashCalculate(bytes))
		//log.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(spw.target) == -1 {
			break
		} else {
			counter += 1
		}
	}
	log.Printf("Counter result value: %v\n", counter)

	return counter, hash[:], nil
}

/*
Реализует интерфейс хэш-калькулятора
*/
func (spw *SelfProofOfWork) HashCalculate(data []byte) []byte {
	// В данной реализации пусть будет SHA256
	hash := sha256.Sum256(data)
	return hash[:]
}
