package block

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type HashCalculator interface {
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

// Сериализация блока в байтовый слайс
func (block *Block)SerializeBlock() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Кодируем TimeOfCreation (int64)
	if err := binary.Write(buf, binary.LittleEndian, block.TimeOfCreation); err != nil {
		return nil, err
	}

	// Кодируем длину Data и сами данные
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(block.Data))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(block.Data); err != nil {
		return nil, err
	}

	// Кодируем длину PrevBlockHash и сами данные
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(block.PrevBlockHash))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(block.PrevBlockHash); err != nil {
		return nil, err
	}

	// Кодируем длину Hash и сами данные
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(block.Hash))); err != nil {
		return nil, err
	}
	if _, err := buf.Write(block.Hash); err != nil {
		return nil, err
	}

	// Кодируем ProofOfWorkValue (int)
	if err := binary.Write(buf, binary.LittleEndian, int32(block.ProofOfWorkValue)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Десериализация байтового слайса в блок
func DeserializeBlock(data []byte) (*Block, error) {
	buf := bytes.NewReader(data)
	block := &Block{}

	// Декодируем TimeOfCreation (int64)
	if err := binary.Read(buf, binary.LittleEndian, &block.TimeOfCreation); err != nil {
		return nil, err
	}

	// Декодируем длину Data и сами данные
	var dataLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &dataLen); err != nil {
		return nil, err
	}
	block.Data = make([]byte, dataLen)
	if _, err := buf.Read(block.Data); err != nil {
		return nil, err
	}

	// Декодируем длину PrevBlockHash и сами данные
	var prevHashLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &prevHashLen); err != nil {
		return nil, err
	}
	block.PrevBlockHash = make([]byte, prevHashLen)
	if _, err := buf.Read(block.PrevBlockHash); err != nil {
		return nil, err
	}

	// Декодируем длину Hash и сами данные
	var hashLen uint32
	if err := binary.Read(buf, binary.LittleEndian, &hashLen); err != nil {
		return nil, err
	}
	block.Hash = make([]byte, hashLen)
	if _, err := buf.Read(block.Hash); err != nil {
		return nil, err
	}

	// Декодируем ProofOfWorkValue (int)
	var powValue int32
	if err := binary.Read(buf, binary.LittleEndian, &powValue); err != nil {
		return nil, err
	}
	block.ProofOfWorkValue = int(powValue)

	return block, nil
}

// /*
// ProofofWork описывает интерфейс для структур,
// способных подтвердить работу
// */
// type ProofOfWork interface {
// 	PWExecute(block *Block, pwValue int) (int, []byte, error)
// }

/*
NewBlock создает новый блок

	data - данные для нового блока
	prewBlochHash - хеш предыдущего блока
	pw - объект интерфеса для подтверждения работы
*/
func NewBlock(data []byte, prevBlockHash []byte) (*Block, error) {
	block := &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             data,
		PrevBlockHash:    prevBlockHash,
		Hash:             []byte{},
		ProofOfWorkValue: 0,
	}

	return block, nil
}

/*
NewBlock создает новый блок в блокчейн
*/
func NewGenesisBlock(hc HashCalculator) *Block {
	hash := hc.HashCalculate([]byte("Genesis block!"))
	log.Printf("Genesis block was successful, hash:\n%x\n", hash)
	return &Block{
		TimeOfCreation:   time.Now().Unix(),
		Data:             []byte("Genesis block!"),//Genesis block!
		PrevBlockHash:    []byte{},
		Hash:             hash,
		ProofOfWorkValue: -1,//0
	}
}

/*
SetHash устанавливает хеш блока
*/
func (b *Block) SetPOWAndHash(pow int, hc HashCalculator) error {
	b.ProofOfWorkValue = pow
	data, err := b.SerializeBlock()
	if err != nil {
		return fmt.Errorf("SetHash failed: %v", err)
	}

	b.Hash = hc.HashCalculate(data)
	return nil
}