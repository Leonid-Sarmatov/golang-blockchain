package transaction

import (
	"bytes"
	"fmt"
	"log"
	"time"
	"encoding/binary"

	"node/internal/block"
	"node/internal/blockchain"
)

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
type hashCalulator interface {
	HashCalculate(data []byte) []byte
}

/* Выход транзакции */
type TransactionOutput struct {
	Value            int    // Условная "монета"
	RecipientAddress []byte // Условный "адрес" кошелька
	TimeOfCreation   int64  // Время создания
	Hash             []byte // Хэш выхода
}

/* Вход транзакции */
type TransactionInput struct {
	//PreviousTransactionID []byte // Идентификатор предыдущей транзакции
	PreviousOutputHash    []byte // Хэш выхода, к которому подключен данный вход
	PublicKey             []byte // Публичный ключ отправителя
}

/* Транзакция */
type Transaction struct {
	//ID             []byte
	TimeOfCreation int64
	Inputs         []TransactionInput
	Outputs        []TransactionOutput
}

/* Конструктор для выхода транзакции */
func NewTransactionOutput(value int, address []byte, hc hashCalulator) (TransactionOutput, error) {
	output := TransactionOutput{
		Value:            value,
		RecipientAddress: address,
		TimeOfCreation:   time.Now().Unix(),
	}

	bytesOut, err := SerializeTransactionOutput(&output)
	if err != nil {
		return output, fmt.Errorf("Convert transction output to byte slice was failed: %v\n", err)
	}

	output.Hash = hc.HashCalculate(bytesOut)
	return output, nil
}

/*
TransactionOutputPool описывает интерфейс для
пулла доступных выходов транзакций
*/
type TransactionOutputPool interface {
	/* Функция пробует заблокировать выход, false - выход заблокирован ранее, true - успешно заблокирован */
	BlockOutput(output TransactionOutput) (bool, error)
	/* Добавляет новых выход в пулл */
	AddOutputs(outputs []TransactionOutput) error
}

/* Базисная транзакция с пустыми входами */
func NewCoinbaseTransaction(reward int, address, key []byte, hc hashCalulator, pool TransactionOutputPool) (*Transaction, error) {
	input := TransactionInput{
		//PreviousTransactionID: []byte{},
		PreviousOutputHash:    []byte{},
		PublicKey:             key,
	}

	output, err := NewTransactionOutput(reward, address, hc)
	if err != nil {
		return nil, fmt.Errorf("Can not create basis output: %v", err)
	}

	transaction := &Transaction{
		//ID:             []byte{},
		TimeOfCreation: time.Now().Unix(),
		Inputs:         []TransactionInput{input},
		Outputs:        []TransactionOutput{output},
	}

	// bytes, err := transaction.TransactionToBytes()
	// if err != nil {
	// 	return nil, fmt.Errorf("Can not convert transaction to bytes: %v", err)
	// }

	// hash := hc.HashCalculate(bytes)
	// transaction.ID = hash
	err = pool.AddOutputs([]TransactionOutput{output})
	if err != nil {
		return nil, fmt.Errorf("Can not add output to pool: %v", err)
	}

	log.Printf("Новый коин-базис успешно создан! Адрес: %v, Балланс: %v\n", output.RecipientAddress, output.Value)

	return transaction, nil
}

