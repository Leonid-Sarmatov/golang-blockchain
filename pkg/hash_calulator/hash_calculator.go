package hashcalulator

/*
HashCalculate описывает интерфейс для различных
вариантов хеш-калькуляторов
*/
type HashCalculator interface {
	HashCalculate(data []byte) []byte
}