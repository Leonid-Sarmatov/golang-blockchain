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
	Exec(blk *block.Block, cancel <-chan int) (int, error)
}

type blockchainStorage interface {
	/* Функция загружающая из хранилища хэш последнего блока */
	BlockchainGetTip() ([]byte, error)
}

type Miner struct {
	checker powChecker
	solver  powSolver
	Storage blockchainStorage
}

func NewMiner(checker powChecker, solver powSolver, storage blockchainStorage) *Miner {
	return &Miner{
		checker: checker,
		solver:  solver,
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
	startMining <-chan int,
) chan []*transaction.Transaction {
	// Канал пакетом транзакций для будущего блока
	trnsCh := make(chan []*transaction.Transaction)

	// Фоновый процесс получения транзакций и отправки их на майнинг
	go func() {
		// Пакет транзакций
		trns := make([]*transaction.Transaction, 0)
		for {
			select {
			case <-startMining:
				// Не формируем пакет, если транзакций нет вообще
				if len(trns) == 0 {
					continue
				}
				// Копируем пакет транзакций в буфер, отправляем буфер на майнинг, и сбрасываем пакет
				buffer := make([]*transaction.Transaction, len(trns))
				copy(buffer, trns)
				trnsCh <- buffer
				trns = nil
				log.Printf("<miner.go> Пришел сигнал на начало майнинга: пакет транзакций отправлен на обработку")
			case trn := <-inputTransactions:
				log.Printf("<miner.go> Пришла транзакция...")
				// Если nil то переинициализируем
				if trns == nil {
					log.Printf("<miner.go> Пакет был нулевым, инициализация пакета (массива транзакций)")
					trns = make([]*transaction.Transaction, 0)
				}
				// Аккумулируем транзакции в пакет
				log.Printf("<miner.go> Сохранение транзакции в пакет (массив транзакций)")
				trns = append(trns, trn)
				// Если накопилось слишком много транзакций, начинаем майнить
				if len(trns) >= 5 {
					log.Printf("<miner.go> Транзакций накопилось слишком много: майнинг без сигнала")
					// Копируем пакет транзакций в буфер, отправляем буфер на майнинг, и сбрасываем пакет
					buffer := make([]*transaction.Transaction, len(trns))
					copy(buffer, trns)
					trnsCh <- buffer
					trns = nil
				}
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
	transactionPackets <-chan []*transaction.Transaction,
	cancelMining <-chan int,
) chan *block.Block {
	// Канал с блоками на отправку в сеть
	blks := make(chan *block.Block)

	// Фоновый процесс майнинга блоков
	go func() {
		for {
			select {
			case pac := <-transactionPackets:
				log.Printf("<miner.go> Получен пакет транзакций для майнига блока")
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

				blk, err := block.NewBlock(slice, tip)
				if err != nil {
					log.Printf("<miner.go> Не удалось сформировать новый блок! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Сформирован блок с транзакциями!")

				pow, err := miner.solver.Exec(blk, cancelMining)
				if err != nil {
					log.Printf("<miner.go> Ошибка при подсчете proof-of-work! Майнинг отменен. Ошибка: %v", err)
					continue
				}
				log.Printf("<miner.go> Для сформированного блока успешно посчитан proof-of-work!")

				if pow >= 0 {
					blk.ProofOfWorkValue = pow
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

/*func (miner *Miner) Mining(
	ctx context.Context,
	inputTransactions <-chan *transaction.Transaction,
	startMining <-chan int,
	abortMining <-chan int,
) chan *block.Block {
	// Канал с готовыми блоками
	outputCh := make(chan *block.Block)

	// Канал пакетом транзакций для будущего блока
	trnsCh := make(chan []*transaction.Transaction)

	// Канал старта майнинга надо разветвить
	starts := Tee(startMining, 2)

	// // Контексты для корректного завершения работы
	ctx1, close1 := context.WithCancel(context.Background())
	ctx2, close2 := context.WithCancel(context.Background())

	// Фоновый процесс получения транзакций и отправки их на майнинг
	go func() {
		// Пакет транзакций
		trns := make([]*transaction.Transaction, 0)
		for {
			select {
			case <-starts[0]:
				// Не формируем пакет, если транзакций нет вообще
				if len(trns) == 0 {
					continue
				}
				// Копируем пакет транзакций в буфер, отправляем буфер на майнинг, и сбрасываем пакет
				buffer := make([]*transaction.Transaction, len(trns))
				copy(buffer, trns)
				trnsCh <- buffer
				trns = nil
			case trn := <-inputTransactions:
				// Если nil то переинициализируем
				if trns == nil {
					trns = make([]*transaction.Transaction, 0)
				}
				// Аккумулируем транзакции в пакет
				trns = append(trns, trn)
			case <-ctx1.Done():
				// Обработка корректного завершения
				return
			}
		}
	}()

	// Фоновый процесс формирования блока из пакета транзакций
	go func() {
		for {
			select {
			case <-starts[1]:
				// Получение кончика блокчейна, который станет предудущим хэшом формируемого блока
				tip, err := miner.Storage.BlockchainGetTip()
				if err != nil {
					log.Printf("<miner.go> Не удалось получить кончик! Майнинг отменен. Ошибка: %v", err)
					continue
				}

				// Пакет транзакций это полезная нагрузка блока, трансформируем в байтовый слайс
				slice, err := transaction.SerializeTransactions(<-trnsCh)
				if err != nil {
					log.Printf("<miner.go> Не удалось сериализовать транзакции! Майнинг отменен. Ошибка: %v", err)
					continue
				}

				blk, err := block.NewBlock(slice, tip)
				if err != nil {
					log.Printf("<miner.go> Не удалось сформировать новый блок! Майнинг отменен. Ошибка: %v", err)
					continue
				}

				pow, err := miner.solver.Exec(blk, abortMining)
				if err != nil {
					log.Printf("<miner.go> Ошибка при подсчете proof-of-work! Майнинг отменен. Ошибка: %v", err)
					continue
				}

				if pow >= 0 {
					blk.ProofOfWorkValue = pow
					outputCh <- blk
					log.Printf("<miner.go> Блок успешно создан. Ожидает запись и отправку в сеть")
				} else {
					log.Printf("<miner.go> Майнинг отменен. Блок никуда не пойдет")
				}
			case <-ctx2.Done():
				// Обработка корректного завершения
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Корректно завершаем все процессы
				close1()
				close2()
				close(outputCh)
				close(trnsCh)
			}
		}
	}()

	return outputCh
}*/

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
