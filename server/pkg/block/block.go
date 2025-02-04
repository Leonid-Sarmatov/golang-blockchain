package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	//"golang_blockchain/pkg/hash_calulator"
	"log"
	"time"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

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
	PWExecute(block *Block, pwValue int) (int, []byte, error)
}

/*
NewBlock создает новый блок в блокчейн

	data - данные для нового блока
	prewBlochHash - хеш предыдущего блока
	pw - объект интерфеса для подтверждения работы
*/
func NewBlock(data []byte, prewBlochHash []byte, pw ProofOfWork, pwValue int) (*Block, error) {
	// Подготавливаем блок
	block := &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             data,
		PrevBlockHash:    prewBlochHash,
		Hash:             []byte{},
		ProofOfWorkValue: 0,
	}

	// Проверяем работу
	val, hash, err := pw.PWExecute(block, pwValue)
	if err != nil {
		return block, fmt.Errorf("Invalid proof-of-work, blok was not create: %v\n", err)
	}
	
	block.ProofOfWorkValue = val
	block.Hash = hash

	return block, nil
}

/*
NewBlock создает новый блок в блокчейн
*/
func NewGenesisBlock(hc hashCalulator) *Block {
	hash := hc.HashCalculate([]byte("Genesis block!"))
	log.Printf("Genesis block was successful, hash:\n%x\n", hash)
	return &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             []byte("Genesis block!"),//Genesis block!
		PrevBlockHash:    []byte{},
		Hash:             hash,
		ProofOfWorkValue: 0,//0
	}
}
