package iterator

/* Интерфейс итерируемой коллекции */
type IterableCollection[T any] interface {
	CreateIterator() Iterator[T]
}

/* Интерфейс для любого итератора */
type Iterator[T any] interface {
	HasNext() (bool, error)
	Next() (T, error)
	Current() (T, error)
}
