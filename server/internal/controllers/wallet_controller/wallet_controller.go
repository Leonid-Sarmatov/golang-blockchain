package walletcontroller

import (
	"fmt"
	balancecalculator "golang_blockchain/internal/services/balance_calculator"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
	"log"
)

type balanceCalculator interface {
	GetByAddress(address []byte, iter iterator.Iterator[*block.Block]) (int, error)
}

type mediator interface {
	CreateBlocksIterator() (iterator.Iterator[*block.Block], error)
	CreateNewCoinBaseTransaction(reward int, address, key []byte) (*transaction.Transaction, error)
	AddBlock(block *block.Block, pwValue int) error
	CreateBlock(data []byte) (*block.Block, error)
}

type WalletController struct {
	balanceCalc balanceCalculator
	mediator    mediator
}

func NewWalletController(m mediator) (*WalletController, error) {
	bc, err := balancecalculator.NewBalanceCalculator()
	if err != nil {
		return nil, fmt.Errorf("Start wallet controller was failed: %v", err)
	}

	log.Printf("Контроллер кошельков успешно запущен!")
	
	return &WalletController{
		mediator: m,
		balanceCalc: bc,
	}, nil
}

/*
AddBlock добавляет новый блок, и проверяет proof-of-work

Аргументы:
  - []byte: data данные блока (полезная нагрузка в виде транзакции)
  - int: pwValue доказательство работы

Возвращает:
  - error: ошибка
*/
func (controller *WalletController) CreateWallet(address []byte) error {
	log.Printf("Адрес создаваемого кошелька: %v", address)
	transaction, err := controller.mediator.CreateNewCoinBaseTransaction(10, address, address)
	if err != nil {
		return fmt.Errorf("Create wallet was failed: %v", err)
	}

	byteTransaction, err := transaction.TransactionToBytes()
	if err != nil {
		return fmt.Errorf("Create wallet was failed: %v", err)
	}

	b, err := controller.mediator.CreateBlock(byteTransaction)
	if err != nil {
		return fmt.Errorf("Create wallet was failed: %v", err)
	}

	err = controller.mediator.AddBlock(b, -1)
	if err != nil {
		return fmt.Errorf("Create wallet was failed: %v", err)
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
func (controller *WalletController) GetBalanceByPublicKey(address []byte) (int, error) {
	iter, err := controller.mediator.CreateBlocksIterator()
	if err != nil {
		return -1, fmt.Errorf("Count balance was failed: %v", err)
	}

	// Подсчет балланса
	res, err := controller.balanceCalc.GetByAddress(address, iter)
	if err != nil {
		return -1, fmt.Errorf("Count balance was failed: %v", err)
	}

	return res, nil
}
