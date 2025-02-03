package transactioncontroller

import (
	"fmt"
	"golang_blockchain/internal/services/balance_calculator"
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

/*
balanceCalculator описываем интерфейс
для системы подсчета балланса пользователя
*/
type balanceCalculator interface {
	GetByAddress(address []byte, iter iterator.Iterator[*block.Block]) (int, error)
}

type mediator interface {
	CreateBlocksIterator() (iterator.Iterator[*block.Block], error)
	AddBlock(data []byte, pwValue int) error
}

/*
Контроллер странзакций
*/
type TransactionController struct {
	outputsPool       transaction.TransactionOutputPool
	balanceCalculator balanceCalculator
	hashCalculator hashCalulator
	//blockchain        *blockchain.Blockchain
	mediator mediator
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

	// Инициализация калькулятора балланса
	calc, err := balancecalculator.NewBalanceCalculator()
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	transactionController.balanceCalculator = calc

	log.Printf("Контроллер транзакций успешно запущен!")

	return &transactionController, nil
}

/*
CreateNewCoinBase создает базисную транзакцию, то есть создает кошелек

Аргументы:
  - int: reward первичный балланс
  - []byte: address получатель первичного балланса
  - []byte: key публичный адрес кошелька

Возвращает:
  - error: ошибка
*/
func (controller *TransactionController) CreateNewCoinBase(reward int, address, key []byte) error {
	// Создание транзакции
	t, err := transaction.NewCoinbaseTransaction(reward, address, key, controller.hashCalculator, controller.outputsPool)
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	err = controller.mediator.AddBlock(data, 0)
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	return nil
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
func (controller *TransactionController) CreateCoinTransfer(
	amount int, recipientAddress, senderAddress []byte,
) error {
	iter, err := controller.mediator.CreateBlocksIterator()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание транзакции
	t, err := transaction.NewTransferTransaction(
		amount, recipientAddress, senderAddress,
		iter, controller.hashCalculator, controller.outputsPool,
	)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	err = controller.mediator.AddBlock(data, 0)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	return nil
}
