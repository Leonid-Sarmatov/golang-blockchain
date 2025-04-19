package main

import (
	// "bytes"
	// "context"
	// "fmt"
	"log"
	"time"
	// "log"

	"node/internal/adapters/blockchain_adapter"
	"node/internal/adapters/miner"
	"node/internal/adapters/pow"
	"node/internal/adapters/storage"
	"node/internal/adapters/transport/replicator"
	"node/internal/adapters/transport/server/grpc_server"
	"node/internal/blockchain"
	"node/internal/core"
	//"node/internal/transaction"
	//"time"
	//"github.com/go-redis/redis/v8"
)

func main() {
	time.Sleep(8 * time.Second)

	hc := pow.NewHashCalculator()

	s := storage.NewBBoltDBDriver()
	b := blockchain.NewBlockchain(s, hc)
	
	checker := pow.NewProofOfWorkCheker()
	solver := pow.NewProofOfWorkSolver()

	ba := blockchainadapter.NewBlockchainAdapter(b)
	_, err := ba.TryLoadSavedBlockchain()
	if err != nil {
		log.Fatalf("Не удалось загрузить блокчейн с диска: %v", err)
	}

	g := grpcserver.NewServer(ba, ba)
	go func() {
		g.Start()
	}()
	
	m := miner.NewMiner(checker, solver, s)
	err = m.Init()
	if err != nil {
		log.Fatalf("Не удалось инициализировать майнер: %v", err)
	}

	r := replicator.NewRedisAdapter()
	err = r.Init()
	if err != nil {
		log.Fatalf("Не удалось инициализировать redis: %v", err)
	}

	c := core.NewCore(r, r, r, ba, m)
	c.Init()

	for {
		
	}

	// testTransaction := transaction.Transaction{
	// 	TimeOfCreation: time.Now().Unix(),
	// 	Inputs: []transaction.TransactionInput{
	// 		{
	// 			PreviousOutputHash: []byte("bubilda"),
	// 			PublicKey: []byte("pipapupa"),
	// 		},
	// 	},
	// 	Outputs: []transaction.TransactionOutput{
	// 		{
	// 			Value: -1,
	// 			RecipientAddress: []byte("bubilda"),
	// 			TimeOfCreation: time.Now().Unix(),
	// 			Hash: []byte("pipapupa"),
	// 		},
	// 	},
	// }

	// time.Sleep(2 * time.Second)

	// redisConn := replicator.NewRedisAdapter()
	// redisConn.Init()

	// go func() {
	// 	for tr := range redisConn.TransactionReceiverProcess("transactions1") {
	// 		fmt.Printf("Received transaction:\n")
	// 		fmt.Printf("Time: %d\n", tr.TimeOfCreation)
	// 		fmt.Printf("Input PublicKey: %s\n", tr.Inputs[0].PublicKey)
	// 		fmt.Printf("Output Recipient: %s\n\n", tr.Outputs[0].RecipientAddress)
	// 	}
	// }()

	// for {

	// }

	// // Отправляем транзакции каждую секунду
	// ctx := context.Background()
	// counter := 1

	// fmt.Println(" >>> ", redisConn.RedisClient)
	
	// for {
	// 	// Создаем уникальную транзакцию для каждой итерации
	// 	newTr := testTransaction
	// 	newTr.TimeOfCreation = time.Now().Unix()
	// 	newTr.Inputs[0].PublicKey = []byte(fmt.Sprintf("pubkey-%d", counter))
	// 	newTr.Outputs[0].RecipientAddress = []byte(fmt.Sprintf("recipient-%d", counter))
		
	// 	// Сериализуем транзакцию
	// 	var buf bytes.Buffer
	// 	if err := transaction.SerializeTransaction(&buf, newTr); err != nil {
	// 		log.Fatalf("Serialization error: %v", err)
	// 	}

	// 	// Отправляем в Redis
	// 	err := redisConn.RedisClient.Publish(ctx, "transactions1", buf.Bytes()).Err()
	// 	if err != nil {
	// 		log.Printf("Publish error: %v", err)
	// 	} else {
	// 		fmt.Printf("Sent transaction #%d\n", counter)
	// 	}

	// 	counter++
	// 	time.Sleep(2 * time.Second)
	// }
}