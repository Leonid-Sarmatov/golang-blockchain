package blockchain

import (
	"fmt"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/boltdb"
	"golang_blockchain/pkg/hash_calulator"
	"golang_blockchain/pkg/iterator"
	//"golang_blockchain/pkg/proof_of_work"
	"log"
)

/* Контроллер блокчейна */
type BlockchainController struct {
	blockchain *blockchain.Blockchain
}

/* Конструктор */
func NewBlockchainController() (*BlockchainController, error) {
	// Хранилище блокчейна (база данных)
	storage := boltdb.NewBBoltDBDriver()

	// хэш-калькулятор
	hc := hashcalulator.NewHashCalculator()

	// Инициализация блокчейна
	b, err := blockchain.NewBlockchain(storage, hc)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}

	log.Printf("Контроллер блокчейна успешно запущен!")

	return &BlockchainController{
		blockchain: b,
	}, nil
}

/*
AddBlock добавляет блок в блокчейн и записывает его на диск

Аргументы:
  - []byte: data блок (в байтовом представлении)
  - int: pwValue 

Возвращает:
  - error: ошибка
*/
func (controller *BlockchainController) AddBlock(b *block.Block, pwValue int) error {
	return controller.blockchain.AddBlockToBlockchain(b, pwValue, controller.blockchain.HachCalc)
}

/*
CreateBlock создает блок, и записывает 
в него полезную нагрузку (транзакцию в байтовом представлении)

Аргументы:
  - []byte: data данные блока (полезная нагрузка в виде транзакции)
  - int: pwValue доказательство работы

Возвращает:
  - error: ошибка
*/
func (controller *BlockchainController) CreateBlock(data []byte) (*block.Block, error) {
	return controller.blockchain.CreateNewBlock(data)
}

/*
CreateIterator возвращает абстрактный итератор по
блокам в блокчейне, первый блок - самый новый в блокчейне

Возвращает:
  - iterator.Iterator[*block.Block]: экземпляр итератора
  - error: ошибка
*/
func (controller *BlockchainController) CreateIterator() (iterator.Iterator[*block.Block], error) {
	iter, err := controller.blockchain.CreateIterator()
	return iter, err
}

/*
GetBlockByHash возвращает блок с заданным хэшом

Возвращает:
  - *block.Block: указатель на блок
  - error: ошибка
*/
func (controller *BlockchainController) GetBlockByHash(hash []byte) (*block.Block, error) {
	return nil, nil
}

/*
GetAllBlocks возвращает все блоки в блокчейне

Возвращает:
  - []*block.Block: слайс указателей на блоки
  - error: ошибка
*/
func (controller *BlockchainController) GetAllBlocks() ([]*block.Block, error) {
	return nil, nil
}