/* Обычная транзакция с переводом коинов */
func NewTransferTransaction(
	amount int, recipientAddress, senderAddress []byte,
	iter blockchain.Iterator[*block.Block], hc hashCalulator, pool TransactionOutputPool,
) (*Transaction, error) {

	// Входы транзакции и суммарный счет
	inputs := make([]TransactionInput, 0)
	totalInputValue := 0

Metka:
	for ok, _ := iter.HasNext(); ok; ok, _ = iter.HasNext() {
		currentBlock, err := iter.Current()
		if err != nil {
			return nil, fmt.Errorf("Searching transaction was failed: %v", err)
		}

		log.Printf("Хэш блока: %v\n Хэш предыдущего блока: %v\n\n", currentBlock.Hash, currentBlock.PrevBlockHash)

		// // Расшифровываем информацию блока, то есть содержащуюся в нем транзакцию
		// transactionBytes := currentValue.Data
		// transaction := &Transaction{}
		// err = transaction.BytesToTransaction(transactionBytes)
		// if err != nil {
		// 	return nil, fmt.Errorf("Can not convert bytes to transaction: %v", err)
		// }

		// Расшифровываем информацию блока, извлекаем список транзакций
		transactions, err := DeserializeTransactions(currentBlock.Data)
		if err != nil {
			return nil, fmt.Errorf("Can not convert bytes to transactions: %v", err)
		}

		// Обходим транзакции из блока
		for _, transaction := range transactions {
			// Обходим выходы транзакции аккумулируя выходы и баланс отправителя
			for _, output := range transaction.Outputs {
				// fmt.Printf(
				// 	"Адрес пользователя: %v, Баланс пользоватя (Для рассматриваемого выхода транзакции) = %v\n",
				// 	string(output.RecipientAddress), output.Value,
				// )

				if bytes.Equal(output.RecipientAddress, senderAddress) {
					// Проверка доступности выхода
					ok, _ := pool.BlockOutput(output)
					if !ok {
						continue
					}

					totalInputValue += output.Value
					input := TransactionInput{
						//PreviousTransactionID: transaction.ID,
						PreviousOutputHash:    output.Hash,
						PublicKey:             senderAddress,
					}

					inputs = append(inputs, input)
				}

				if totalInputValue >= amount {
					break Metka
				}
			}
		}
		// Переход к следующему блоку в блокчейне
		iter.Next()
	}

	// Проверка накопленного баланса
	if totalInputValue < amount {
		return nil, fmt.Errorf("Insufficient funds on balance")
	}

	outputs := make([]TransactionOutput, 0, 2)

	output1, err := NewTransactionOutput(amount, recipientAddress, hc)
	if err != nil {
		return nil, fmt.Errorf("Output create error: %v", err)
	}
	outputs = append(outputs, output1)
	log.Printf("Пользователь адреса %v получает перевод %v\n", recipientAddress, amount)

	// Если отправителю нужна сдача то добавляем  выход со сдачей
	output2, err := NewTransactionOutput(totalInputValue-amount, senderAddress, hc)
	if err != nil {
		return nil, fmt.Errorf("Output create error: %v", err)
	}
	outputs = append(outputs, output2)
	log.Printf("Пользователь адреса %v получает сдачу %v\n", senderAddress, totalInputValue-amount)

	// Создание структуры транзакции и подсчет хэша
	transaction := &Transaction{
		//ID:             []byte{},
		TimeOfCreation: time.Now().Unix(),
		Inputs:         inputs,
		Outputs:        outputs,
	}
	// bytes, err := transaction.TransactionToBytes()
	// if err != nil {
	// 	return nil, fmt.Errorf("Can not convert transaction to bytes: %v", err)
	// }
	// hash := hc.HashCalculate(bytes)
	// transaction.ID = hash

	// Добавление в пулл новых выходов
	pool.AddOutputs(outputs)

	return transaction, nil
}

/* Перевод выхода в строку */
func TransactionOutputToString(to TransactionOutput) string {
	return string(to.Hash)
}

/*
=======================================================
========= Функции сериализации/десериализации =========
=======================================================
*/

/*
SerializeTransactionOutput сериализует выход транзакции в байтовый слайс

Аргументы:
  - *TransactionOutput: output выход транзакции

Возвращает:
  - error: ошибка
  - []byte: готовый байтовый слайс с транзакциями
*/
func SerializeTransactionOutput(output *TransactionOutput) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Запись количества коинов
	if err := binary.Write(buf, binary.LittleEndian, int32(output.Value)); err != nil {
		return nil, err
	}

	// Запись длины адреса кошелька
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(output.RecipientAddress))); err != nil {
		return nil, err
	}
	// Запись самого адреса кошелька
	if err := writeBytes(buf, output.RecipientAddress); err != nil {
		return nil, err
	}

	// Запись времени создания
	if err := binary.Write(buf, binary.LittleEndian, output.TimeOfCreation); err != nil {
		return nil, err
	}

	// Запись длины хэша
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(output.Hash))); err != nil {
		return nil, err
	}
	// Запись самого хэша
	if err := writeBytes(buf, output.Hash); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

/*
SerializeTransaction сериализует слайс транзакций в байтовый слайс

Аргументы:
  - Transaction: transaction транзакция для записи

Возвращает:
  - []byte: сериализованная транзакция
  - error: ошибка
*/
func SerializeTransaction(buf *bytes.Buffer, tx Transaction) error {

	// Запись времени создания
	if err := binary.Write(buf, binary.LittleEndian, tx.TimeOfCreation); err != nil {
		return err
	}

	// Запись количества входов
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(tx.Inputs))); err != nil {
		return err
	}
	// Запись самих входов
	for _, in := range tx.Inputs {
		if err := writeBytes(buf, in.PreviousOutputHash); err != nil {
			return err
		}
		if err := writeBytes(buf, in.PublicKey); err != nil {
			return err
		}
	}

	// Запись количества выходов
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(tx.Outputs))); err != nil {
		return err
	}
	// Запись самих выходов
	for _, out := range tx.Outputs {
		// Записываем Value как int32.
		if err := binary.Write(buf, binary.LittleEndian, int32(out.Value)); err != nil {
			return err
		}
		if err := writeBytes(buf, out.RecipientAddress); err != nil {
			return err
		}
		if err := binary.Write(buf, binary.LittleEndian, out.TimeOfCreation); err != nil {
			return err
		}
		if err := writeBytes(buf, out.Hash); err != nil {
			return err
		}
	}

	return nil
}

