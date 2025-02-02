package balancecalculator

import (
	"bytes"
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/blockchain"
)

type BalanceCalculator struct{
	blockchain *blockchain.Blockchain
}

func NewBalanceCalculator(b *blockchain.Blockchain) (*BalanceCalculator, error) {
	return &BalanceCalculator{
		blockchain: b,
	}, nil
}

func (bc *BalanceCalculator)GetByAddress(address []byte) (int, error) {
	iter, err := bc.blockchain.CreateIterator()
	if err != nil {
		return -1, fmt.Errorf("Can not create iterator by blockchain: %v", err)
	}

	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentValue, err := iter.Current()
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

		iter.Next()
	}

	// Подсчитываем все значения
	res := 0
	for _, val := range outputs {
		res += val.Value
	}

	return res, nil
}
