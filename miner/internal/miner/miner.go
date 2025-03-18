package miner

import (
	"fmt"
	"log"
	"math/big"
	"miner/pkg/block"
	hashcalulator "miner/pkg/hash_calculator"
)

const (
	maxNonce   = 256 * 65535
	targetBits = 20
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

type Miner struct {
	hachCalc hashCalulator
}

func NewMiner() *Miner {
	return &Miner{
		hachCalc: hashcalulator.NewHashCalculator(),
	}
}

func (m *Miner) Do(data []byte) (int, error) {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var hashInt big.Int
	var hash [32]byte
	counter := 0

	log.Printf("Mining the block containing: %x\n", data)

	block, err := block.DeserializeBlock(data)
	if err != nil {
		return -1, fmt.Errorf("Не смог десериализовать: %v", err)
	}

	s, _ := block.SerializeBlock()

	fmt.Printf("Хэш блока до подсчета POW: %x\n", m.hachCalc.HashCalculate(data))
	fmt.Printf("Хэш блока до подсчета POW: %x\n", m.hachCalc.HashCalculate(s))

	for {
		// Перебираем nonce до тех пор, пока не найдем подходящий хэш
		block.ProofOfWorkValue = counter // Устанавливаем текущее значение nonce
		bytes, err := block.SerializeBlock()
		if err != nil {
			return -1, fmt.Errorf("Can not calculate hash from block: %v\n", err)
		}

		// Вычисляем хэш блока
		hash = [32]byte(m.hachCalc.HashCalculate(bytes))
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
