package minercontroller

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/hash_calulator"
	"golang_blockchain/pkg/iterator"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type proofOfWorkCheker interface {
	Check(t *transaction.Transaction, value int) (bool, error)
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
	AddBlock(data []byte, pwValue int) error
}

/*
Контроллер странзакций
*/
type MinerController struct {
	hashCalculator hashCalulator
	queue          transactionQueue
	cheker         proofOfWorkCheker
	mediator       mediator
}

/**/
func NewMinerController(m mediator) (*MinerController, error) {
	var minerController MinerController
	minerController.mediator = m

	hc := hashcalulator.NewHashCalculator()
	minerController.hashCalculator = hc

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
func (controller *MinerController) GetWorkForMining(rewardAddress []byte) ([]byte, []byte, error) {
	rewardTransaction, err := controller.mediator.CreateNewCoinBaseTransaction(1, rewardAddress, rewardAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	mainTransaction, err := controller.queue.PullTransaction()
	if err != nil {
		return nil, nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	bytesRewardTransaction, err := rewardTransaction.TransactionToBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	bytesMainTransaction, err := mainTransaction.TransactionToBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("Can not return work for minig: %v", err)
	}

	return bytesRewardTransaction, bytesMainTransaction, nil
}

/*
SendCompletedWork

Аргументы:
  - bytesRewardTransaction []byte: адрес получателя вознаграждения
  - bytesMainTransaction []byte: адрес получателя вознаграждения
  - bytesRewardTransaction int: адрес получателя вознаграждения
  - bytesMainTransaction int: адрес получателя вознаграждения

Возвращает:
  - error: ошибка
*/
func (controller *MinerController) SendCompletedWork(
	bytesRewardTransaction, bytesMainTransaction []byte,
	rewardTransactionPOW, mainTransactionPOW int,
) error {
	rewartT := &transaction.Transaction{}
	err := rewartT.BytesToTransaction(bytesRewardTransaction)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	rewardPOWOK, err := controller.cheker.Check(rewartT, rewardTransactionPOW)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	mainT := &transaction.Transaction{}
	err = mainT.BytesToTransaction(bytesMainTransaction)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	mainPOWOK, err := controller.cheker.Check(mainT, mainTransactionPOW)
	if err != nil {
		return fmt.Errorf("The work cannot be accepted: %v", err)
	}

	if mainPOWOK && rewardPOWOK {
		err = controller.mediator.AddBlock(bytesRewardTransaction, rewardTransactionPOW)
		if err != nil {
			return fmt.Errorf("The work cannot be accepted: %v", err)
		}

		err = controller.mediator.AddBlock(bytesMainTransaction, mainTransactionPOW)
		if err != nil {
			return fmt.Errorf("The work cannot be accepted: %v", err)
		}
	}
	return fmt.Errorf("The work cannot be accepted: %v", err)
}
