package pool

import (
	"sync"
)

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
	delete(p.Map, p.KeyFunc(output))
	p.Mu.Unlock()
	return true, nil
}

func (p *Pool[T]) AddOutputs(outputs []T) error {
	p.Mu.Lock()
	for _, output := range outputs {
		p.Map[p.KeyFunc(output)] = &output
	}
	p.Mu.Unlock()
	return nil
}