package redis

import (
	"bytes"
	"context"
	"log"
	transaction "node/internal/transaction"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	redisClient *redis.ClusterClient
}

func NewRedisAdapter() *RedisAdapter {
	return &RedisAdapter{}
}

/*
Init инициализирует подключение к redis и
выполняет всю необходимую подготовку к работе

Аргументы:
  - hostPortString string: строка с хостом и портом

Возвращает:
  - error: ошибка
*/
func (r *RedisAdapter) Init(hostPortString []string) error {
	r.redisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: hostPortString,
	})

	return nil
}

/*
TransactionReceiver возвращает канал с поступающими транзакциями

Аргументы:
  - transactionSubName string: строка названием публикации с транзакциями

Возвращает:
  - chan *transaction.Transaction: канал с указателями на транзакции
*/
func (r *RedisAdapter) TransactionReceiver(transactionSubName string) chan *transaction.Transaction {
	output := make(chan *transaction.Transaction, 16)
	ctx := context.Background()

	go func() {
		pubsub := r.redisClient.Subscribe(ctx, transactionSubName)
		defer pubsub.Close()
		defer close(output)

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				log.Printf("[STOP GORUTINE] Can not receive transaction from redis: %v", err)
				return
			}

			buf := bytes.NewReader([]byte(msg.Payload))
			tr, err := transaction.DeserializeTransaction(buf)
			if err != nil {
				log.Printf("[STOP GORUTINE] Can not deserialization transaction: %v", err)
				return
			}

			output <- &tr
		}
	}()

	return output
}

