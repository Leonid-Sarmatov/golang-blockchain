package blockchainadapter

import (
	"context"
	"fmt"
	"log"
	"bytes"
	//"node/internal/adapters/pow"
	//"node/internal/adapters/storage"
	"node/internal/block"
	"node/internal/blockchain"
	"node/internal/transaction"
)

/*
Адаптер для высокоуровневой работы с блокчейном
*/
type BlockchainAdapter struct {
	blockchain *blockchain.Blockchain
}

/*
NewBlockchainAdapter конструктор для
высокоуровневого адаптера над блокчейном
*/
func NewBlockchainAdapter(b *blockchain.Blockchain) *BlockchainAdapter {
	return &BlockchainAdapter{
		blockchain: b,
	}
}

/*
Init инициализация адаптера

Возвращает:
  - error: ошибка
*/
func (adapter *BlockchainAdapter) Init() error {
	// // Инициализация хранилища
	// storage := storage.NewBBoltDBDriver()
	// // Инициализация хэш калькулятора
	// ch := pow.NewHashCalculator()
	// // Вызыв конструктора блокчейна и инициализация
	// adapter.blockchain = blockchain.NewBlockchain(storage, ch)
	return nil
}

/*
TryLoadSavedBlockchain попытка загрузить блокчейн
из внутреннего хранилища данного узла

Возвращает:
  - bool: успех/неуспех
  - error: ошибка
*/
func (adapter *BlockchainAdapter) TryLoadSavedBlockchain() (bool, error) {
	err := adapter.blockchain.LoadSavedBlockchain()
	if err != nil {
		return false, fmt.Errorf("Can not load saved blockchain: %v", err)
	}

	return true, nil
}

/*
TryNetworkLoadBlockchain попытка загрузить блокчейн
из внутреннего хранилища данного узла

Возвращает:
  - bool: успех/неуспех
  - error: ошибка
*/
func (adapter *BlockchainAdapter) TryNetworkLoadBlockchain() (bool, error) {
	return false, nil
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
func (adapter *BlockchainAdapter) BlockSaveProcess(ctx context.Context, input <-chan *block.Block) chan error {
	output := make(chan error)

	// Фоновый процесс чтения и записи блоков
	go func() {
		for {
			select {
			case blk := <-input:
				// Чтение канала с блоками и запись блока на диск
				err := adapter.blockchain.AddBlockToBlockchain(blk)
				if err != nil {
					output <- fmt.Errorf("Can not add block: %v", err)
				}
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
AlreadyExistBlockFilter сравнивает хэш приходящего
блока, и последнего блока в блокчейне, если они совпали,
значит пришедший по сети блок был только что создан самим узлом

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - input <-chan *block.Block: поступающие блоки

Возвращает:
  - chan *block.Block: отфильтрованные блоки
*/
func (adapter *BlockchainAdapter) AlreadyExistBlockFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block {
	output := make(chan *block.Block)

	go func() {
		for {
			select {
			case blk := <-input:
				// Проверка по хэшу, был ли этот блок записан только что
				if adapter.blockchain.IsAlreadyExistBlock(blk) {
					log.Printf("<blockchain_adapter.go> Блок только что был записан, пропускаем")
					continue
				}
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

func (adapter *BlockchainAdapter) GetBalance(address []byte) (int32, error) {
	iter, err := adapter.blockchain.CreateIterator()
	if err != nil {
		return -1, fmt.Errorf("Can not create iterator by blockchain")
	}

	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentBlock, err := iter.Current()
		if err != nil {
			return -1, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		// Расшифровываем информацию блока, извлекаем список транзакций
		transactions, err := transaction.DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return -1, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}

		// Определяем входы входящие в блок и выходы выходящие из блока
		ins := make(map[string]interface{})
		outs := make(map[string]*transaction.TransactionOutput)
		for _, tx := range transactions {
			for _, out := range tx.Outputs {
				outs[string(out.Hash)] = &out
			}

			for _, in := range tx.Inputs {
				if _, ok := outs[string(in.PreviousOutputHash)]; ok {
					delete(outs, string(in.PreviousOutputHash))
				} else {
					ins[string(in.PreviousOutputHash)] = 0
				}
			}
		}

		// Запоминаем все входы
		for hash, _ := range ins {
			inputs[hash] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for hash, out := range outs {
			// Если хэш выхода не используется входом, значит добавляем в словарь
			if _, ok := inputs[hash]; !ok {
				outputs[hash] = out
			} else {
				delete(outputs, hash)
			}
		}

		iter.Next()
	}

	// Подсчитываем все значения
	res := 0
	for _, val := range outputs {
		if bytes.Equal(val.RecipientAddress, address) {
			res += val.Value
		}
	}

	return int32(res), nil
}

func (adapter *BlockchainAdapter) GetFreeTransactionsOutputs() ([]*transaction.TransactionOutput, error) {
	iter, err := adapter.blockchain.CreateIterator()
	if err != nil {
		return nil, fmt.Errorf("Can not create iterator by blockchain")
	}

	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentBlock, err := iter.Current()
		if err != nil {
			return nil, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		// Расшифровываем информацию блока, извлекаем список транзакций
		transactions, err := transaction.DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return nil, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}

		// Определяем входы входящие в блок и выходы выходящие из блока
		ins := make(map[string]interface{})
		outs := make(map[string]*transaction.TransactionOutput)
		for _, tx := range transactions {
			for _, out := range tx.Outputs {
				outs[string(out.Hash)] = &out
			}

			for _, in := range tx.Inputs {
				if _, ok := outs[string(in.PreviousOutputHash)]; ok {
					delete(outs, string(in.PreviousOutputHash))
				} else {
					ins[string(in.PreviousOutputHash)] = 0
				}
			}
		}

		// Запоминаем все входы
		for hash, _ := range ins {
			log.Printf("Вход: HASH = %x", hash)
			inputs[hash] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for hash, out := range outs {
			// Если хэш выхода не используется входом, значит добавляем в словарь
			if _, ok := inputs[hash]; !ok {
				outputs[hash] = out
			} else {
				delete(outputs, hash)
			}
		}

		iter.Next()
	}

	// Создаем список выходов
	res := make([]*transaction.TransactionOutput, len(outputs))
	i := 0
	for _, val := range outputs {
		res[i] = val
		i += 1
	}

	return res, nil
}