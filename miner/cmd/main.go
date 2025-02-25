package main

import (
	"encoding/base64"
	//"encoding/gob"
	//"fmt"
	"log"
	"miner/internal/http_client"
	"miner/internal/miner"
	//"miner/pkg/block"
	"time"
)

func main() {
	client := httpclient.NewHttpCleint()
	miner := miner.NewMiner()

	// block1 := &block.Block{
	// 	TimeOfCreation: time.Now().Unix(),
	// 	Data: []byte{1, 2, 3, 4},
	// 	Hash: []byte{5, 6, 7, 8},
	// 	PrevBlockHash: []byte{9, 10, 11, 12},
	// 	ProofOfWorkValue: -1,
	// }

	// block2 := &block.Block{
	// 	TimeOfCreation: time.Now().Unix(),
	// 	Data: []byte{1, 2, 3, 4},
	// 	Hash: []byte{5, 6, 7, 8},
	// 	PrevBlockHash: []byte{9, 10, 11, 12},
	// 	ProofOfWorkValue: -1,
	// }

	// a, _ := block1.BlockToBytes()
	// fmt.Printf("Блок 1 = %x\n", a)

	// b, _ := block2.BlockToBytes()
	// fmt.Printf("Блок 2 = %x\n", b)

	// block3 := &block.Block{}
	// block3.BytesToBlock(a)
	// fmt.Println(block3.ProofOfWorkValue)

	// block4 := &block.Block{}
	// block4.BytesToBlock(b)
	// fmt.Println(block4.ProofOfWorkValue)


	// a, _ = block1.BlockToBytes()
	// fmt.Printf("Блок 3 = %x\n", a)

	// b, _ = block2.BlockToBytes()
	// fmt.Printf("Блок 4 = %x\n", b)

	// return
	for {
		time.Sleep(5 * time.Second)

		work, err := client.GiveWorkRequest()
		if err != nil {
			log.Printf("Ошибка при получении работы: %v\n", err)
			continue
		}

		decodedWork, _ := base64.StdEncoding.DecodeString(work.Block)
		pow, err := miner.Do(decodedWork)
		if err != nil {
			log.Printf("Ошибка при выполнении работы 1: %v\n", err)
			continue
		}

		// decodedWork2, _ := base64.StdEncoding.DecodeString(work.MainBlock)
		// pow2, err := miner.Do(decodedWork2)
		// if err != nil {
		// 	log.Printf("Ошибка при выполнении работы 2: %v\n", err)
		// 	continue
		// }

		err = client.SendCompletedWorkRequest(work, pow)
		if err != nil {
			log.Printf("Ошибка при отправке работы: %v\n", err)
			continue
		}
	}
}