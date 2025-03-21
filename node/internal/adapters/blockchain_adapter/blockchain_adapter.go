package blockchainadapter

import (
	"fmt"
	"node/internal/adapters/pow"
	"node/internal/adapters/storage"
	"node/internal/block"
	"node/internal/blockchain"
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
func NewBlockchainAdapter() *BlockchainAdapter {
	return &BlockchainAdapter{}
}

/*
Init инициализация адаптера

Возвращает:
  - error: ошибка
*/
func (adapter *BlockchainAdapter) Init() error {
	// Инициализация хранилища
	storage := storage.NewBBoltDBDriver()
	// Инициализация хэш калькулятора
	ch := pow.NewHashCalculator()
	// Вызыв конструктора блокчейна и инициализация
	adapter.blockchain = blockchain.NewBlockchain(storage, ch)
	return nil
}

/*
TryLoadSavedBlockchain попытка загрузить блокчейн
из внутреннего хранилища данного узла

Возвращает:
  - bool: успех/неуспех
  - error: ошибка
*/
func (adapter *BlockchainAdapter)TryLoadSavedBlockchain() (bool, error) {
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
func (adapter *BlockchainAdapter)TryNetworkLoadBlockchain() (bool, error) {
	return false, nil
}

func (adapter *BlockchainAdapter)BlockLoader(input <-chan *block.Block) chan error {
	output := make(chan error)

	// Фоновый процесс чтения и записи блоков
	go func() {
		defer close(output)

		for {
			// Чтение канала с блоками
			for blk := range input {
				err := adapter.blockchain.AddBlockToBlockchain(blk)
				if err != nil {
					output <- fmt.Errorf("Can not add block: %v", err)
				}
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			// log.Println("PubSub channel with transactions closed. Reconnecting...")
            // time.Sleep(1 * time.Second)
		}
	}()

	return output
}

