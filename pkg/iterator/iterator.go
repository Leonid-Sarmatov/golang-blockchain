package iterator

/* Интерфейс итерируемой коллекции */
type IterableCollection interface {
	CreateIterator() Iterator
}

/* Интерфейс для любого итератора */
type Iterator interface {
	HasNext() (bool, error)
	Next() (interface{}, error)
	Current() (interface{}, error)
}
