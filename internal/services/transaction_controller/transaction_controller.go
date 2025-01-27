package transactioncontroller

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"log"

	"golang_blockchain/internal/services/pool"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/boltdb"
	proofofwork "golang_blockchain/pkg/proof_of_work"
)

/*
Контроллер странзакций. Для удобства
используем встраивание блокчейна
*/
type TransactionController struct {
	OutputsPool transaction.TransactionOutputPool
	blockchain.Blockchain
}

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

	transactionController.Blockchain = *blockchain

	// Инициализация пулла свободных выходов
	transactionController.OutputsPool = pool.NewPool[transaction.TransactionOutput](transaction.TransactionOutputToString)

	log.Printf("Контроллер транзакций успешно запущен!")

	return &transactionController, nil
}

func (controller *TransactionController) CreateNewCoinBase(reward int, address, key []byte) error {
	// Создание транзакции
	t, err := transaction.NewCoinbaseTransaction(reward, address, key, controller.HashCalc, controller.OutputsPool)
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	err = controller.AddBlockToBlockchain(data)
	if err != nil {
		return fmt.Errorf("Coinbase transaction was failed: %v", err)
	}

	return nil
}

func (controller *TransactionController) CreateCoinTransfer(amount int, recipientAddress, senderAddress []byte) error {
	// Итератор по блокчейну для поиска выходов
	iter, err := controller.CreateIterator()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание транзакции
	t, err := transaction.NewTransferTransaction(
		amount, recipientAddress, senderAddress, 
		iter, controller.HashCalc, controller.OutputsPool,
	)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	// Создание блока из транзакции и лобавление блока в блокчейн
	data, err := t.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	err = controller.AddBlockToBlockchain(data)
	if err != nil {
		return fmt.Errorf("Transfer transaction was failed: %v", err)
	}

	return nil
}