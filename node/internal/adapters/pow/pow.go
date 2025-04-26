package pow

import (
	"crypto/sha256"
	"log"
	"node/internal/block"

	//"log"
	"fmt"
	"math/big"
)

const (
	maxNonce   = 256 * 65535
	targetBits = 20
)

type hashCalculator struct{}

func NewHashCalculator() *hashCalculator {
	return &hashCalculator{}
}

func (hc *hashCalculator) HashCalculate(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

type proofOfWorkCheker struct {
	target *big.Int
	hashCalculator
}

func NewProofOfWorkCheker() *proofOfWorkCheker {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	return &proofOfWorkCheker{
		target:         target,
		hashCalculator: *NewHashCalculator(),
	}
}

func (pow *proofOfWorkCheker) Check(block *block.Block) (bool, error) {
	//log.Printf("Проверка доказательства работы...")
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int

	data, err := block.SerializeBlockWithoutHash()
	if err != nil {
		return false, err
	}
	// Вычисляем хэш блока
	hashInt.SetBytes(pow.HashCalculate(data))

	log.Printf("<pow.go> Хэш в блоке: %x", block.Hash)

	// Проверяем, удовлетворяет ли хэш целевому значению
	if hashInt.Cmp(target) == -1 {
		log.Printf("<pow.go> Доказательство работы подтверждено. Значение: %v", block.ProofOfWorkValue)
		return true, nil // Хэш подходит
	} else {
		log.Printf("<pow.go> Доказательство работы не подтверждено. Значение: %v", block.ProofOfWorkValue)
		return false, nil // Хэш не подходит
	}
}

type proofOfWorkSolver struct {
	hashCalculator
}

func NewProofOfWorkSolver() *proofOfWorkSolver {
	return &proofOfWorkSolver{
		hashCalculator: *NewHashCalculator(),
	}
}

func (solver *proofOfWorkSolver) Exec(blk *block.Block, cancel <-chan interface{}) error {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	counter := 0
	var hash []byte

	for {
		// Перебираем nonce до тех пор, пока не найдем подходящий хэш
		blk.ProofOfWorkValue = counter // Устанавливаем текущее значение nonce
		bytes, err := blk.SerializeBlockWithoutHash()
		if err != nil {
			return fmt.Errorf("Can not calculate proof-of-work from block: %v", err)
		}

		// Вычисляем хэш блока
		hash = solver.HashCalculate(bytes)
		hashInt.SetBytes(hash)

		// Проверяем, удовлетворяет ли хэш целевому значению
		if hashInt.Cmp(target) == -1 {
			blk.Hash = hash
			blk.ProofOfWorkValue = counter
			log.Printf("<pow.go> HASH: %x,  POW: %v", blk.Hash, blk.ProofOfWorkValue)
			return nil // Хэш подходит, завершаем цикл
		} else {
			counter++ // Увеличиваем nonce и продолжаем поиск
		}

		// Отмена подсвета POW
		select {
		case <-cancel:
			log.Printf("<pow.go> Отмена подсчета proof-of-work!")
			return fmt.Errorf("Calculate POW was canceled")
		default:
			continue
		}
	}
}
