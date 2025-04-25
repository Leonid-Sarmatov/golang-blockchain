package main

import (
	// "bytes"
	// "context"
	// "fmt"
	"log"
	"time"

	// "log"

	"node/internal/adapters/blockchain"
	"node/internal/adapters/miner"
	"node/internal/adapters/pow"
	"node/internal/adapters/storage"
	"node/internal/adapters/transport/replicator"
	"node/internal/adapters/transport/server/grpc_server"
	//"node/internal/blockchain"
	"node/internal/core"
	//"node/internal/transaction"
	//"time"
	//"github.com/go-redis/redis/v8"
)

func main() {

	// ch1 := make(chan int)
	// ch2 := make(chan int)

	// pipapupa := make(map[chan int]string)

	// pipapupa[ch1] = "1234"
	// pipapupa[ch2] = "5678"

	// fmt.Printf("pipapupa = %v", pipapupa)
	// return

	time.Sleep(6 * time.Second)

	// Инициализация калькулятора хэшей
	hc := pow.NewHashCalculator()

	// Инициализация дискового хранилища
	s := storage.NewBBoltDBDriver()

	// Инициализация и загрузка блокчейна
	b := blockchain.NewBlockchain(s, hc)
	err := b.TryLoadSavedBlockchain()
	if err != nil {
		log.Fatalf("Не удалось инициализировать блокчейн: %v", err)
	}
	
	// Инициализация системы проверки proof-of-work
	checker := pow.NewProofOfWorkCheker()

	// Инициализация системы решения задачи proof-of-work
	solver := pow.NewProofOfWorkSolver()

	// ba := blockchainadapter.NewBlockchainAdapter(b)
	// _, err := ba.TryLoadSavedBlockchain()
	// if err != nil {
	// 	log.Fatalf("Не удалось загрузить блокчейн с диска: %v", err)
	// }

	r := replicator.NewRedisAdapter()
	err = r.Init()
	if err != nil {
		log.Fatalf("Не удалось инициализировать redis: %v", err)
	}

	g := grpcserver.NewServer(b, b)
	go func() {
		g.Start()
	}()
	
	m := miner.NewMiner(checker, solver, s, r, hc)
	err = m.Init()
	if err != nil {
		log.Fatalf("Не удалось инициализировать майнер: %v", err)
	}

	c := core.NewCore(r, r, r, b, m)
	c.Init()

	for {
		
	}

	for {
		
	}
}