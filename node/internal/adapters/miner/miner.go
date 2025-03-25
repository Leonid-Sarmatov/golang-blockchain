package miner

import (
	"log"
	"node/internal/block"
	"node/internal/transaction"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
// type hashCalulator interface {
// 	HashCalculate(data []byte) []byte
// }

type powChecker interface {
	/* Функция проверки доказательства работы*/
	Chech(blk *block.Block) (bool, error)
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
	solver powSolver
	Storage  blockchainStorage
}

func NewMiner(checker powChecker, solver powSolver) *Miner {
	return &Miner{
		checker: checker,
		solver: solver,
	}
}

func (miner *Miner)Init() error {
	return nil
}

func (miner *Miner)Mining(
	inputTransactions <-chan *transaction.Transaction,
	startMining <-chan int,
	abortMining <-chan int,
	) chan *block.Block {
		// Канал с готовыми блоками 
		outputCh := make(chan *block.Block)

		// Канал пакетом транзакций для будущего блока
		trnsCh := make(chan []*transaction.Transaction)

		// Фоновый процесс получения транзакций и отправки их на майнинг
		go func() {
			// Пакет транзакций
			trns := make([]*transaction.Transaction, 0)
			for {
				select {
				case <- startMining:
					// Не формируем пакет, если транзакций нет вообще
					if len(trns) == 0 {
						continue
					}
					// Копируем пакет транзакций в буфер, отправляем буфер на майнинг, и сбрасываем пакет
					buffer := make([]*transaction.Transaction, len(trns))
					copy(buffer, trns)
					trnsCh <- buffer
					trns = nil
				case trn := <- inputTransactions:
					// Если nil то переинициализируем
					if trns == nil {
						trns = make([]*transaction.Transaction, 0)
					}
					// Аккумулируем транзакции в пакет
					trns = append(trns, trn)
				}
			}
		}()

		// Фоновый процесс формирования блока из пакета транзакций
		go func() {
			for {
				select {
				case <- startMining:
					// Получение кончика блокчейна, который станет предудущим хэшом формируемого блока
					tip, err := miner.Storage.BlockchainGetTip()
					if err != nil {
						log.Printf("<miner.go> Не удалось получить кончик! Майнинг отменен. Ошибка: %v", err)
						continue
					}

					// Пакет транзакций это полезная нагрузка блока, трансформируем в байтовый слайс
					slice, err :=transaction.SerializeTransactions(<-trnsCh)
					if err != nil {
						log.Printf("<miner.go> Не удалось сериализовать транзакции! Майнинг отменен. Ошибка: %v", err)
						continue
					}

					blk, err := block.NewBlock(slice, tip)
					if err != nil {
						log.Printf("<miner.go> Не удалось сформировать новый блок! Майнинг отменен. Ошибка: %v", err)
						continue
					}
					
					miner.solver.Exec(blk, abortMining)
				}
			}
		}()

	return outputCh
}

