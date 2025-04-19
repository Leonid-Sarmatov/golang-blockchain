package blockchain

import (
	"bytes"
	"fmt"
	"log"
	"node/internal/block"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

/*
Интерфейс для итератора
*/
type Iterator[T any] interface {
	HasNext() (bool, error)
	Next() (T, error)
	Current() (T, error)
}

/*
Интерфейс для механизма сохранения блокчейна
*/
type blockchainStorage interface {
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

/*
Блокчейн
*/
type Blockchain struct {
	Storage  blockchainStorage
	TipHash  []byte
	HachCalc hashCalulator
}

/*
NewBlockchain конструктор блокчейна

Аргументы:
  - storage BlockchainStorage: абстрактное хранилище
  - hc hashCalulator: абстрактный хэш-калькулятор

Возвращает:
  - *Blockchain: указатель на блокчейн
  - error: ошибка
*/
func NewBlockchain(storage blockchainStorage, hc hashCalulator) *Blockchain {
	return &Blockchain{
		Storage:  storage,
		HachCalc: hc,
	}
}

/*
NewBlockchain загрузка блокчейна с диска,
вовзращает ошибку, в случае неудачи

Возвращает:
  - error: ошибка
*/
func (bc *Blockchain) LoadSavedBlockchain() error {
	// Проверяем создан ли блокчейн на диске
	ok, err := bc.Storage.IsBlockchainExist()
	if err != nil {
		return fmt.Errorf("Check exist blockchain was failed: %v", err)
	}

	if !ok {
		fmt.Println("<blockchain.go> Блокчейн не создан! Приступаю к созданию...")
		// Если нет, то создаем генезис блок
		g := block.NewGenesisBlock(bc.HachCalc)
		err = bc.Storage.MakeNewBlockchain(g)
		if err != nil {
			return fmt.Errorf("Create genesis block was failed: %v", err)
		}
	}

	log.Println("<blockchain.go> Загружаю кончик...")

	// Загружаем кончик генезис блока
	tip, err := bc.Storage.BlockchainGetTip()
	if err != nil {
		return fmt.Errorf("Can not init blockchain tip: %v", err)
	}
	bc.TipHash = tip

	log.Printf("<blockchain.go> Блокчейн успешно инициализирован! Значение кончика: %x", tip)

	return nil
}

/*func NewBlockchain(storage BlockchainStorage, hc hashCalulator) (*Blockchain, error) {
	// Подготавливаем структуру
	blockchain := &Blockchain{
		Storage:  storage,
		HachCalc: hc,
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
}*/

func (bc *Blockchain) AddBlockToBlockchain(b *block.Block) error {
	if bytes.Compare(b.PrevBlockHash, bc.TipHash) != 0 {
		log.Printf("<blockchain.go> Хэши не совпали, блок не сохранен! Значение кончика %x, а в блоке записан %x", bc.TipHash, b.PrevBlockHash)
		return fmt.Errorf("Saving new block to blockchain was failed: %v", "prev-block-hach not equal tip-hash")
	}

	bs, err := b.SerializeBlock()
	if err != nil {
		log.Printf("<blockchain.go> Блок не удалось сериализовать")
		return fmt.Errorf("Saving new block to blockchain was failed: %v", err)
	}

	b.Hash = bc.HachCalc.HashCalculate(bs)

	err = bc.Storage.WriteNewBlock(b, bc.TipHash)
	if err != nil {
		log.Printf("<blockchain.go> Не удалось сохранить блок: %v", err)
		return fmt.Errorf("Saving new block to blockchain was failed: %v", err)
	}

	log.Printf("<blockchain.go> Новый блок в блокчейн успешно добавлен! Хеш последнего блока: %x\n", b.Hash)

	return nil
}

func (bc *Blockchain) IsAlreadyExistBlock(b *block.Block) bool {
	return bytes.Compare(b.Hash, bc.TipHash) == 0
}

// func (bc *Blockchain) CreateNewBlock(data []byte) (*block.Block, error) {
// 	// Получаем кончик блокчейна
// 	tip, err := bc.Storage.BlockchainGetTip()
// 	if err != nil {
// 		return nil, fmt.Errorf("Can not get blockchain tip: %v", err)
// 	}

// 	newBlock, err := block.NewBlock(data, tip)
// 	if err != nil {
// 		return nil, fmt.Errorf("Creating new block to blockchain was failed: %v", err)
// 	}

// 	return newBlock, nil
// }

/*
=======================================================
================ Итератор по блокчейну ================
=======================================================
*/
func (bc *Blockchain) CreateIterator() (Iterator[*block.Block], error) {
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

	if current.ProofOfWorkValue == -1 || len(current.PrevBlockHash) == 0 {
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
