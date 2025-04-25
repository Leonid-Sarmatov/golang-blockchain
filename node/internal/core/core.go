package core

import (
	"context"
	"log"
	"node/internal/block"
	"node/internal/transaction"
)

type blockchain interface {
	/* Загрузка блокчейна из локального хранилища */
	TryLoadSavedBlockchain() error
	/* Загрузка блокчейна из сети */
	//TryNetworkLoadBlockchain() (error)
	/* Запуск процесса сохранения блока из канала */
	BlockSaveProcess(ctx context.Context, input <-chan *block.Block) chan *block.Block
	/* Запуск процесса отброса существующих блоков */
	AlreadyExistBlockFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block
}

type transactionReceiver interface {
	/* Запуск процесса приема транзакций от мем-пулла */
	TransactionReceiverProcess(id string) chan *transaction.Transaction
}

type blockTransmitter interface {
	/* Запуск процесса отправки созданных блоков в сеть */
	BlockTransmitterProcess(ctx context.Context, blks <-chan *block.Block, id string)
}

type blockReceiver interface {
	/* Запуск процесса приема блоков, созданных другими узлами сети*/
	BlockReceiverProcess(id string) chan *block.Block
}

type miner interface {
	/* Запуск процесса прослушивания транзакций, по сигналу отправляет пакет транзакций на майнинг */
	TransactionListnerProcess(context.Context, <-chan *transaction.Transaction, <-chan int) chan []*transaction.Transaction
	/* Запуск процесса майнинга, по сигналу майнинг прерывается */
	MiningProcess(context.Context, <-chan []*transaction.Transaction, <-chan int) chan *block.Block
}

// type replicator interface {

// }

type Core struct {
	blockchain
	transactionReceiver
	blockReceiver
	blockTransmitter
	miner
	//replicator
}

/*
NewCore конструктор для ядра

Возвращает:
  - *Core: экземпляр структуры ядра
*/
func NewCore(br blockReceiver, bt blockTransmitter, tr transactionReceiver, b blockchain, m miner) *Core {
	return &Core{
		blockchain:          b,
		transactionReceiver: tr,
		blockReceiver:       br,
		blockTransmitter:    bt,
		miner:               m,
	}
}

/*
GetWorkForMining возвращает подготовленную для
майнинга работу: транзакция вознаграждения и
главная транзакция (в виде байтовых слайсов)

Аргументы:
  - rewardAddress []byte: адрес получателя вознаграждения

Возвращает:
  - []byte: транзакция вознаграждения
  - []byte: главная транзакция
  - error: ошибка
*/

/*
Init инициализирцет бизнес логику приложения

Возвращает:
  - error: ошибка
*/
func (core *Core) Init() error {
	// Запуск процесса получения транзакций, и получение канала с транзакциями из мем-пулла
	tranRecChan := core.transactionReceiver.TransactionReceiverProcess("transactions1")

	// Запуск процесса получения блоков из сети, и получение канала с приходящими блоками
	blRecChan := core.blockReceiver.BlockReceiverProcess("blocks1")

	// Запуск процесса фильтрации блоков, отбрасывание уже сохраненных блоков
	//blFilterCtx, blFilterCtxCancel := context.WithCancel(context.Background())
	blFilterChan := core.blockchain.AlreadyExistBlockFilter(context.Background(), blRecChan)

	// Разделение канала с блоками, первый для сохранения блоков, остальные для работы майнинга
	selfSaveChan := make(chan *block.Block)
	cancelMining := make(chan int)
	startMining := make(chan int)
	go func() {
		for {
			select {
			case blk := <-blFilterChan:
				log.Printf("1 > ")
				selfSaveChan <- blk
				log.Printf("1 >> ")
				cancelMining <- 1
				log.Printf("1 >>> ")
			}
		}
	}()

	// Запуск процесса получения транзакций и формирования пакетов для формирования блока
	packetsChan := core.miner.TransactionListnerProcess(context.Background(), tranRecChan, startMining)

	// Запуск процесса майнинга блоков
	minerBlkChan := core.miner.MiningProcess(context.Background(), packetsChan, cancelMining)

	// Запуск процесса сохранения блоков на диск
	networkSendBlkChan := core.blockchain.BlockSaveProcess(context.Background(), minerBlkChan)

	// Запуск процесса отправки блоков в сеть
	core.blockTransmitter.BlockTransmitterProcess(context.Background(), networkSendBlkChan, "blocks1")

	return nil
}
