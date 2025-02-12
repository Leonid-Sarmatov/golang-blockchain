package proofofwork

import (
	"crypto/sha256"
	"fmt"
	"miner/pkg/block"
	"log"
	"math/big"
)

const (
	maxNonce   = 256 * 65535
	targetBits = 15
)

type ProofOfWorkCheker struct {
	target *big.Int
}

func NewProofOfWorkCheker() *ProofOfWorkCheker {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWorkCheker{target: target}
}

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

func (pow *ProofOfWorkCheker) Check(data []byte, value int, hc hashCalulator) (bool, error)  {
	var hashInt big.Int
	// Вычисляем хэш блока
	hash := hc.HashCalculate(data)
	hashInt.SetBytes(hash)

	// Проверяем, удовлетворяет ли хэш целевому значению
	if hashInt.Cmp(pow.target) == -1 {
		return true, nil // Хэш подходит
	} else {
		return false, nil // Хэш не подходит
	}
}

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
		bytes, err := block.SerializeBlock()
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
