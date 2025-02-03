package walletcontroller

import (
	"fmt"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
)

type balanceCalculator interface {
	GetByAddress(address []byte, iter iterator.Iterator[*block.Block]) (int, error)
}

type mediator interface {
	CreateBlocksIterator() (iterator.Iterator[*block.Block], error)
}

type WalletController struct {
	balanceCalc balanceCalculator
	mediator    mediator
}

func NewWalletController(m mediator) (*WalletController, error) {
	return &WalletController{
		mediator: m,
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
func (controller *WalletController) CreateNewWallet() error {
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
