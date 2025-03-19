package redisadapter

import (
	"bytes"
	"context"
	"log"
	"time"
	transaction "node/internal/transaction"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	RedisClient *redis.Client
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
func (r *RedisAdapter) Init() error {
	ctx := context.Background()

	// Конфигурация Sentinel
	sentinelAddrs := []string{"localhost:26379", "localhost:26380", "localhost:26381"}
	masterName := "mymaster"
	password := "mypassword"

	// Создание клиента
	r.RedisClient = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
	})

	// Проверка подключения
	if err := r.RedisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connect error: %v", err)
		return err
	}
	log.Printf("Successful connect to redis sentinel!")

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
	// Канал с поступающими транзакциями
	output := make(chan *transaction.Transaction, 16)

	// Фоновый процесс записи поступающих транзакций в канал
	go func() {
		defer close(output)

		for {
			// Подключение к нужному PubSub
			pubsub := r.RedisClient.Subscribe(context.Background(), transactionSubName)
			defer pubsub.Close()
			ch := pubsub.Channel()

			// Чтение канала с сообщениями
			for msg := range ch {
				buf := bytes.NewReader([]byte(msg.Payload))
				tr, err := transaction.DeserializeTransaction(buf)
				if err != nil {
					log.Printf("Can not deserialization transaction: %v", err)
					break
				}
				output <- &tr
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			log.Println("PubSub channel closed. Reconnecting...")
            time.Sleep(1 * time.Second)
		}
	}()

	return output
}

