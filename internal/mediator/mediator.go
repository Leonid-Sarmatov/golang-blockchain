package mediator

import (
	"fmt"
	blockchaincontroller "golang_blockchain/internal/controllers/blockchain_controller"
	transactioncontroller "golang_blockchain/internal/controllers/transaction_controller"
	walletcontroller "golang_blockchain/internal/controllers/wallet_controller"
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
	/* CreateNewCoinBase создает базисную транзакцию */
	CreateNewCoinBase(reward int, address, key []byte) error
	/* CreateCoinTransfer создает обычную транзакцию, переводит коины */
	CreateCoinTransfer(amount int, recipientAddress, senderAddress []byte) error
}

type walletController interface {
	CreateNewWallet() error
}

type Mediator struct {
	blockchaincontroller  blockchainController
	transactioncontroller transactionController
	walletcontroller      walletController
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
	wc, err := walletcontroller.NewWalletController()
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.walletcontroller = wc

	return &mediator, nil
}

/*
=======================================================
=========== Вызовы к контроллеру блокчейна ============
=======================================================
*/

func (m *Mediator) MediatorAddBlock(data []byte, pwValue int) error {
	return m.blockchaincontroller.AddBlock(data, pwValue)
}

func (m *Mediator) MediatorCreateIterator() (iterator.Iterator[*block.Block], error) {
	iter, err := m.blockchaincontroller.CreateIterator()
	return iter, err
}

func (m *Mediator) MediatorGetBlockByHash(hash []byte) (*block.Block, error) {
	return nil, nil
}

func (m *Mediator) MediatorGetAllBlocks() ([]*block.Block, error) {
	return nil, nil
}

/*
=======================================================
=========== Вызовы к контроллеру транзакций ===========
=======================================================
*/

func (m *Mediator) MediatorCreateNewCoinBase(reward int, address, key []byte) error {
	return m.transactioncontroller.CreateNewCoinBase(reward, address, key)
}

func (m *Mediator) CreateCoinTransfer(amount int, recipientAddress, senderAddress []byte) error {
	return m.transactioncontroller.CreateCoinTransfer(amount, recipientAddress, senderAddress)
}