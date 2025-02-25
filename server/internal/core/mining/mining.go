package mining

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/internal/services/transaction_queue"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/hash_calulator"
	"golang_blockchain/pkg/iterator"
	"golang_blockchain/pkg/proof_of_work"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type proofOfWorkCheker interface {
	Check(b *block.Block, value int, hc proofofwork.HashCalulator) (bool, error)
}

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

/*
workQueue очередь работ для майнеров,
работа для майнера состоит из подсчета
proof-of-work для двух транзакций ()
*/
type transactionQueue interface {
	PushTransaction(t *transaction.Transaction) error
	PullTransaction() (*transaction.Transaction, error)
}

type mediator interface {
	CreateBlocksIterator() (iterator.Iterator[*block.Block], error)
	CreateNewCoinBaseTransaction(reward int, address, key []byte) (*transaction.Transaction, error)
	AddBlock(b *block.Block, pwValue int) error
	CreateBlock(data []byte) (*block.Block, error)
}

/*
Контроллер странзакций
*/
type MinerController struct {
	hashCalculator proofofwork.HashCalulator
	queue          transactionQueue
	cheker         proofOfWorkCheker
	mediator       mediator
}

/* Конструктор */
func NewMinerController(m mediator) (*MinerController, error) {
	var minerController MinerController
	minerController.mediator = m

	hc := hashcalulator.NewHashCalculator()
	minerController.hashCalculator = hc

	q := transactionqueue.NewTransactionQueue()
	minerController.queue = q

	ch := proofofwork.NewProofOfWorkCheker()
	minerController.cheker = ch

	return &minerController, nil
}

/*
CreateNewCoinBase добавляет транзакцию в очередь на обработку

Аргументы:
  - *transaction.Transaction: указатель на транзакцию

Возвращает:
  - error: ошибка
*/
func (controller *MinerController) AddTransactionToProcessing(t *transaction.Transaction) error {
	return controller.queue.PushTransaction(t)
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
func (controller *MinerController) GetWorkForMining(rewardAddress []byte) ([]byte, error) {
	// Берем основную транзакцию из очереди
	mainTransaction, err := controller.queue.PullTransaction()
	if err != nil {
		return nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	// Создаем транзакцию вознаграждения
	rewardTransaction, err := controller.mediator.CreateNewCoinBaseTransaction(1, rewardAddress, rewardAddress)
	if err != nil {
		return nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	// Формируем список транзакций и парсим его в байтовый слайс
	transactions, err := transaction.SerializeTransactions([]transaction.Transaction{*rewardTransaction, *mainTransaction})
	if err != nil {
		return nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	// Формируем блок с транзакциями
	b, err := controller.mediator.CreateBlock(transactions)
	if err != nil {
		return nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	// Сериализуем блок в байты
	bytesBlock, err := b.SerializeBlock()
	if err != nil {
		return nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	return bytesBlock, nil
}

/*
SendCompletedWork

Аргументы:
  - bytesRewardTransaction []byte: транзакция вознаграждения
  - bytesMainTransaction []byte: основная (полезная) транзакция
  - bytesRewardTransaction int: pow транзакции вознаграждения
  - bytesMainTransaction int: pow основной транзакции

Возвращает:
  - error: ошибка
*/
func (controller *MinerController) SendCompletedWork(bytesBlock []byte, POW int) error {
	// Преобразуем байты в список транзакций
	b, err := block.DeserializeBlock(bytesBlock)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	POWOK, err := controller.cheker.Check(b, POW, controller.hashCalculator)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	// Если все прошло успешно, записываем блок в блокчейн
	if POWOK {
		err = controller.mediator.AddBlock(b, POW)
		if err != nil {
			return fmt.Errorf("The work cannot be accepted: %v", err)
		}
		return nil
	}
	return fmt.Errorf("The work cannot be accepted: proof-of-work is not valid, %v", POW)
}
