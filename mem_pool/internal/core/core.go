package core

import (
	"log"
	grpcclient "mem_pool/internal/adapters/transport/client/grpc_client"
	"mem_pool/internal/transaction"
)

type transactionOutputPool interface {
	/* Функция пробует заблокировать выход */
	BlockOutput(output transaction.TransactionOutput) error
	/* Добавляет новык выходы в пулл */
	AddOutputs(outputs []transaction.TransactionOutput) error
	/* Возвращает список всех транзакций с незаблокированными выходами */
	GetAllUnlockOutputs() ([]transaction.TransactionOutput, error)
}

// type addOutputs interface {
// 	AddOutputs(outs []*transaction.TransactionOutput) error
// }

// type removeOutputs interface {
// 	RemoveOutputs(outs []*transaction.TransactionOutput) error
// }

type addTransaction interface {
	AddTransaction(trn transaction.Transaction) error
}

type getBalance interface {
	GetBalance(address []byte) (int32, error)
}

// type getFreeTransactionsOutputs interface {
// 	GetFreeTransactionsOutputs() ([]*transaction.TransactionOutput, error)
// }

type hashCalculator interface {
	HashCalculate(data []byte) []byte
}

type Core struct {
	//addOutputs
	//removeOutputs
	transactionOutputPool
	addTransaction
	getBalance
	//getFreeTransactionsOutputs
	hashCalculator
	grpc_client *grpcclient.Client
}

func NewCore(pool transactionOutputPool, at addTransaction, gb getBalance, hc hashCalculator, gc *grpcclient.Client) *Core {
	return &Core{
		transactionOutputPool: pool,
		addTransaction: at,
		getBalance: gb,
		hashCalculator: hc,
		grpc_client: gc,
	}
}

func (core *Core) Init() error {
	return nil
}

func (core *Core) CreateCoinTransferTransaction(amount int, recipientAddress, senderAddress []byte) (error) {
	tr, err := transaction.NewTransferTransaction(amount, recipientAddress, senderAddress, core.hashCalculator, core.transactionOutputPool)
	if err != nil {
		log.Printf("<core.go> Не удалось выполнить операцию по созданию транзакции перевода средств: %v", err)
		return err
	}

	err = core.addTransaction.AddTransaction(*tr)
	if err != nil {
		log.Printf("<core.go> Не удалось выполнить отправку транзакции на обработку: %v", err)
		return err
	}
	return nil
}

// func (core *Core) AddTransactionToProcessing(t transaction.Transaction) error {
// 	err := core.addTransaction.AddTransaction(t)
// 	if err != nil {
// 		log.Printf("<core.go> Не удалось выполнить отправку транзакции на обработку: %v", err)
// 		return err
// 	}
// 	return nil
// }

func (core *Core) CreateWallet(address []byte) error {
	tr, err := transaction.NewCoinbaseTransaction(10, address, address, core.hashCalculator, core.transactionOutputPool)
	if err != nil {
		log.Printf("<core.go> Не удалось выполнить операцию по созданию базисной транзакции: %v", err)
		return err
	}

	err = core.addTransaction.AddTransaction(*tr)
	if err != nil {
		log.Printf("<core.go> Не удалось выполнить отправку транзакции на обработку: %v", err)
		return err
	}

	log.Printf("<core.go> Кошелек (базисная транзакция) успешно создан")
	return nil
}

func (core *Core) GetWalletBalance(address []byte) (int, error) {
	res, err := core.getBalance.GetBalance(address)
	if err != nil {
		log.Printf("<core.go> Не удалось выполнить операцию подсчету баланса кошелька: %v", err)
		return -1, err
	}
	log.Printf("<core.go> Баланс кошелька успешно прочитан: %v", res)

	log.Printf("<core.go> Проверка баланса по gRPC...")
	res2, err2 := core.grpc_client.GetBalance(address)
	if err2 != nil {
		log.Printf("<core.go> Ошибка gRPC: %v", err)
		//return -1, err
	}
	log.Printf("<core.go> Баланс кошелька по gRPC: %v", res2)

	return int(res), nil
}
