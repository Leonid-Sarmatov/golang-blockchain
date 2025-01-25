package main

import (
	"fmt"
	block "golang_blockchain/pkg/block"
	blockchain "golang_blockchain/pkg/blockchain"
	boltdb "golang_blockchain/pkg/boltdb"
	proofofwork "golang_blockchain/pkg/proof_of_work"

	transaction "golang_blockchain/internal/services/transaction"
)

func main() {
	// Создаем проверяльщика работы
	spw := proofofwork.NewProofOfWork()

	// Создаем подключение к базе данных
	c := boltdb.NewBBoltDBDriver()

	myBlockchain, err := blockchain.NewBlockchain(c, spw, spw)
	fmt.Println(err)

	pool := transaction.NewPool()

	myIterator, err := myBlockchain.CreateIterator()
	fmt.Println(err)
	t1, err := transaction.NewCoinbaseTransaction(10, []byte("Alice"), []byte("Alice"), spw, pool)
	fmt.Println(err)
	data, err := t1.TransactionToBytes()
	fmt.Println(err)
	myBlockchain.AddBlockToBlockchain(data)

	myIterator, err = myBlockchain.CreateIterator()
	fmt.Println(err)
	t2, err := transaction.NewCoinbaseTransaction(10, []byte("Bob"), []byte("Bob"), spw, pool)
	fmt.Println(err)
	data, err = t2.TransactionToBytes()
	fmt.Println(err)
	myBlockchain.AddBlockToBlockchain(data)

	myIterator, err = myBlockchain.CreateIterator()
	fmt.Println(err)
	t3, err := transaction.NewCoinbaseTransaction(20, []byte("Minnya"), []byte("Minnya"), spw, pool)
	fmt.Println(err)
	data, err = t3.TransactionToBytes()
	fmt.Println(err)
	myBlockchain.AddBlockToBlockchain(data)

	myIterator, err = myBlockchain.CreateIterator()
	fmt.Println(err)
	t4, err := transaction.NewTransferTransaction(7, []byte("Bob"), []byte("Minnya"), myIterator, spw, pool)
	fmt.Println(err)
	data, err = t4.TransactionToBytes()
	fmt.Println(err)
	myBlockchain.AddBlockToBlockchain(data)

	myIterator, err = myBlockchain.CreateIterator()
	fmt.Println(err)

	for ok, _ := myIterator.HasNext(); ok; ok, _ = myIterator.HasNext() {
		currentValue, err := myIterator.Current()
		if err != nil {
			break
		}
		// Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		transactionBytes := currentValue.(*block.Block).Data
		transaction := &transaction.Transaction{}
		err = transaction.BytesToTransaction(transactionBytes)
		if err != nil {
			fmt.Printf("Can not convert bytes to transaction: %v", err)
		}

		// Обходим выходы транзакции аккумулируя выходы и баланс отправителя
		for _, output := range transaction.Outputs {
			fmt.Printf(
				"Адрес пользователя: %v, Баланс пользоватя (Для рассматриваемого выхода транзакции) = %v\n",
				string(output.RecipientAddress), output.Value,
			)

		}
		// Переход к следующему блоку в блокчейне
		myIterator.Next()
	}

}
