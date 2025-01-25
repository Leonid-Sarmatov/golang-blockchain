package transaction

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"sync"
	"time"
	"strconv"

	"golang_blockchain/pkg/block"
	"golang_blockchain/pkg/iterator"
)

type TransactionOutput struct {
	Value            int    // Условная "монета"
	RecipientAddress []byte // Условный "адрес" кошелька
}

type TransactionInput struct {
	PreviousTransactionID []byte // Идентификатор предыдущей транзакции
	OutputIndex           int    // Индекс выхода в предыдушей транзакции
	PublicKey             []byte // Публичный ключ отправителя
}

/* Транзакция */
type Transaction struct {
	ID      []byte
	TimeOfCreation int64
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

/*
HashCalculate описывает интерфейс для различных
вариантов хеш-калькуляторов
*/
type HashCalculator interface {
	HashCalculate(data []byte) []byte
}

/*
HashCalculate описывает интерфейс для
пулла доступных выходов транзакций
*/
type TransactionOutputPool interface {
	/* Функция пробует заблокировать выход, false - выход заблокирован ранее, true - успешно заблокирован */
	BlockOutput(output *TransactionOutput) (bool, error)
	/* Добавляет новых выход в пулл */
	AddOutput(output *TransactionOutput) error 
}

type Pool struct {
	Map map[string]*TransactionOutput
	Mu sync.Mutex
}

func NewPool() *Pool {
	return &Pool{
		Map: make(map[string]*TransactionOutput, 0),
		Mu: sync.Mutex{},
	}
}

func (p *Pool) BlockOutput(output *TransactionOutput) (bool, error) {
	delete(p.Map, strconv.Itoa(output.Value)+string(output.RecipientAddress))
	return true, nil
}

func (p *Pool) AddOutput(output *TransactionOutput) error {
	p.Map[strconv.Itoa(output.Value)+string(output.RecipientAddress)] = output
	return nil
}


func NewCoinbaseTransaction(
		reward int, address, key []byte, 
		hc HashCalculator, pool TransactionOutputPool,
		) (*Transaction, error) {
	input := TransactionInput{
		PreviousTransactionID: []byte{},
		OutputIndex: -1,
		PublicKey: key,
	}

	output := TransactionOutput{
		Value: reward,
		RecipientAddress: address,
	}

	transaction := &Transaction{
		ID: []byte{},
		TimeOfCreation:   time.Now().Unix(),
		Inputs: []TransactionInput{input},
		Outputs: []TransactionOutput{output},
	}

	bytes, err := transaction.TransactionToBytes()
	if err != nil {
		return nil, fmt.Errorf("Can not convert transaction to bytes: %v", err)
	}

	hash := hc.HashCalculate(bytes)
	transaction.ID = hash
	pool.AddOutput(&output)

	log.Printf("Новый коин-базис успешно создан! Адрес пользователя: %v, Балланс пользователя\n", string(address))

	return transaction, nil
}

func NewTransferTransaction(
		amount int, recipientAddress, senderAddress []byte, 
		blockchain iterator.Iterator, hc HashCalculator, pool TransactionOutputPool,
	) (*Transaction, error) {

	// Входы транзакции и суммарный счет
	inputs := make([]TransactionInput, 0)
	totalInputValue := 0

	Metka:
	for ok, _ := blockchain.HasNext(); ok; ok, _ = blockchain.HasNext() {
		currentValue, err := blockchain.Current()
		if err != nil {
			return nil, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		fmt.Printf("Информация блока: %v\n", string(currentValue.(*block.Block).Data))
		fmt.Printf("Хэш блока: %v\n", currentValue.(*block.Block).Hash)
		fmt.Printf("Хэш предыдущего блока: %v\n", currentValue.(*block.Block).PrevBlockHash)

		// Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		transactionBytes := currentValue.(*block.Block).Data
		transaction := &Transaction{}
		err = transaction.BytesToTransaction(transactionBytes)
		if err != nil {
			return nil, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		}

		// Обходим выходы транзакции аккумулируя выходы и баланс отправителя
		for index, output := range transaction.Outputs {
			fmt.Printf(
				"Адрес пользователя: %v, Баланс пользоватя (Для рассматриваемого выхода транзакции) = %v\n",
				string(output.RecipientAddress), output.Value,
			)
			
			if bytes.Equal(output.RecipientAddress, senderAddress) {
				// Проверка доступности выхода
				ok, err := pool.BlockOutput(&output)
				if !ok || err != nil {
					continue
				}
				
				totalInputValue += output.Value
				inputs = append(inputs, TransactionInput{
					PreviousTransactionID: transaction.ID,
					OutputIndex: index,
					PublicKey: senderAddress,
				})
			}

			if totalInputValue >= amount {
				break Metka
			} 
		}
		// Переход к следующему блоку в блокчейне
		blockchain.Next()
	}

	// Проверка накопленного баланса
	if totalInputValue < amount {
        return nil, fmt.Errorf("Insufficient funds on balance")
    }

	// Создаем выход для получателя
	outputs := []TransactionOutput{
		{
			Value: amount,
			RecipientAddress: recipientAddress,
		},
	}
	pool.AddOutput(&outputs[0])
	log.Printf("Пользователь адреса %v получает сдачу %v\n", senderAddress, amount)

	// Если отправителю нужна сдача то добавляем  выход со сдачей
	outputs = append(outputs, TransactionOutput{
		Value: totalInputValue - amount,
		RecipientAddress: senderAddress,
	}) 
	pool.AddOutput(&outputs[1])
	log.Printf("Пользователь адреса %v получает сдачу %v\n", senderAddress, totalInputValue - amount)

	transaction := &Transaction{
		ID: []byte{},
		TimeOfCreation:   time.Now().Unix(),
		Inputs: inputs,
		Outputs: outputs,
	}

	bytes, err := transaction.TransactionToBytes()
	if err != nil {
		return nil, fmt.Errorf("Can not convert transaction to bytes: %v", err)
	}

	hash := hc.HashCalculate(bytes)
	transaction.ID = hash

	return transaction, nil
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