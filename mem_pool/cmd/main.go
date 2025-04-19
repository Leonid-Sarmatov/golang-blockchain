package main

import (
	// "bytes"
	// "context"
	// "fmt"
	"log"

	"mem_pool/internal/adapters/pow"
    "mem_pool/internal/adapters/redis_adapter"

	"mem_pool/internal/adapters/transport/client/grpc_client"
	"mem_pool/internal/adapters/transport/server/http_server/handlers/coins_transfer"
	"mem_pool/internal/adapters/transport/server/http_server/handlers/create_wallet"
	"mem_pool/internal/adapters/transport/server/http_server/handlers/get_wallet_balance"
	"mem_pool/internal/core"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	time.Sleep(4 * time.Second)

	hc := pow.NewHashCalculator()

	c := grpcclient.NewClient()
	c.Start()

	rds := redisadapter.NewRedisAdapter()
	rds.Init()

	core := core.NewCore(rds, rds, rds, hc, c)
	core.Init()

	r := gin.Default()

	r.GET("/api/v1/wallet/balance", getwalletbalance.NewGetWallelBalanceHandler(core))
	r.POST("/api/v1/transfer", coinstransfer.NewCoinTransferHandler(core))
	r.POST("/api/v1/wallet/create", createwallet.NewCreateWalletHandler(core))

	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Server was started!")
	if err := s.ListenAndServe(); err != nil {
		log.Printf("Server was stoped: %v", err)
	}

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

	// redisConn := redisadapter.NewRedisAdapter()
	// redisConn.Init()

	// // Отправляем транзакции каждую секунду
	// ctx := context.Background()
	// counter := 1

	// fmt.Println(" >>> ", redisConn.Client)
	
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
	// 	err := redisConn.Client.Publish(ctx, "transactions1", buf.Bytes()).Err()
	// 	if err != nil {
	// 		log.Printf("Publish error: %v", err)
	// 	} else {
	// 		fmt.Printf("Sent transaction #%d\n", counter)
	// 	}

	// 	counter++
	// 	time.Sleep(2 * time.Second)
	// }
}
