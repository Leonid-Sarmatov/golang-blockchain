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

		err = client.SendCompletedWorkRequest(work, pow)
		if err != nil {
			log.Printf("Ошибка при отправке работы: %v\n", err)
			continue
		}
	}
}