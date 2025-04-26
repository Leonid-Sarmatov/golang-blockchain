package blockchain

import (
	"bytes"
	"context"
	"fmt"
	"log"

	//"node/internal/adapters/pow"
	//"node/internal/adapters/storage"
	"node/internal/block"
	//"node/internal/blockchain"
	"node/internal/transaction"
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
	/* Хранилище блокчейна на диске */
	storage blockchainStorage
	/* Хэш последнего блока (кончик блокчейна) */
	tip []byte
	/* Калькулятор хэшей */
	hc hashCalulator
}

/*
NewBlockchainAdapter конструктор блокчейна
*/
func NewBlockchain(s blockchainStorage, hc hashCalulator) *Blockchain {
	return &Blockchain{
		storage: s,
		hc:      hc,
	}
}

/*
Init инициализация адаптера

Возвращает:
  - error: ошибка
*/
func (blockchain *Blockchain) TryLoadSavedBlockchain() error {
	// Проверяем создан ли блокчейн на диске
	ok, err := blockchain.storage.IsBlockchainExist()
	if err != nil {
		return fmt.Errorf("Check exist blockchain was failed: %v", err)
	}

	if !ok {
		fmt.Println("<blockchain.go> Блокчейн не создан! Приступаю к созданию...")
		// Если нет, то создаем генезис блок
		g := block.NewGenesisBlock(blockchain.hc)
		err = blockchain.storage.MakeNewBlockchain(g)
		if err != nil {
			return fmt.Errorf("Create genesis block was failed: %v", err)
		}
	}

	log.Println("<blockchain.go> Загружаю кончик...")

	// Загружаем кончик генезис блока
	tip, err := blockchain.storage.BlockchainGetTip()
	if err != nil {
		return fmt.Errorf("Can not init blockchain tip: %v", err)
	}
	blockchain.tip = tip

	log.Printf("<blockchain.go> Блокчейн успешно инициализирован! Значение кончика: %x", tip)

	return nil
}

/*
AddBlockToBlockchain пробует добавить новый блок
в блокчейн узла, в случае неудачи возвращает ошибку

Аргументы:
  - b *block.Block: указатель на блок

Возвращает:
  - error: ошибка
*/
func (blockchain *Blockchain) AddBlockToBlockchain(b *block.Block) error {
	if bytes.Compare(b.PrevBlockHash, blockchain.tip) != 0 {
		log.Printf(
			"<blockchain.go> Хэши не совпали, блок не сохранен! Значение кончика %x, а в блоке записан %x",
			blockchain.tip, b.PrevBlockHash,
		)
		return fmt.Errorf("Saving new block to blockchain was failed: %v", "prev-block-hach not equal tip-hash")
	}

	bs, err := b.SerializeBlock()
	if err != nil {
		log.Printf("<blockchain.go> Блок не удалось сериализовать")
		return fmt.Errorf("Saving new block to blockchain was failed: %v", err)
	}

	b.Hash = blockchain.hc.HashCalculate(bs)

	err = blockchain.storage.WriteNewBlock(b, blockchain.tip)
	if err != nil {
		log.Printf("<blockchain.go> Не удалось сохранить блок: %v", err)
		return fmt.Errorf("Saving new block to blockchain was failed: %v", err)
	}
	blockchain.tip = b.Hash

	log.Printf("<blockchain.go> Новый блок в блокчейн успешно добавлен! Хеш последнего блока: %x\n", b.Hash)

	return nil
}

/*
IsAlreadyExistBlock сравнивает хеш блока с хешом кончика

Аргументы:
  - b *block.Block: указатель на блок

Возвращает:
  - bool: true - совпадает, false - не совпадает
*/
func (blockchain *Blockchain) IsAlreadyExistBlock(b *block.Block) bool {
	//log.Printf("Хэш кончика = %x,   Хэш текущего блока = %x,   Хэш предыдущего блока = %x", blockchain.tip, b.Hash, b.PrevBlockHash)
	return bytes.Compare(b.Hash, blockchain.tip) == 0
}

