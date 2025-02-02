package transactioncontroller

import (
	"fmt"
	"golang_blockchain/internal/services/balance_calculator"
	"golang_blockchain/internal/services/transaction"
	"log"

	"golang_blockchain/internal/services/pool"
	"golang_blockchain/pkg/blockchain"
)

/*
balanceCalculator описываем интерфейс 
для системы подсчета балланса пользователя
*/
type balanceCalculator interface {
	GetByAddress(address []byte) (int, error)
}

/*
Контроллер странзакций
*/
type TransactionController struct {
	outputsPool       transaction.TransactionOutputPool
	balanceCalculator balanceCalculator
	blockchain        *blockchain.Blockchain
}

/* Конструктор */
func NewTransactionController(bc *blockchain.Blockchain) (*TransactionController, error) {
	var transactionController TransactionController

	transactionController.blockchain = bc

	// Инициализация пулла свободных выходов
	pool, err := pool.NewOutputsPool(bc)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	transactionController.outputsPool = pool

	// Инициализация калькулятора балланса
	calc, err := balancecalculator.NewBalanceCalculator(bc)
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
	t, err := transaction.NewCoinbaseTransaction(reward, address, key, controller.blockchain.HashCalc, controller.outputsPool)
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	err = controller.blockchain.AddBlockToBlockchain(data, 0)
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
	// Создание транзакции
	t, err := transaction.NewTransferTransaction(
		amount, recipientAddress, senderAddress,
		controller.blockchain, controller.blockchain.HashCalc, controller.outputsPool,
	)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	err = controller.blockchain.AddBlockToBlockchain(data, 0)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	return nil
}

/*
GetBalanceByPublicKey обходит весь блокчейн с транзакциями, и считает балланс пользователя

Аргументы:
  - []byte: address публичный адрес для поиска балланса

Возвращает:
  - int: балланс кошелька
  - error: ошибка
*/
func (controller *TransactionController) GetBalanceByPublicKey(address []byte) (int, error) {
	// Подсчет балланса
	res, err := controller.balanceCalculator.GetByAddress(address)
	if err != nil {
		return -1, fmt.Errorf("Count balance was failed: %v", err)
	}

	return res, nil
}
