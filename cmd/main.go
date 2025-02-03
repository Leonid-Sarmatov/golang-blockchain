package main

import (
	"fmt"

	//"golang_blockchain/internal/services/transaction"
	transactioncontroller "golang_blockchain/internal/services/transaction_controller"
	//"golang_blockchain/pkg/block"
)

func main() {
	controller, err := transactioncontroller.NewTransactionController()
	fmt.Println(err)

	//controller.CreateNewCoinBase(20, []byte("Alice"), []byte("Alice"))
	//controller.CreateNewCoinBase(20, []byte("Bob"), []byte("Bob"))
	//controller.CreateNewCoinBase(80, []byte("Minnya"), []byte("Minnya"))

	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Alice")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Bob")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Minnya")))

	//controller.CreateCoinTransfer(5, []byte("Alice"), []byte("Bob"))

	/*myIterator, err := controller.Blockchain.CreateIterator()
	fmt.Println(err)

	for ok, _ := myIterator.HasNext(); ok; ok, _ = myIterator.HasNext() {
		currentValue, err := myIterator.Current()
		if err != nil {
			break
		}
		// Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		transactionBytes := currentValue.Data
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
	}*/

	fmt.Println(controller.GetBalanceByPublicKey([]byte("Alice")))
	fmt.Println(controller.GetBalanceByPublicKey([]byte("Bob")))
	fmt.Println(controller.GetBalanceByPublicKey([]byte("Minnya")))
}
