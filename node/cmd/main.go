package main

import (
	"context"
	"fmt"
	"log"
	redisadapter "node/internal/adapters/redis_adapter"
	"node/internal/transaction"
	"time"
	"bytes"

	//"github.com/go-redis/redis/v8"
)

func main() {
	// ctx := context.Background()

	// // Конфигурация Sentinel
	// sentinelAddrs := []string{"localhost:26379", "localhost:26380", "localhost:26381"}
	// masterName := "mymaster"
	// password := "mypassword"

	// // Создание клиента
	// client := redis.NewFailoverClient(&redis.FailoverOptions{
	// 	MasterName:    masterName,
	// 	SentinelAddrs: sentinelAddrs,
	// 	Password:      password,
	// })

	// // Проверка подключения
	// if err := client.Ping(ctx).Err(); err != nil {
	// 	log.Fatalf("Ошибка подключения: %v", err)
	// }
	// fmt.Println("Успешное подключение к Redis через Sentinel!")

	// // Тест записи и чтения
	// key := "test_key_2"
	// value := "Hello Redis 2!"
	
	// if err := client.Set(ctx, key, value, 10*time.Second).Err(); err != nil {
	// 	log.Fatalf("Ошибка записи: %v", err)
	// }
	
	// result, err := client.Get(ctx, key).Result()
	// if err != nil {
	// 	log.Fatalf("Ошибка чтения: %v", err)
	// }
	
	// fmt.Printf("Получено значение: %s\n", result)

	testTransaction := transaction.Transaction{
		TimeOfCreation: time.Now().Unix(),
		Inputs: []transaction.TransactionInput{
			{
				PreviousOutputHash: []byte("bubilda"),
				PublicKey: []byte("pipapupa"),
			},
		},
		Outputs: []transaction.TransactionOutput{
			{
				Value: -1,
				RecipientAddress: []byte("bubilda"),
				TimeOfCreation: time.Now().Unix(),
				Hash: []byte("pipapupa"),
			},
		},
	}

	redisConn := redisadapter.NewRedisAdapter()
	redisConn.Init()

	go func() {
		for {
			for tr := range redisConn.TransactionReceiver("transactions") {
				fmt.Printf("Received transaction:\n")
				fmt.Printf("Time: %d\n", tr.TimeOfCreation)
				fmt.Printf("Input PublicKey: %s\n", tr.Inputs[0].PublicKey)
				fmt.Printf("Output Recipient: %s\n\n", tr.Outputs[0].RecipientAddress)
			}
		}
	}()

	// Отправляем транзакции каждую секунду
	ctx := context.Background()
	counter := 1

	fmt.Println(" >>> ", redisConn.RedisClient)
	
	for {
		// Создаем уникальную транзакцию для каждой итерации
		newTr := testTransaction
		newTr.TimeOfCreation = time.Now().Unix()
		newTr.Inputs[0].PublicKey = []byte(fmt.Sprintf("pubkey-%d", counter))
		newTr.Outputs[0].RecipientAddress = []byte(fmt.Sprintf("recipient-%d", counter))
		
		// Сериализуем транзакцию
		var buf bytes.Buffer
		if err := transaction.SerializeTransaction(&buf, newTr); err != nil {
			log.Fatalf("Serialization error: %v", err)
		}

		// Отправляем в Redis
		err := redisConn.RedisClient.Publish(ctx, "transactions", buf.Bytes()).Err()
		if err != nil {
			log.Printf("Publish error: %v", err)
		} else {
			fmt.Printf("Sent transaction #%d\n", counter)
		}

		counter++
		time.Sleep(1 * time.Second)
	}
}