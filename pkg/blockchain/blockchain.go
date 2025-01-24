package blockchain

import (
	"fmt"
	block "golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
	proofofwork "golang_blockchain/pkg/proof_of_work"
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
	HashCalc block.HashCalculator
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
	hc block.HashCalculator, pow block.ProofOfWork) (*Blockchain, error) {
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
		hashCalculator := proofofwork.NewProofOfWork()
		g := block.NewGenesisBlock(hashCalculator)
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
func (bc *Blockchain) AddBlockToBlockchain(data []byte) error {
	// Получаем кончик блокчейна
	tip, err := bc.Storage.BlockchainGetTip()
	if err != nil {
		return fmt.Errorf("Can not get blockchain tip: %v", err)
	}

	newBlock, err := block.NewBlock(data, tip, bc.POW)
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
Реализует интерфейс IterableCollection, для
того что бы можно было итерироваться по блокчейну
*/
func (bc *Blockchain) CreateIterator() (iterator.Iterator, error) {
	var iterator blockchainIterator

	tip, err := bc.Storage.BlockchainGetTip()
	if err != nil {
		return nil, fmt.Errorf("Can not create iterator: %v", err)
	}

	iterator.currentHash = tip
	iterator.blockchain = bc

	return &iterator, nil
}

/* Структура итератора по блокчейну */
type blockchainIterator struct {
	blockchain  *Blockchain
	currentHash []byte
}

func (i *blockchainIterator) Next() (interface{}, error) {
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

func (i *blockchainIterator) HasNext() (bool, error) {
	current, err := i.Current()
	if err != nil {
		return false, err
	}

	if current.(*block.Block).ProofOfWorkValue == 0 || len(current.(*block.Block).PrevBlockHash) == 0 {
		return false, nil
	}

	return true, nil
}

func (i *blockchainIterator) Current() (interface{}, error) {
	block, err := i.blockchain.Storage.GetExistBlockByHash(i.currentHash)
	if err != nil {
		return nil, fmt.Errorf("Iterator can not load current element: %v", err)
	}

	return block, nil
}
