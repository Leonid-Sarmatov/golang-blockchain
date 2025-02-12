package proofofwork

import (
	//"crypto/sha256"
	"fmt"
	"golang_blockchain/pkg/block"
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
type HashCalulator interface {
	HashCalculate(data []byte) []byte
}

func (pow *ProofOfWorkCheker) Check(block *block.Block, value int, hc HashCalulator) (bool, error)  {
	fmt.Printf("Доказательство работы: %v\n", value)

	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	var hash [32]byte

	// data2, err := block.BlockToBytes()
	// if err != nil {
	// 	return false, err
	// }
	// // Вычисляем хэш блока
	// hash = [32]byte(hc.HashCalculate(data2))
	// hashInt.SetBytes(hash[:])

	// fmt.Printf("Хэш блока до установки POW: %x\n", hash)
	// fmt.Printf("Хэш установленый в блоке: %x\n", block.Hash)
	// fmt.Printf("Хэш предыдущего блока установленый в блоке: %x\n", block.PrevBlockHash)
	// fmt.Printf("Доказательство работы установленое в блоке: %v\n", block.ProofOfWorkValue)
	// fmt.Printf("Время создания блока установленое в блоке: %v\n", block.TimeOfCreation)

	block.ProofOfWorkValue = value

	data, err := block.SerializeBlock()
	if err != nil {
		return false, err
	}
	// Вычисляем хэш блока
	hash = [32]byte(hc.HashCalculate(data))
	hashInt.SetBytes(hash[:])

	fmt.Printf("Хэш блока после установки POW: %x\n", hash)
	fmt.Printf("Хэш установленый в блоке: %x\n", block.Hash)
	fmt.Printf("Хэш предыдущего блока установленый в блоке: %x\n", block.PrevBlockHash)
	fmt.Printf("Доказательство работы установленое в блоке: %v\n", block.ProofOfWorkValue)
	fmt.Printf("Время создания блока установленое в блоке: %v\n", block.TimeOfCreation)

	fmt.Println()

	// Проверяем, удовлетворяет ли хэш целевому значению
	if hashInt.Cmp(target) == -1 {
		return true, nil // Хэш подходит
	} else {
		return false, nil // Хэш не подходит
	}
}

func Pipapupa(data []byte, hachCalc HashCalulator) (int, error) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	var hash [32]byte
	counter := 0

	log.Printf("Mining the block containing: %x\n", data)

	block, err := block.DeserializeBlock(data)
	if err != nil {
		return -1, err
	}

	fmt.Printf("Хэш блока до подсчета POW: %x\n", hachCalc.HashCalculate(data))

	for {
		// Перебираем nonce до тех пор, пока не найдем подходящий хэш
		block.ProofOfWorkValue = counter // Устанавливаем текущее значение nonce
		bytes, err := block.SerializeBlock()
		if err != nil {
			return -1, fmt.Errorf("Can not calculate hash from block: %v\n", err)
		}

		// Вычисляем хэш блока
		hash = [32]byte(hachCalc.HashCalculate(bytes))
		hashInt.SetBytes(hash[:])

		// Проверяем, удовлетворяет ли хэш целевому значению
		if hashInt.Cmp(target) == -1 {
			break // Хэш подходит, завершаем цикл
		} else {
			counter++ // Увеличиваем nonce и продолжаем поиск
		}
	}

	//log.Printf("Counter result value: %v\n", counter)
	fmt.Printf("Хэш блока: %x\n", hash)
	fmt.Printf("Доказательство работы: %v\n", counter)
	fmt.Printf("Хэш установленый в блоке: %x\n", block.Hash)
	fmt.Printf("Хэш предыдущего блока установленый в блоке: %x\n", block.PrevBlockHash)
	fmt.Printf("Доказательство работы установленое в блоке: %v\n", block.ProofOfWorkValue)
	fmt.Printf("Время создания блока установленое в блоке: %v\n", block.TimeOfCreation)
	return counter, nil
}
