package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

/* Единица блокчейна */
type Block struct {
	TimeOfCreation   int64
	Data             []byte
	PrevBlockHash    []byte
	Hash             []byte
	ProofOfWorkValue int
}

/*
BlockToBytesBlock преобразует экземпляр
структуры блока в байтовый слайс
*/
func (b *Block) BlockToBytes() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		return nil, fmt.Errorf("Convert block to byte slice was failed: %v\n", err)
	}

	return result.Bytes(), nil
	// // Используем библиотеку для работы с бинарными данными
	// binary.Write(&result, binary.LittleEndian, b.TimeOfCreation)
	// result.Write(b.Data)
	// result.Write(b.PrevBlockHash)
	// binary.Write(&result, binary.LittleEndian, int64(b.ProofOfWorkValue))

	// return result.Bytes()
}

/*
BytesToBlock парсит бинарное представление
блока в структуру

	clice - бинарные данные
*/
func (b *Block) BytesToBlock(clice []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(clice))
	return decoder.Decode(b)
}

/*
ProofofWork описывает интерфейс для структур,
способных подтвердить работу
*/
type ProofOfWork interface {
	PWExecute(block *Block) (int, []byte, error)
}

/*
NewBlock создает новый блок в блокчейн

	data - данные для нового блока
	prewBlochHash - хеш предыдущего блока
	pw - объект интерфеса для подтверждения работы
*/
func NewBlock(data string, prewBlochHash []byte, pw ProofOfWork) (*Block, error) {
	// Подготавливаем блок
	block := &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             []byte(data),
		PrevBlockHash:    prewBlochHash,
		Hash:             []byte{},
		ProofOfWorkValue: 0,
	}

	// Проверяем работу
	val, hash, err := pw.PWExecute(block)
	if err != nil {
		return block, fmt.Errorf("Invalid proof-of-work, blok was not create: %v\n", err)
	}
	block.ProofOfWorkValue = val
	block.Hash = hash

	return block, nil
}

/*
HashCalculate описывает интерфейс для различных
вариантов хеш-калькуляторов
*/
type HashCalculator interface {
	HashCalculate(data []byte) []byte
}

/*
NewBlock создает новый блок в блокчейн
*/
func NewGenesisBlock(hc HashCalculator) *Block {
	hash := hc.HashCalculate([]byte("Genesis block!"))
	log.Printf("Genesis block was successful, hash:\n%x\n", hash)
	return &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             []byte("Genesis block!"),
		PrevBlockHash:    []byte{},
		Hash:             hash,
		ProofOfWorkValue: 0,
	}
}
