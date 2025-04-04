package core

import (
	"mem_pool/internal/transaction"
)

type addOutputs interface {
	AddOutputs(outs []*transaction.TransactionOutput) error
}

type removeOutputs interface {
	RemoveOutputs(outs []*transaction.TransactionOutput) error
}

type addTransaction interface {
	AddTransaction(trn transaction.Transaction) error
}

type Core struct {
	addOutputs
	removeOutputs
	addTransaction
}

func NewCore(ao addOutputs, ro removeOutputs, ar addTransaction) *Core {
	return &Core{
		addOutputs: ao,
		removeOutputs: ro,
		addTransaction: ar,
	}
}

func (core *Core) Init() error {
	return nil
}

func (core *Core) CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (*transaction.Transaction, error) {
	return nil, nil
}

func (core *Core) AddTransactionToProcessing(t *transaction.Transaction) error {
	return nil
}

func (core *Core) CreateWallet(address []byte) error {
	return nil
}

func (core *Core) GetWalletBalance(address []byte) (int, error) {
	return -1, nil
}
