package blockchain

import (
	"bytes"
	"encoding/binary"
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
func (b *Block) BlockToBytes() []byte {
	var result bytes.Buffer

	// Используем библиотеку для работы с бинарными данными
	binary.Write(&result, binary.LittleEndian, b.TimeOfCreation)
	result.Write(b.Data)
	result.Write(b.PrevBlockHash)
	binary.Write(&result, binary.LittleEndian, int64(b.ProofOfWorkValue))

	return result.Bytes()
}

/* Сама структура блокчейна */
type Blockchain struct {
	Blocks []*Block
}

/*
ProofofWork описывает интерфейс для структур,
способных подтвердить работу
*/
type ProofOfWork interface {
	PWExecute(bytesBlock []byte) (int, []byte)
}

/*
NewBlock создает новый блок в блокчейн
	data - данные для нового блока
	prewBlochHash - хеш предыдущего блока
	pw - объект интерфеса для подтверждения работы
*/
func NewBlock(data string, prewBlochHash []byte, pw ProofOfWork) *Block {
	// Подготавливаем блок
	block := &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             []byte(data),
		PrevBlockHash:    prewBlochHash,
		Hash:             []byte{},
		ProofOfWorkValue: 0,
	}

	// Проверяем работу
	val, hash := pw.PWExecute(block.BlockToBytes())
	block.ProofOfWorkValue = val
	block.Hash = hash

	return block
}

/*
ProofofWork описывает интерфейс для различных
вариантов хеш-калькуляторов
*/
type HashCalculator interface {
	HashCalculate(data []byte) []byte
}

/*
NewBlock создает новый блок в блокчейн
*/
func NewGenesisBlock(hc HashCalculator) *Block {
    return &Block{
		TimeOfCreation: time.Now().Unix(),
		Data: []byte("Genesis block!"),
		PrevBlockHash: []byte{},
		Hash: hc.HashCalculate([]byte("Genesis block!")),
		ProofOfWorkValue: 0,
	}
}
