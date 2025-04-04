package outputpool

import (
	"bytes"
	"context"
	"log"
	"mem_pool/internal/transaction"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	client *redis.Client
	mu *sync.Mutex
	ctx context.Context
}

func NewRedisAdapter() *RedisAdapter {
	return &RedisAdapter{
		mu: &sync.Mutex{},
		ctx: context.Background(),
	}
}

func (adapter *RedisAdapter) Init() error {
	ctx := context.Background()

	// Конфигурация Sentinel
	sentinelAddrs := []string{"localhost:26379", "localhost:26380", "localhost:26381"}
	masterName := "mymaster"
	password := "mypassword"

	// Создание клиента Redis через Failover
	adapter.client = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
	})

	// Проверка подключения
	if err := adapter.client.Ping(ctx).Err(); err != nil {
		log.Printf("<redis.go> Не удалось подключиться к Redis: %v", err)
		return err
	}
	log.Printf("<redis.go> Успешное подключение к Redis!")
	return nil
}

func (adapter *RedisAdapter)AddOutputs(outs []*transaction.TransactionOutput) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	for _, out := range outs {
		outBytes, err := transaction.SerializeTransactionOutput(out)
		if err != nil {
			return err
		}

		if err := adapter.client.Set(adapter.ctx, string(out.Hash), outBytes, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (adapter *RedisAdapter)RemoveOutputs(outs []*transaction.TransactionOutput) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	for _, out := range outs {
		outBytes, err := transaction.SerializeTransactionOutput(out)
		if err != nil {
			return err
		}

		if err := adapter.client.Set(adapter.ctx, string(out.Hash), outBytes, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (adapter *RedisAdapter)AddTransaction(trn transaction.Transaction) error {
	// Код сериализации странзакции
	buff := &bytes.Buffer{}
	err := transaction.SerializeTransaction(buff, trn)
	if err != nil {
		return err
	}

	err = adapter.client.Publish(adapter.ctx, "transactions1", buff.Bytes()).Err()
	if err != nil {
		return err
	}

	return nil
}

/*
Код клиента:

func (r *RedisReplicator) Init() error {
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

// Подключение к нужному PubSub
pubsub := r.RedisClient.Subscribe(context.Background(), transactionSubName)
*/