/*
TryLoadSavedBlockchain попытка загрузить блокчейн
из внутреннего хранилища данного узла

Возвращает:
  - bool: успех/неуспех
  - error: ошибка
*/
/*func (adapter *BlockchainAdapter) TryLoadSavedBlockchain() (bool, error) {
	err := adapter.blockchain.LoadSavedBlockchain()
	if err != nil {
		return false, fmt.Errorf("Can not load saved blockchain: %v", err)
	}

	return true, nil
}*/

/*
TryNetworkLoadBlockchain попытка загрузить блокчейн из сети

Возвращает:
  - bool: успех/неуспех
  - error: ошибка
*/
/*func (adapter *BlockchainAdapter) TryNetworkLoadBlockchain() (bool, error) {
	return false, nil
}*/

/*
AlreadyExistBlockFilter сравнивает хэш приходящего
блока, и последнего блока в блокчейне, если они совпали,
значит пришедший по сети блок был только что создан самим узлом

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - input <-chan *block.Block: поступающие блоки

Возвращает:
  - chan *block.Block: отфильтрованные блоки
*/
func (blockchain *Blockchain) AlreadyExistBlockFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block {
	output := make(chan *block.Block)

	go func() {
		for {
			select {
			case blk := <-input:
				log.Printf("<blockchain.go> Блок пришел на фильтрацию")
				// Проверка по хэшу, был ли этот блок записан только что
				if blockchain.IsAlreadyExistBlock(blk) {
					log.Printf("<blockchain.go> Этот блок был только что сохранент, фильтр не пройден")
					continue
				}
				log.Printf("<blockchain.go> Фильтр пройден, отправка блока для сохранение на диск")
				output <- blk
			case <-ctx.Done():
				// Корректное завершение функции
				close(output)
				return
			}
		}
	}()

	return output
}

/*
BlockSaveProcess принимает канал с блоками и сохраняет
все приходящие блоки, ошибки записи поступают в выходной кана

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - input <-chan *block.Block: поступающие блоки

Возвращает:
  - chan error: ошибки
*/
func (blockchain *Blockchain) BlockSaveProcess(ctx context.Context, input <-chan *block.Block) chan *block.Block {
	output := make(chan *block.Block)

	// Фоновый процесс чтения и записи блоков
	go func() {
		for {
			select {
			case blk := <-input:
				log.Printf("<blockchain_adapter.go> Получен блок для сохранения на диск")
				// Чтение канала с блоками и запись блока на диск
				err := blockchain.AddBlockToBlockchain(blk)
				if err != nil {
					log.Printf("<blockchain.go> Ошибка сохранения блока на диск: %v", err)
					continue
				}
				log.Printf("<blockchain.go> Блок успешно записан в блокчейн на диске, отправка блока в сеть")
				output <- blk
			case <-ctx.Done():
				// Корректное завершение функции
				close(output)
				return
			}
		}
	}()

	return output
}

/*
GetBalance подсчитывает баланс кошелька,
итерируясь по всему сохраненному блокчейну

Аргументы:
  - address []byte: адрес кошелька

Возвращает:
  - int32: баланс кошелька
  - error: ошибка
*/
func (blockchain *Blockchain) GetBalance(address []byte) (int32, error) {
	// Словари с фходами и выходами для отсева занятых выходов
	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	// Итерируемся по транзакциям
	blockchain.TransactionIterator(
		func(t *transaction.Transaction) {
			// Добавление входов
			for _, in := range t.Inputs {
				inputs[string(in.PreviousOutputHash)] = struct{}{}
			}

			// Отсев соединений вход-выход
			for _, out := range t.Outputs {
				if _, ok := inputs[string(out.Hash)]; !ok {
					// Если хэш выхода не используется входом, значит добавляем в словарь
					outputs[string(out.Hash)] = &out
				} else {
					// Если хэш выхода занимает вход, удаляем обоих из словарей
					delete(outputs, string(out.Hash))
					delete(inputs, string(out.Hash))
				}
			}

		},
	)

	// Подсчитываем все значения
	res := 0
	for _, val := range outputs {
		if bytes.Equal(val.RecipientAddress, address) {
			res += val.Value
		}
	}

	return int32(res), nil
}

