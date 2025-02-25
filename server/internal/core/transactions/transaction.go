package transactions

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"log"

	"golang_blockchain/internal/services/pool"
	"golang_blockchain/pkg/block"

	//"golang_blockchain/pkg/blockchain"
	hashcalulator "golang_blockchain/pkg/hash_calulator"
	"golang_blockchain/pkg/iterator"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

type mediator interface {
	CreateBlocksIterator() (iterator.Iterator[*block.Block], error)
	AddBlock(b *block.Block, pwValue int) error
}

/*
Контроллер странзакций
*/
type TransactionController struct {
	outputsPool       transaction.TransactionOutputPool
	hashCalculator    hashCalulator
	mediator          mediator
}

/* Конструктор */
func NewTransactionController(m mediator) (*TransactionController, error) {
	var transactionController TransactionController
	transactionController.mediator = m

	hc := hashcalulator.NewHashCalculator()
	transactionController.hashCalculator = hc

	iter, err := m.CreateBlocksIterator()
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}

	// Инициализация пулла свободных выходов
	pool, err := pool.NewOutputsPool(iter)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	transactionController.outputsPool = pool

	log.Printf("Контроллер транзакций успешно запущен!")

	return &transactionController, nil
}

/*
CreateNewCoinBaseTransaction создает базисную транзакцию, то есть создает кошелек

Аргументы:
  - int: reward первичный балланс
  - []byte: address получатель первичного балланса
  - []byte: key публичный адрес кошелька

Возвращает:
  - error: ошибка
  - *transaction.Transaction: указатель на транзакцию
*/
func (controller *TransactionController) CreateNewCoinBaseTransaction(reward int, address, key []byte) (*transaction.Transaction, error) {
	// Создание транзакции
	t, err := transaction.NewCoinbaseTransaction(reward, address, key, controller.hashCalculator, controller.outputsPool)
	if err != nil {
		return nil, fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	return t, nil
}

/*
CreateCoinTransferTransaction создает обычную транзакцию, переводит коины

Аргументы:
  - int: amount сумма перевода
  - []byte: recipientAddress публичный адрес получателя
  - []byte: senderAddress публичный адрес отправителя

Возвращает:
  - error: ошибка
*/
func (controller *TransactionController) CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (
	*transaction.Transaction, error,
	) {
	if amount < 1 {
		return nil, fmt.Errorf("Transfer transaction was failed: amount < 1")
	}

	iter, err := controller.mediator.CreateBlocksIterator()
	if err != nil {
		return nil, fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание транзакции
	t, err := transaction.NewTransferTransaction(
		amount, recipientAddress, senderAddress,
		iter, controller.hashCalculator, controller.outputsPool,
	)
	if err != nil {
		return nil, fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	return t, nil
}

/*
CreateCoinTransfer создает обычную транзакцию, переводит коины

Аргументы:
  - int: amount сумма перевода
  - []byte: recipientAddress публичный адрес получателя
  - []byte: senderAddress публичный адрес отправителя

Возвращает:
  - error: ошибка
*/
func (controller *TransactionController) TransactionsToBytes(t []transaction.Transaction) ([]byte, error) {
	data, err := transaction.SerializeTransactions(t)
	return data, err
}
