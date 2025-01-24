package transation

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TransactionOutput struct {
	Value            int    // Условная "монета"
	RecipientAddress string // Условный "адрес" кошелька
}

type TransactionInput struct {
	PreviousTransactionID []byte // Идентификатор предыдущей транзакции
	OutputIndex           int    // Индекс выхода в предыдушей транзакции
	PublicKey             string // Публичный ключ отправителя
}

/* Транзакция */
type Transaction struct {
	ID      []byte
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

func NewCoinbaseTransaction() (*Transaction, error) {
	return nil, nil
}

func NewTransferTransaction() (*Transaction, error) {
	return nil, nil
}

/*
TransactionToBytes преобразует экземпляр
структуры блока в байтовый слайс
*/
func (b *Transaction) TransactionToBytes() ([]byte, error) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		return nil, fmt.Errorf("Convert transction to byte slice was failed: %v\n", err)
	}

	return result.Bytes(), nil
}

/*
BytesToTransaction парсит бинарное представление
блока в структуру

	clice - бинарные данные
*/
func (b *Transaction) BytesToTransaction(clice []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(clice))
	return decoder.Decode(b)
}
