package balancecalculator

import (
	//"bytes"
	"bytes"
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
	"log"
)

type BalanceCalculator struct {
}

func NewBalanceCalculator() (*BalanceCalculator, error) {
	return &BalanceCalculator{}, nil
}

func (bc *BalanceCalculator) GetByAddress(address []byte, iter iterator.Iterator[*block.Block]) (int, error) {
	outputs := make(map[string]*transaction.TransactionOutput)
	inputs := make(map[string]interface{})

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentBlock, err := iter.Current()
		if err != nil {
			return 0, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		log.Printf("Текущий блок: HASH = %x", currentBlock.Hash)

		// Расшифровываем информацию блока, извлекаем список транзакций
		transactions, err := transaction.DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return 0, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}

		// Определяем входы входящие в блок и выходы выходящие из блока
		ins := make(map[string]interface{})
		outs := make(map[string]*transaction.TransactionOutput)
		for _, tx := range transactions {
			for _, out := range tx.Outputs {
				outs[string(out.Hash)] = &out
			}

			for _, in := range tx.Inputs {
				if _, ok := outs[string(in.PreviousOutputHash)]; ok {
					delete(outs, string(in.PreviousOutputHash))
				} else {
					ins[string(in.PreviousOutputHash)] = 0
				}
			}
		}

		//log.Printf("Входы блока: %v", ins)
		//log.Printf("Выходы блока: %v", outs)

		// Запоминаем все входы
		for hash, _ := range ins {
			log.Printf("Вход: HASH = %x", hash)
			inputs[hash] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for hash, out := range outs {
			log.Printf("Выход: HASH = %x, получатель = %v, значение = %v", hash, string(out.RecipientAddress), out.Value)
			// Если хэш выхода не используется входом, значит добавляем в словарь
			if _, ok := inputs[hash]; !ok {
				outputs[hash] = out
			} else {
				delete(outputs, hash)
			}
		}

		iter.Next()
	}

	// Подсчитываем все значения
	res := 0
	for _, val := range outputs {
		if bytes.Equal(val.RecipientAddress, address) {
			res += val.Value
		}
	}

	return res, nil
}