/*
DeserializeTransaction восстанавливает транзакцию из байтового слайса

Аргументы:
  - []byte: байтовое представление транзакции

Возвращает:
  - error: ошибка
  - []Transaction: список транзакций
*/
func DeserializeTransaction(buf *bytes.Reader) (Transaction, error) {
	var tx Transaction

	if err := binary.Read(buf, binary.LittleEndian, &tx.TimeOfCreation); err != nil {
		return tx, err
	}

	// Чтение количества входов
	var inCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &inCount); err != nil {
		return tx, err
	}
	// Чтение самих входов 
	tx.Inputs = make([]TransactionInput, inCount)
	for j := uint32(0); j < inCount; j++ {
		var in TransactionInput

		prevOutHash, err := readBytes(buf)
		if err != nil {
			return tx, err
		}
		in.PreviousOutputHash = prevOutHash

		pubKey, err := readBytes(buf)
		if err != nil {
			return tx, err
		}
		in.PublicKey = pubKey

		tx.Inputs[j] = in
	}

	// Чтение количества выходов
	var outCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &outCount); err != nil {
		return tx, err
	}
	// Чтение самих выходов
	tx.Outputs = make([]TransactionOutput, outCount)
	for k := uint32(0); k < outCount; k++ {
		var out TransactionOutput

		var value int32
		if err := binary.Read(buf, binary.LittleEndian, &value); err != nil {
			return tx, err
		}
		out.Value = int(value)

		recip, err := readBytes(buf)
		if err != nil {
			return tx, err
		}
		out.RecipientAddress = recip

		if err := binary.Read(buf, binary.LittleEndian, &out.TimeOfCreation); err != nil {
			return tx, err
		}

		hash, err := readBytes(buf)
		if err != nil {
			return tx, err
		}
		out.Hash = hash

		tx.Outputs[k] = out
	}

	return tx, nil
}

/*
SerializeTransactions сериализует слайс транзакций в байтовый слайс

Аргументы:
  - []Transaction: transactions список транзакций для записи

Возвращает:
  - []byte: готовый байтовый слайс с транзакциями
  - error: ошибка
*/
func SerializeTransactions(transactions []*Transaction) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Запись количества транзакций
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(transactions))); err != nil {
		return nil, err
	}
	// Запись каждой транзакции 
	for _, tx := range transactions {
		err := SerializeTransaction(buf, *tx)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

/*
DeserializeTransactions восстанавливает слайс транзакций из байтового слайса

Аргументы:
  - []byte: data слайс с закодированными транзакциями

Возвращает:
  - []Transaction: список транзакций
  - error: ошибка
*/
func DeserializeTransactions(data []byte) ([]*Transaction, error) {
	// Ридер для вычитывания сериализованных данных
	buf := bytes.NewReader(data)

	// Выитывание количества транзакций в сериализованных данных
	var txCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &txCount); err != nil {
		return nil, err
	}

	// Создание слайса транзакций и поочередное вычитывание их из ридера
	txs := make([]*Transaction, txCount)
	for i := uint32(0); i < txCount; i++ {
		tx, err := DeserializeTransaction(buf)
		if err != nil {
			return nil, err
		}
		txs[i] = &tx
	}
	return txs, nil
}

/*
writeBytes записывает слайс байтов в буфер, 
предварительно записывая его длину (uint32)

Аргументы:
  - *bytes.Buffer: buf указатель на буфер для записи
  - []byte: data данные для записи в буфер

Возвращает:
  - error: ошибка
*/
func writeBytes(buf *bytes.Buffer, data []byte) error {
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	if _, err := buf.Write(data); err != nil {
		return err
	}
	return nil
}

/*
readBytes вычитывает из ридера размер байтового
массива и сам байтовый массив

Аргументы:
  - *bytes.Reader: buf указатель на буфер для чтения

Возвращает:
  - []byte: прочитанный слайс
  - error: ошибка
*/
func readBytes(buf *bytes.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return nil, err
	}
	data := make([]byte, length)
	n, err := buf.Read(data)
	if err != nil {
		return nil, err
	}
	if uint32(n) != length {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, n)
	}
	return data, nil
}