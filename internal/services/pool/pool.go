package pool

import (
	"golang_blockchain/internal/services/transaction"
	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
	"sync"
	"fmt"
)

type OutputsPool struct {
	Map map[string]*transaction.TransactionOutput
	Mu  sync.Mutex
}

func NewOutputsPool(blockchain iterator.Iterator[*block.Block]) (*OutputsPool, error) {
	var pool OutputsPool
	pool.Map = make(map[string]*transaction.TransactionOutput)
	pool.Mu = sync.Mutex{}
	inputs := make(map[string]interface{})

	for ok, _ := blockchain.HasNext(); ok; ok, _ = blockchain.HasNext() {
		currentValue, err := blockchain.Current()
		if err != nil {
			return nil, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		// Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		transactionBytes := currentValue.Data
		tran := &transaction.Transaction{}
		err = tran.BytesToTransaction(transactionBytes)
		if err != nil {
			return nil, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}


		// Запоминаем все входы
		for _, input := range tran.Inputs {
			inputs[string(input.PreviousOutputHash)] = 1
		}

		// Обходим выходы транзакции запоминая все выходы
		for _, output := range tran.Outputs {
			// Если хэш выхода не используется входом, значит добавляем в словарь
			if _, ok := inputs[string(output.Hash)]; !ok {
				pool.Map[string(output.Hash)] = &output
			} else {
				delete(pool.Map, string(output.Hash))
			}
		}

		blockchain.Next()
	}

	return &pool, nil
}

func (p *OutputsPool) BlockOutput(output transaction.TransactionOutput) (bool, error) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	if _, ok := p.Map[transaction.TransactionOutputToString(output)]; ok {
		delete(p.Map, transaction.TransactionOutputToString(output))
		return true, nil
	}
	return false, nil
}

func (p *OutputsPool) AddOutputs(outputs []transaction.TransactionOutput) error {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	for _, output := range outputs {
		p.Map[transaction.TransactionOutputToString(output)] = &output
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