package transactioncontroller

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"

	//"golang_blockchain/internal/services/pool"
	//"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/boltdb"
	proofofwork "golang_blockchain/pkg/proof_of_work"
)

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
	//transactionController.OutputsPool = pool.NewPool()
}
