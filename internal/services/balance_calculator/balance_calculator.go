package balancecalculator

import (
	"bytes"
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
)

type BalanceCalculator struct{}

func NewBalanceCalculator() *BalanceCalculator {
	return &BalanceCalculator{}
}

func (bc *BalanceCalculator)GetByAddress(address []byte, blockchain iterator.Iterator[*block.Block]) (int, error) {
	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	for ok, _ := blockchain.HasNext(); ok; ok, _ = blockchain.HasNext() {
		currentValue, err := blockchain.Current()
		if err != nil {
			return -1, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		// Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		transactionBytes := currentValue.Data
		tran := &transaction.Transaction{}
		err = tran.BytesToTransaction(transactionBytes)
		if err != nil {
			return -1, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}


		// Запоминаем все входы
		for _, input := range tran.Inputs {
			inputs[string(input.PreviousOutputHash)] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for _, output := range tran.Outputs {
			if bytes.Equal(output.RecipientAddress, address) {
				// Если хэш выхода не используется входом, значит добавляем в словарь
				if _, ok := inputs[string(output.Hash)]; !ok {
					outputs[string(output.Hash)] = &output
				} else {
					delete(outputs, string(output.Hash))
				}
			}
		}

		blockchain.Next()
	}

	// Подсчитываем все значения
	res := 0
	for _, val := range outputs {
		res += val.Value
	}

	return res, nil
}
