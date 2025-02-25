package mediator

import (
	"fmt"
    "golang_blockchain/internal/core/blockchain"
	"golang_blockchain/internal/core/mining"
	"golang_blockchain/internal/core/transactions"
	"golang_blockchain/internal/core/wallets"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
)

type blockchainController interface {
	/* AddBlock добавляет подтверждение работы и записывает блок в блокчейн */
	AddBlock(b *block.Block, pwValue int) error
	/* CreateBlock создает новый блок и добавляет в него данные*/
	CreateBlock(data []byte) (*block.Block, error)
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
	CreateWallet(address []byte) error
	/* GetBalanceByPublicKey подсчитывает балланс кошелька */
	GetBalanceByPublicKey(address []byte) (int, error)
}

type minerController interface {
	/* AddTransactionToProcessing отправляет перевод средств в очередь обработки */
	AddTransactionToProcessing(t *transaction.Transaction) error
	/* GetWorkForMining выдает работу для майнера */
	GetWorkForMining(rewardAddress []byte) ([]byte, error)
	/* SendCompletedWork принимает работу на проверку */
	SendCompletedWork(bytesTransaction []byte, POW int) error
}

type Mediator struct {
	blockchaincontroller  blockchainController
	transactioncontroller transactionController
	walletcontroller      walletController
	minercontroller       minerController
}

func NewMediator() (*Mediator, error) {
	var mediator Mediator

	// Загрузка контроллера блокчейна
	chc, err := blockchain.NewBlockchainController()
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.blockchaincontroller = chc

	// Загрузка контроллера транзакций
	trc, err := transactions.NewTransactionController(&mediator)
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.transactioncontroller = trc

	// Загрузка контроллера кошельков
	wc, err := wallets.NewWalletController(&mediator)
	if err != nil {
		return nil, fmt.Errorf("Mediator spawn was failed: %v", err)
	}
	mediator.walletcontroller = wc

	// Загрузка контроллера майнеров
	mc, err := mining.NewMinerController(&mediator)
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

func (m *Mediator) AddBlock(block *block.Block, pwValue int) error {
	return m.blockchaincontroller.AddBlock(block, pwValue)
}

func (m *Mediator) CreateBlock(data []byte) (*block.Block, error) {
	return m.blockchaincontroller.CreateBlock(data)
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

func (m *Mediator) CreateWallet(address []byte) error {
	return m.walletcontroller.CreateWallet(address)
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
func (m *Mediator) AddTransactionToProcessing(t *transaction.Transaction) error {
	return m.minercontroller.AddTransactionToProcessing(t)
}

func (m *Mediator) GetWorkForMining(rewardAddress []byte) ([]byte, error) {
	return m.minercontroller.GetWorkForMining(rewardAddress)
}

func (m *Mediator) SendCompletedWork(bytesTransaction []byte, POW int,
) error {
	return m.minercontroller.SendCompletedWork(bytesTransaction, POW)
}
