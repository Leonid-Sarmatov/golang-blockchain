package mediator

import (
	"fmt"
	blockchaincontroller "golang_blockchain/internal/controllers/blockchain_controller"
	minercontroller "golang_blockchain/internal/controllers/miner_controller"
	transactioncontroller "golang_blockchain/internal/controllers/transaction_controller"
	walletcontroller "golang_blockchain/internal/controllers/wallet_controller"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
)

type blockchainController interface {
	/* AddBlock добавляет новый блок, и проверяет proof-of-work */
	AddBlock(data []byte, pwValue int) error
	/* CreateIterator возвращает абстрактный итератор по блокам в блокчейне */
	CreateIterator() (iterator.Iterator[*block.Block], error)
	/* GetBlockByHash возвращает блок с заданным хэшом */
	GetBlockByHash(hash []byte) (*block.Block, error)
	/* GetAllBlocks возвращает все блоки в блокчейне */
	GetAllBlocks() ([]*block.Block, error)
}

type transactionController interface {
	/* CreateNewCoinBaseTransaction создает базисную транзакцию */
	CreateNewCoinBaseTransaction(reward int, address, key []byte) (*transaction.Transaction, error)
	/* CreateCoinTransferTransaction создает обычную транзакцию, переводит коины */
	CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (*transaction.Transaction, error)
}

type walletController interface {
	/* CreateNewWallet создает новый кошелек */
	CreateNewWallet() error
	/* GetBalanceByPublicKey подсчитывает балланс кошелька */
	GetBalanceByPublicKey(address []byte) (int, error)
}

type minerController interface {
	GetWorkForMining(rewardAddress []byte) ([]byte, []byte, error)
	SendCompletedWork(bytesRewardTransaction, bytesMainTransaction []byte, rewardTransactionPOW, mainTransactionPOW int) error
}

type Mediator struct {
	blockchaincontroller  blockchainController
	transactioncontroller transactionController
	walletcontroller      walletController
	minercontroller minerController
}

func NewMediator() (*Mediator, error) {
	var mediator Mediator

	// Загрузка контроллера блокчейна
	chc, err := blockchaincontroller.NewBlockchainController()
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.blockchaincontroller = chc

	// Загрузка контроллера транзакций
	trc, err := transactioncontroller.NewTransactionController(&mediator)
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.transactioncontroller = trc

	// Загрузка контроллера кошельков
	wc, err := walletcontroller.NewWalletController(&mediator)
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.walletcontroller = wc

	// Загрузка контроллера майнеров
	mc, err := minercontroller.NewMinerController(&mediator)
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.minercontroller = mc

	return &mediator, nil
}

/*
=======================================================
=========== Вызовы к контроллеру блокчейна ============
=======================================================
*/

func (m *Mediator) AddBlock(data []byte, pwValue int) error {
	return m.blockchaincontroller.AddBlock(data, pwValue)
}

func (m *Mediator) CreateBlocksIterator() (iterator.Iterator[*block.Block], error) {
	iter, err := m.blockchaincontroller.CreateIterator()
	return iter, err
}

func (m *Mediator) GetBlockByHash(hash []byte) (*block.Block, error) {
	return nil, nil
}

func (m *Mediator) GetAllBlocks() ([]*block.Block, error) {
	return nil, nil
}

/*
=======================================================
=========== Вызовы к контроллеру транзакций ===========
=======================================================
*/

func (m *Mediator) CreateNewCoinBaseTransaction(reward int, address, key []byte) (*transaction.Transaction, error) {
	return m.transactioncontroller.CreateNewCoinBaseTransaction(reward, address, key)
}

func (m *Mediator) CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (*transaction.Transaction, error) {
	return m.transactioncontroller.CreateCoinTransferTransaction(amount, recipientAddress, senderAddress)
}

/*
=======================================================
=========== Вызовы к контроллеру кошельков ============
=======================================================
*/

func (m *Mediator) CreateNewWallet() error {
	return m.walletcontroller.CreateNewWallet()
}

func (m *Mediator) GetWalletBalance(address []byte) (int, error) {
	res, err := m.walletcontroller.GetBalanceByPublicKey(address)
	return res, err
}

/*
=======================================================
============ Вызовы к контроллеру майнеров ============
=======================================================
*/

func (m *Mediator) GetWorkForMining(rewardAddress []byte) ([]byte, []byte, error) {
	return m.minercontroller.GetWorkForMining(rewardAddress)
}

func (m *Mediator) SendCompletedWork(
	bytesRewardTransaction, bytesMainTransaction []byte, 
	rewardTransactionPOW, mainTransactionPOW int,
	) (error) {
	return m.minercontroller.SendCompletedWork(bytesRewardTransaction, bytesMainTransaction, rewardTransactionPOW, mainTransactionPOW)
}