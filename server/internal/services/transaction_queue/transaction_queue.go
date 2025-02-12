package transactionqueue

import (
	"container/list"
	"fmt"
	"golang_blockchain/internal/services/transaction"
	"sync"
)

// Queue - потокобезопасная FIFO-очередь
type TransactionQueue struct {
	mu    sync.Mutex
	items *list.List
}

// NewQueue создает новую очередь
func NewTransactionQueue() *TransactionQueue {
	return &TransactionQueue{items: list.New()}
}

// Enqueue добавляет элемент в очередь
func (q *TransactionQueue) PushTransaction(t *transaction.Transaction) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items.PushBack(t)
	return nil
}

// Dequeue удаляет и возвращает первый элемент очереди
func (q *TransactionQueue) PullTransaction() (*transaction.Transaction, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	elem := q.items.Front()
	if elem == nil {
		return nil, fmt.Errorf("Can not get element from queue, queue is empty")
	}
	q.items.Remove(elem)
	return elem.Value.(*transaction.Transaction), nil
}

// Len возвращает количество элементов в очереди
func (q *TransactionQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.items.Len()
}


