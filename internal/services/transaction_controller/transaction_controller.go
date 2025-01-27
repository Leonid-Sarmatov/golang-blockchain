package transactioncontroller

import (
	"fmt"
	"golang_blockchain/internal/services/balance_calculator"
	"golang_blockchain/internal/services/transaction"
	"log"

	"golang_blockchain/internal/services/pool"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/boltdb"
	proofofwork "golang_blockchain/pkg/proof_of_work"
)

/*
Контроллер странзакций
*/
type TransactionController struct {
	outputsPool       transaction.TransactionOutputPool
	blockchain        *blockchain.Blockchain
	balanceCalculator *balancecalculator.BalanceCalculator
}

/* Конструктор */
func NewTransactionController() (*TransactionController, error) {
	var transactionController TransactionController

	// Хранилище блокчейна (база данных)
	storage := boltdb.NewBBoltDBDriver()

	// Механизм проверки работы (он же и хешь-калькулятор)
	pwork := proofofwork.NewProofOfWork()

	// Инициализация блокчейна
	blockchain, err := blockchain.NewBlockchain(storage, pwork, pwork)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	transactionController.blockchain = blockchain

	// Инициализация пулла свободных выходов
	iter, err := blockchain.CreateIterator()
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	p, err := pool.NewOutputsPool(iter)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	transactionController.outputsPool = p

	// Инициализация калькулятора балланса
	transactionController.balanceCalculator = balancecalculator.NewBalanceCalculator()

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

	err = controller.blockchain.AddBlockToBlockchain(data)
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
	// Итератор по блокчейну для поиска выходов
	iter, err := controller.blockchain.CreateIterator()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание транзакции
	t, err := transaction.NewTransferTransaction(
		amount, recipientAddress, senderAddress,
		iter, controller.blockchain.HashCalc, controller.outputsPool,
	)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	err = controller.blockchain.AddBlockToBlockchain(data)
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
	// Создание итератора по блокчейну
	iter, err := controller.blockchain.CreateIterator()
	if err != nil {
		return -1, fmt.Errorf("Count balance was failed: %v", err)
	}

	// Подсчет балланса
	res, err := controller.balanceCalculator.GetByAddress(address, iter)
	if err != nil {
		return -1, fmt.Errorf("Count balance was failed: %v", err)
	}

	return res, nil
}
