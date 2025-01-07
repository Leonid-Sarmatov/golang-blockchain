package main

import (
	"fmt"
	block "golang_blockchain/pkg/block"
	blockchain "golang_blockchain/pkg/blockchain"
	boltdb "golang_blockchain/pkg/boltdb"
	proofofwork "golang_blockchain/pkg/proof_of_work"
)

func main() {
	// Создаем проверяльщика работы
	spw := proofofwork.NewProofOfWork()

	// Создаем генезис
	genesis := block.NewGenesisBlock(spw)

	// Добавляем блок для генезиса
	block.NewBlock("Hello, blockchain!", genesis.Hash, spw)

	// Создаем подключение к базе данных
	c := boltdb.NewBBoltDBDriver()

	myBlockchain, err := blockchain.NewBlockchain(c, spw, spw)
	fmt.Println(err)

	myBlockchain.AddBlockToBlockchain("Пипапупа")
	myBlockchain.AddBlockToBlockchain("Полиморфин")
	myBlockchain.AddBlockToBlockchain("Берсеркерум")
	c.Iterator()

	// // Записываем несколько ключей
	// for i := 0; i < 10; i += 1 {
	// 	c.WriteValue([]byte("MyBacket"), []byte(strconv.Itoa(i)), []byte("pipapupa"+strconv.Itoa(i)))
	// }

	// // Печатаем данные
	// for i := 0; i < 10; i += 1 {
	// 	val, err := c.ReadValue([]byte("MyBacket"), []byte(strconv.Itoa(i)))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	fmt.Printf("Value = %v\n", string(val))
	// }

}
