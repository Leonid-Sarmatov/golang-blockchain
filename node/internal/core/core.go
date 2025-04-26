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
	//BlockSaveProcess(ctx context.Context, input <-chan *block.Block) chan *block.Block
	/* Запуск процесса отброса существующих блоков */
	AlreadyExistBlockFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block
	/* Проверка блока на его существание в блокчейне */
	IsAlreadyExistBlock(b *block.Block) bool
	/* Сохранение блока на диск */
	AddBlockToBlockchain(b *block.Block) error
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
	//TransactionListnerProcess(context.Context, <-chan *transaction.Transaction) chan []*transaction.Transaction
	/* Запуск процесса майнинга, по сигналу майнинг прерывается */
	MiningProcess(context.Context, <-chan *transaction.Transaction, <-chan *block.Block) chan *block.Block
}

type powChecker interface {
	/* Функция проверки доказательства работы*/
	Check(blk *block.Block) (bool, error)
}

type hashCalculator interface {
	/* Функция получения хэша */
	HashCalculate(data []byte) []byte
}

// type replicator interface {

// }

type Core struct {
	blockchain
	transactionReceiver
	blockReceiver
	blockTransmitter
	miner
	powChecker
	hashCalculator
	//transactionOutputPool
	//replicator
}

/*
NewCore конструктор для ядра

Возвращает:
  - *Core: экземпляр структуры ядра
*/
func NewCore(
	br blockReceiver,
	bt blockTransmitter,
	tr transactionReceiver,
	b blockchain,
	m miner,
	pc powChecker,
	hc hashCalculator,
	//top transactionOutputPool,
) *Core {
	return &Core{
		blockchain:          b,
		transactionReceiver: tr,
		blockReceiver:       br,
		blockTransmitter:    bt,
		miner:               m,
		powChecker:          pc,
		hashCalculator:      hc,
		//transactionOutputPool: top,
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
	trnansactionRecChan := core.transactionReceiver.TransactionReceiverProcess("transactions1")

	// Запуск процесса получения блоков из сети, и получение канала с приходящими блоками
	networkBlockRecChan := core.blockReceiver.BlockReceiverProcess("blocks1")

	// Запуск процессов фильтрации блоков
	filterChan := core.CheckPOWFilter(context.Background(), networkBlockRecChan)
	blockFilterChan := core.AlreadyExistBlockFilter(context.Background(), filterChan)

	// Разделение канала с блоками, первый для сохранения блоков, второй для процесса отмены майнинга
	blocksCnahs := Tee(blockFilterChan, 2)

	// Запуск процесса майнинга блоков
	minerBlkChan := core.miner.MiningProcess(context.Background(), trnansactionRecChan, blocksCnahs[0])

	// Запуск процесса сохранения блоков на диск
	core.BlockSaveProcess(context.Background(), blocksCnahs[1])

	// Запуск процесса отправки блоков в сеть
	core.blockTransmitter.BlockTransmitterProcess(context.Background(), minerBlkChan, "blocks1")

	return nil
}

func (core *Core) Init2() error {

	//
	return nil
}

/*
AlreadyExistBlockFilter сверяет последний записанный 
на диск блок с заданным блоком, если они совпадают то блок отбрасывается

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - input <-chan *block.Block: приходящие блоки

Возвращает:
  - chan *block.Block: канал с блоками, которые прошли фильтр
*/
func (core *Core) AlreadyExistBlockFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block {
	output := make(chan *block.Block)

	go func() {
		for {
			select {
			case blk := <-input:
				log.Printf("<blockchain.go> Блок пришел на фильтрацию")
				// Проверка по хэшу, был ли этот блок записан только что
				if core.blockchain.IsAlreadyExistBlock(blk) {
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
AlreadyExistBlockFilter проверяет доказательство работы,
если оно не корректно то блок отбрасывается

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - input <-chan *block.Block: приходящие блоки

Возвращает:
  - chan *block.Block: канал с блоками, которые прошли фильтр
*/
func (core *Core) CheckPOWFilter(ctx context.Context, input <-chan *block.Block) chan *block.Block {
	output := make(chan *block.Block)

	go func() {
		for {
			select {
			case blk := <-input:
				log.Printf("<core.go> Блок пришел на фильтрацию")
				// Проверка POW
				ok, err := core.powChecker.Check(blk)
				if err != nil {
					log.Printf("<core.go> Не удалось проверить POW")
					continue
				}
				if !ok {
					log.Printf("<core.go> В блоке некорректный POW, фильтр не пройден")
					continue
				}
				log.Printf("<core.go> Фильтр пройден, отправка блока дальше")
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
*/
func (core *Core) BlockSaveProcess(ctx context.Context, input <-chan *block.Block) {
	// Фоновый процесс чтения и записи блоков
	go func() {
		for {
			select {
			case blk := <-input:
				log.Printf("<core.go> Получен блок для сохранения на диск")
				// Чтение канала с блоками и запись блока на диск
				err := core.blockchain.AddBlockToBlockchain(blk)
				if err != nil {
					log.Printf("<core.go> Ошибка сохранения блока на диск: %v", err)
					continue
				}
				log.Printf("<core.go> Блок успешно записан в блокчейн на диске")
			case <-ctx.Done():
				// Корректное завершение функции
				return
			}
		}
	}()

}

func Tee[T any](input <-chan T, n int) []chan T {
	outputs := make([]chan T, n)
	for i := 0; i < n; i += 1 {
		outputs[i] = make(chan T)
	}

	go func() {
		for value := range input {
			for i := 0; i < n; i += 1 {
				outputs[i] <- value
			}
		}

		for _, ch := range outputs {
			close(ch)
		}
	}()

	return outputs
}