/*
GetFreeTransactionsOutputs находит все свободные
выходы транзакций, итерируясь по всему блокчейну на диске

Возвращает:
  - []*transaction.TransactionOutput: слайс транзакций
  - error: ошибка
*/
func (blockchain *Blockchain) GetFreeTransactionsOutputs() ([]*transaction.TransactionOutput, error) {
	// Словари с фходами и выходами для отсева занятых выходов
	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	// Итерируемся по транзакциям
	blockchain.TransactionIterator(
		func(t *transaction.Transaction) {
			// Добавление входов
			for _, in := range t.Inputs {
				inputs[string(in.PreviousOutputHash)] = struct{}{}
			}

			// Отсев соединений вход-выход
			for _, out := range t.Outputs {
				if _, ok := inputs[string(out.Hash)]; !ok {
					// Если хэш выхода не используется входом, значит добавляем в словарь
					outputs[string(out.Hash)] = &out
				} else {
					// Если хэш выхода занимает вход, удаляем обоих из словарей
					delete(outputs, string(out.Hash))
					delete(inputs, string(out.Hash))
				}
			}

		},
	)

	// Создаем список выходов
	res := make([]*transaction.TransactionOutput, len(outputs))
	i := 0
	for _, val := range outputs {
		res[i] = val
		i += 1
	}

	return res, nil
}

/*
TransactionIterator обходит все транзакции блокчейна
в обратном порядке и к каждой транзакции применяет заданную функцию

Аргументы:
  - f func(*transaction.Transaction): пользовательская функция

Возвращает:
  - error: ошибка
*/
func (blockchain *Blockchain) TransactionIterator(f func(*transaction.Transaction)) error {
	iter, err := blockchain.CreateIterator()
	if err != nil {
		return fmt.Errorf("Can not create iterator by blockchain")
	}

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		// Получение текущего блока
		currentBlock, err := iter.Current()
		if err != nil {
			return fmt.Errorf("Can not get current block: %v", err)
		}

		// Извлечение списка транзакций
		transactions, err := transaction.DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return fmt.Errorf("Transaction deserialization failed: %v", err)
		}

		// Инверсное итерирование по транзакциям c применением функции к каждой транзакции
		for i := len(transactions) - 1; i >= 0; i -= 1 {
			f(transactions[i])
		}

		// Переход к следующему блоку
		iter.Next()
	}

	return nil
}

/*
=======================================================
================ Итератор по блокчейну ================
=======================================================
*/
func (blockchain *Blockchain) CreateIterator() (Iterator[*block.Block], error) {
	var iterator blockchainIterator[*block.Block]

	tip, err := blockchain.storage.BlockchainGetTip()
	if err != nil {
		return nil, fmt.Errorf("Can not create iterator: %v", err)
	}

	iterator.currentHash = tip
	iterator.blockchain = blockchain

	return &iterator, nil
}

/* Структура итератора по блокчейну */
type blockchainIterator[T any] struct {
	blockchain  *Blockchain
	currentHash []byte
}

func (i *blockchainIterator[T]) Next() (*block.Block, error) {
	block, err := i.blockchain.storage.GetExistBlockByHash(i.currentHash)
	if err != nil {
		return nil, fmt.Errorf("Iterator can not load next element: %v", err)
	}

	block, err = i.blockchain.storage.GetExistBlockByHash(block.PrevBlockHash)
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
	block, err := i.blockchain.storage.GetExistBlockByHash(i.currentHash)
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
