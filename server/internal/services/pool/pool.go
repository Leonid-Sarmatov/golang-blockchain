package pool

import (
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/blockchain"
	"golang_blockchain/pkg/iterator"
	"sync"
)

type OutputsPool struct {
	outputsMap map[string]*transaction.TransactionOutput
	mu         sync.Mutex
	blockchain *blockchain.Blockchain
}

func NewOutputsPool(iter iterator.Iterator[*block.Block]) (*OutputsPool, error) {
	var pool OutputsPool
	pool.outputsMap = make(map[string]*transaction.TransactionOutput)
	pool.mu = sync.Mutex{}
	inputs := make(map[string]interface{})

	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentBlock, err := iter.Current()
		if err != nil {
			return nil, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		// Расшифровываем информацию блока, извлекаем список транзакций
		transactions, err := transaction.DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return nil, fmt.Errorf("Can not convert bytes to transaction: %v", err)
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

		// Запоминаем все входы
		for hash, _ := range ins {
			inputs[hash] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for hash, out := range outs {
			// Если хэш выхода не используется входом, значит добавляем в словарь
			if _, ok := inputs[hash]; !ok {
				pool.outputsMap[hash] = out
			} else {
				delete(pool.outputsMap, hash)
			}
		}


		iter.Next()
	}

	return &pool, nil
}

func (p *OutputsPool) BlockOutput(output transaction.TransactionOutput) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.outputsMap[transaction.TransactionOutputToString(output)]; ok {
		delete(p.outputsMap, transaction.TransactionOutputToString(output))
		return true, nil
	}
	return false, nil
}

func (p *OutputsPool) AddOutputs(outputs []transaction.TransactionOutput) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, output := range outputs {
		p.outputsMap[transaction.TransactionOutputToString(output)] = &output
	}
	return nil
}

/*
	type Pool[T any] struct {
		Map map[string]interface{}
		Mu  sync.Mutex
		KeyFunc func(T) string
	}

	func NewPool[T any](keyfunc func(T) string) *Pool[T] {
		return &Pool[T]{
			Map: make(map[string]interface{}, 0),
			Mu:  sync.Mutex{},
			KeyFunc: keyfunc,
		}
	}

	func (p *Pool[T]) BlockOutput(output T) (bool, error) {
		p.Mu.Lock()
		defer p.Mu.Unlock()

		if _, ok := p.Map[p.KeyFunc(output)]; ok {
			delete(p.Map, p.KeyFunc(output))
			return true, nil
		}
		return false, nil
	}

	func (p *Pool[T]) AddOutputs(outputs []T) error {
		p.Mu.Lock()
		defer p.Mu.Unlock()
		for _, output := range outputs {
			p.Map[p.KeyFunc(output)] = &output
		}
		return nil
	}*/
