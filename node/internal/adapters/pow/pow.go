package pow

import (
	"crypto/sha256"
	"node/internal/block"
	"log"
	"fmt"
	"math/big"
)

const (
	maxNonce   = 256 * 65535
	targetBits = 20
)



type hashCalculator struct {}

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
		target: target,
		hashCalculator: *NewHashCalculator(),
	}
}

func (pow *proofOfWorkCheker) Check(block *block.Block) (bool, error) {
	log.Printf("Проверка доказательства работы...")
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	var hash [32]byte

	data, err := block.SerializeBlock()
	if err != nil {
		return false, err
	}
	// Вычисляем хэш блока
	hash = [32]byte(pow.HashCalculate(data))
	hashInt.SetBytes(hash[:])

	// fmt.Printf("Хэш блока после установки POW: %x\n", hash)
	// fmt.Printf("Хэш установленый в блоке: %x\n", block.Hash)
	// fmt.Printf("Хэш предыдущего блока установленый в блоке: %x\n", block.PrevBlockHash)
	// fmt.Printf("Доказательство работы установленое в блоке: %v\n", block.ProofOfWorkValue)
	// fmt.Printf("Время создания блока установленое в блоке: %v\n", block.TimeOfCreation)
	// fmt.Println()

	// Проверяем, удовлетворяет ли хэш целевому значению
	if hashInt.Cmp(target) == -1 {
		log.Printf("Доказательство работы подтверждено. Значение: %v", block.ProofOfWorkValue)
		return true, nil // Хэш подходит
	} else {
		log.Printf("Доказательство работы не подтверждено. Значение: %v", block.ProofOfWorkValue)
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

func (solver *proofOfWorkSolver)Exec(blk *block.Block, cancel <-chan int) (int, error) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	var hash [32]byte
	counter := 0

	//log.Printf("Mining the block containing: %x\n", data)

	// block, err := block.DeserializeBlock(data)
	// if err != nil {
	// 	return -1, fmt.Errorf("Не смог десериализовать: %v", err)
	// }

	//fmt.Printf("Хэш блока до подсчета POW: %x\n", m.hachCalc.HashCalculate(data))
	//fmt.Printf("Хэш блока до подсчета POW: %x\n", m.hachCalc.HashCalculate(s))

	for {
		// Перебираем nonce до тех пор, пока не найдем подходящий хэш
		blk.ProofOfWorkValue = counter // Устанавливаем текущее значение nonce
		bytes, err := blk.SerializeBlock()
		if err != nil {
			return -1, fmt.Errorf("Can not calculate proof-of-work from block: %v", err)
		}

		// Вычисляем хэш блока
		hash = [32]byte(solver.HashCalculate(bytes))
		hashInt.SetBytes(hash[:])

		// Проверяем, удовлетворяет ли хэш целевому значению
		if hashInt.Cmp(target) == -1 {
			break // Хэш подходит, завершаем цикл
		} else {
			counter++ // Увеличиваем nonce и продолжаем поиск
		}

		// Отмена подсвета POW
		select {
		case <- cancel:
			return -1, fmt.Errorf("Cancel calculate proof-of-work")
		}
	}

	// log.Printf("Counter result value: %v\n", counter)
	// fmt.Printf("Хэш блока: %x\n", hash)
	// fmt.Printf("Доказательство работы: %v\n", counter)
	// fmt.Printf("Хэш установленый в блоке: %x\n", block.Hash)
	// fmt.Printf("Хэш предыдущего блока установленый в блоке: %x\n", block.PrevBlockHash)
	// fmt.Printf("Доказательство работы установленое в блоке: %v\n", block.ProofOfWorkValue)
	// fmt.Printf("Время создания блока установленое в блоке: %v\n", block.TimeOfCreation)
	return counter, nil
}