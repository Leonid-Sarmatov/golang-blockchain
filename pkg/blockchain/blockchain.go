package blockchain

import (
	block "golang_blockchain/pkg/block"
)

/* Сама структура блокчейна */
type Blockchain struct {
	Storage BlockchainStorage
	Blocks []*block.Block
}

/* 
Интерфейс для механизма сохранения 
блокчейна на жесткий диск
*/
type BlockchainStorage interface {
	IsBlockchainExist() (bool, error)
	BlockchainInit() error
}

/* Конструктор */
func NewBlockchain(storage BlockchainStorage) (*Blockchain, error) {
	// Подготавливаем структуру
	blockchain := &Blockchain{
		Storage: storage,
		Blocks: make([]*block.Block, 0),
	}


	return blockchain, nil
}

