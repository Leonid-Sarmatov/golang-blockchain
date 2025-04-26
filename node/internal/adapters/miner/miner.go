package miner

import (
	"context"
	"log"
	"node/internal/block"
	"node/internal/transaction"
)

type powChecker interface {
	/* Функция проверки доказательства работы*/
	Check(blk *block.Block) (bool, error)
}

type powSolver interface {
	/* Функция поиска доказательства работы с возможностью отмены */
	Exec(blk *block.Block, cancel <-chan interface{}) error
}

type blockchainStorage interface {
	/* Функция загружающая из хранилища хэш последнего блока */
	BlockchainGetTip() ([]byte, error)
}

type hashCalculator interface {
	/* Функция получения хэша */
	HashCalculate(data []byte) []byte
}

/*
TransactionOutputPool описывает интерфейс для
пулла доступных выходов транзакций
*/
type transactionOutputPool interface {
	/* Функция пробует заблокировать выход */
	BlockOutput(output transaction.TransactionOutput) error
	/* Добавляет новык выходы в пулл */
	AddOutput(output transaction.TransactionOutput) error
	/* Возвращает список всех транзакций с незаблокированными выходами */
	GetAllUnlockOutputs() ([]*transaction.TransactionOutput, error)
}

type Miner struct {
	checker powChecker
	solver  powSolver
	pool transactionOutputPool
	calc hashCalculator
	Storage blockchainStorage
}

func NewMiner(checker powChecker, solver powSolver, storage blockchainStorage, pool transactionOutputPool, calc hashCalculator) *Miner {
	return &Miner{
		checker: checker,
		solver:  solver,
		pool: pool,
		calc: calc,
		Storage: storage,
	}
}

func (miner *Miner) Init() error {
	return nil
}

/*
TransactionListnerProcess создает процесс обработки приходящих
транзакций и сборку транзакций в пакет для создания блока
с этими транзакциями

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - inputTransactions <-chan *transaction.Transaction: приходящие извне транзакции
  - startMining <-chan int: сигнал на запаковку и отправку пакета на формирование блока

Возвращает:
  - chan []*transaction.Transaction: канал с пакетами дляформирования блоков
*/
func (miner *Miner) TransactionListnerProcess(
	ctx context.Context,
	inputTransactions <-chan *transaction.Transaction,
) chan []*transaction.Transaction {
	// Канал пакетом транзакций для будущего блока
	trnsCh := make(chan []*transaction.Transaction)

	// Фоновый процесс получения транзакций и отправки их на майнинг
	go func() {
		for {
			select {
			case trn := <-inputTransactions:
				log.Printf("<miner.go> Пришла транзакция...")
				// Создание базисной транзакции для вознаграждения данного майнера
				trnR, err := transaction.NewCoinbaseTransaction(
					1, []byte("PetrovichMiner"), []byte("PetrovichMiner"), 
					miner.calc, miner.pool,
				)
				if err != nil {
					log.Printf("<miner.go> Не удалось создать транзакцию вознаграждения майнера")
					continue
				}
				// Отправка транзакции от мем-пулла и транзакции вознаграждения на майнинг
				log.Printf("<miner.go> Отправка транзакций на майнинг нового блока")
				trnsCh <- []*transaction.Transaction{trn, trnR}

			case <-ctx.Done():
				// Обработка корректного завершения
				return
			}
		}
	}()

	return trnsCh
}

/*
MiningProcess создает процесс создания, майнинга нового блока
для сети и отмены майнинга при поступлении нового блока

Аргументы:
  - ctx context.Context: контекст для корректной остановки работы
  - transactionPackets <-chan []*transaction.Transaction: приходящие извне пакеты
  - ancelMining <-chan int: сигнал на отмену майнинга блока

Возвращает:
  - chan *block.Block: канал с блоками на отправку в сеть
*/
func (miner *Miner) MiningProcess(
	ctx context.Context, 
	inputTransactions <-chan *transaction.Transaction, 
	inputBlock <-chan *block.Block,
) chan *block.Block {
	// Канал с блоками на отправку в сеть
	blks := make(chan *block.Block)

	// Фоновый процесс майнинга блоков
	go func() {
		for {
			select {
			case trn := <-inputTransactions:
				log.Printf("<miner.go> Пришла транзакция...")
				// Создание базисной транзакции для вознаграждения данного майнера
				trnR, err := transaction.NewCoinbaseTransaction(
					1, []byte("Miner"), []byte("Miner"), 
					miner.calc, miner.pool,
				)
				if err != nil {
					log.Printf("<miner.go> Не удалось создать транзакцию вознаграждения майнера! Майнинг отменен. Ошибка: %v", err)
					continue
				}

				// Формирование пакета транзакций
				pac := []*transaction.Transaction{trn, trnR}
				
				// Получение кончика блокчейна, который станет предудущим хэшом формируемого блока
				tip, err := miner.Storage.BlockchainGetTip()
				if err != nil {
					log.Printf("<miner.go> Не удалось получить кончик! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Кончик успешно получен!")

				// Пакет транзакций это полезная нагрузка блока, трансформируем в байтовый слайс
				slice, err := transaction.SerializeTransactions(pac)
				if err != nil {
					log.Printf("<miner.go> Не удалось сериализовать транзакции! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Транзакции успешно сериализованы в байтовый слайс!")

				// Конструктур блока
				blk, err := block.NewBlock(slice, tip)
				if err != nil {
					log.Printf("<miner.go> Не удалось сформировать новый блок! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Сформирован блок с транзакциями!")

				// Фоновый процесс отмены майнинга
				cancel := make(chan interface{})
				defer close(cancel)
				go func () {
					for b := range inputBlock {
						if b.TimeOfCreation < blk.TimeOfCreation {
							cancel <- struct{}{}
						}
					}
				}()

				err = miner.solver.Exec(blk, cancel)
				if err != nil {
					log.Printf("<miner.go> Ошибка при подсчете proof-of-work! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Для сформированного блока успешно посчитан proof-of-work")

				if blk.ProofOfWorkValue >= 0 {
					// Задаем POW блоку и отправляем на сохранение
					blks <- blk
					log.Printf("<miner.go> Блок успешно создан. Ожидает запись и отправку в сеть")
				} else {
					log.Printf("<miner.go> Майнинг отменен. Блок никуда не пойдет")
				}
			case <-ctx.Done():
				// Обработка корректного завершения
				return
			}
		}
	}()

	return blks
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
