package blockchain

import (
	"fmt"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/hash_calulator"
	"golang_blockchain/pkg/iterator"
	"log"
)

/*
Предпочитаем композицию вместо наследования.
Пусть блокчейн содержит в себе необходимые
компоненты в виде типов-интерфейсов, чтобы
не быть привязанным к конкретным структурам-классам
*/

/*
Интерфейс для механизма сохранения блокчейна
*/
type BlockchainStorage interface {
	/* Функция проверяющая наличие блокчейна в хранилище */
	IsBlockchainExist() (bool, error)
	/* Функция сохраняющая готовый генезис блок в хранилище */
	MakeNewBlockchain(genesisBlock *block.Block) error
	/* Функция загружающая из хранилища хэш последнего блока */
	BlockchainGetTip() ([]byte, error)
	/* Функция сохраняющая новый блок внутри хранилища */
	WriteNewBlock(newBlock *block.Block, lastHash []byte) error
	/* Функция загружающая из хранилища существующий блок по хэшу */
	GetExistBlockByHash(lastHash []byte) (*block.Block, error)
}

/* Сама структура блокчейна */
type Blockchain struct {
	Storage  BlockchainStorage
	HashCalc hashcalulator.HashCalculator
	POW      block.ProofOfWork
	TipHash  []byte
}

/*
Конструктор. Принимает в себя необходимые компоненты в виде интерфейса

	storage BlockchainStorage - абстрктное хранилище
	hc block.HashCalculator - абстрактный хеш-генератор
	pow block.ProofOfWork - абстрактный подтвердитель работы
*/
func NewBlockchain(storage BlockchainStorage,
	hc hashcalulator.HashCalculator, pow block.ProofOfWork) (*Blockchain, error) {
	// Подготавливаем структуру
	blockchain := &Blockchain{
		Storage:  storage,
		HashCalc: hc,
		POW:      pow,
	}

	log.Println("Приступаю к инициализации...")

	// Проверяем создан ли блокчейн на диске
	ok, err := storage.IsBlockchainExist()
	if err != nil {
		return nil, fmt.Errorf("Check exist blockchain was failed: %v", err)
	}

	if !ok {
		fmt.Println("Блокчейн не создан! Приступаю к созданию...")
		// Если нет, то создаем генезис блок
		g := block.NewGenesisBlock(hc)
		err = storage.MakeNewBlockchain(g)
		if err != nil {
			return nil, fmt.Errorf("Create genesis block was failed: %v", err)
		}
	}
	log.Println("Загружаю кончик...")

	// Загружаем кончик генезис блока
	tip, err := storage.BlockchainGetTip()
	if err != nil {
		return nil, fmt.Errorf("Can not init blockchain tip: %v", err)
	}
	blockchain.TipHash = tip

	log.Printf("Значение кончика получено: %v\n", tip)

	return blockchain, nil
}

/*
AddBlockToBlockchain добавляет новый блок в блокчейн

	data полезная нагрузка блока в виде строки
*/
func (bc *Blockchain) AddBlockToBlockchain(data []byte, pwValue int) error {
	// Получаем кончик блокчейна
	tip, err := bc.Storage.BlockchainGetTip()
	if err != nil {
		return fmt.Errorf("Can not get blockchain tip: %v", err)
	}

	newBlock, err := block.NewBlock(data, tip, bc.POW, pwValue)
	if err != nil {
		return fmt.Errorf("Creating new block to blockchain was failed: %v", err)
	}

	err = bc.Storage.WriteNewBlock(newBlock, tip)
	if err != nil {
		return fmt.Errorf("Saving new block to blockchain was failed: %v", err)
	}

	log.Printf("Новый блок в блокчейн успешно создан! Хеш последнего блока: %x\n", newBlock.Hash)

	return nil
}

/*
=======================================================
================ Итератор по блокчейну ================
=======================================================
*/
func (bc *Blockchain) CreateIterator() (iterator.Iterator[*block.Block], error) {
	var iterator blockchainIterator[*block.Block]

	tip, err := bc.Storage.BlockchainGetTip()
	if err != nil {
		return nil, fmt.Errorf("Can not create iterator: %v", err)
	}

	iterator.currentHash = tip
	iterator.blockchain = bc

	return &iterator, nil
}

/* Структура итератора по блокчейну */
type blockchainIterator[T any] struct {
	blockchain  *Blockchain
	currentHash []byte
}

func (i *blockchainIterator[T]) Next() (*block.Block, error) {
	block, err := i.blockchain.Storage.GetExistBlockByHash(i.currentHash)
	if err != nil {
		return nil, fmt.Errorf("Iterator can not load next element: %v", err)
	}

	block, err = i.blockchain.Storage.GetExistBlockByHash(block.PrevBlockHash)
	if err != nil {
		return nil, fmt.Errorf("Iterator can not load next element: %v", err)
	}

	i.currentHash = block.Hash

	return block, nil
}

func (i *blockchainIterator[T]) HasNext() (bool, error) {
	current, err := i.Current()
	if err != nil {
		return false, err
	}

	if current.ProofOfWorkValue == 0 || len(current.PrevBlockHash) == 0 {
		return false, nil
	}

	return true, nil
}

func (i *blockchainIterator[T]) Current() (*block.Block, error) {
	block, err := i.blockchain.Storage.GetExistBlockByHash(i.currentHash)
	if err != nil {
		return nil, fmt.Errorf("Iterator can not load current element: %v", err)
	}

	return block, nil
}

/*
=======================================================
============= Конец итератора по блокчейну ============
=======================================================
*/