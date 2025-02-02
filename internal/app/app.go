package app

import (
	"fmt"
	"golang_blockchain/internal/controllers/transaction_controller"
	"golang_blockchain/internal/controllers/wallet_controller"

	//"golang_blockchain/internal/services/balance_calculator"
	//"golang_blockchain/internal/services/transaction"
	//"log"

	//"golang_blockchain/internal/services/pool"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/boltdb"

	proofofwork "golang_blockchain/pkg/proof_of_work"
)

type App struct {
	blockchain            *blockchain.Blockchain
	walletController      *walletcontroller.WalletController
	transactionController *transactioncontroller.TransactionController
}

func NewApp() (*App, error) {
	var app App

	// Хранилище блокчейна (база данных)
	storage := boltdb.NewBBoltDBDriver()

	// Механизм проверки работы (он же и хеш-калькулятор)
	pwork := proofofwork.NewProofOfWork()

	// Инициализация блокчейна
	b, err := blockchain.NewBlockchain(storage, pwork, pwork)
	if err != nil {
		return nil, fmt.Errorf("Start transaction controller was failed: %v", err)
	}
	app.blockchain = b

	// Инициализация контроллера транзакций
	tc, err := transactioncontroller.NewTransactionController(app.blockchain)
	if err != nil {
		return nil, fmt.Errorf("App start was failed: %v", err)
	}
	app.transactionController = tc

	return &app, nil
}
