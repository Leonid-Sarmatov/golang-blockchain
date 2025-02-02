package mediator

import (
	blockchaincontroller "golang_blockchain/internal/controllers/blockchain_controller"
	transactioncontroller "golang_blockchain/internal/controllers/transaction_controller"
	walletcontroller "golang_blockchain/internal/controllers/wallet_controller"
)

type Mediator struct {
	blockchaincontroller *blockchaincontroller.BlockchainController
	transactioncontroller *transactioncontroller.TransactionController
	walletcontroller *walletcontroller.WalletController
}

func NewMediator() (*Mediator, error) {
	return nil, nil
}