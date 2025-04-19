package redisadapter

import (
	"bytes"
	"context"
	"fmt"
	"log"

	//"mem_pool/internal/adapters/transport/client/grpc_client"
	"mem_pool/internal/transaction"
	"sync"

	"github.com/go-redis/redis/v8"
)

type RedisAdapter struct {
	Client *redis.Client
	mu     *sync.Mutex
	ctx    context.Context
	//grpc_client *grpcclient.Client
}

func NewRedisAdapter() *RedisAdapter {
	return &RedisAdapter{
		mu:  &sync.Mutex{},
		ctx: context.Background(),
		//grpc_client: gc,
	}
}

func (adapter *RedisAdapter) Init() error {
	ctx := context.Background()

	// Конфигурация Sentinel
	sentinelAddrs := []string{"sentinel1:26379", "sentinel2:26380", "sentinel3:26381"}
	masterName := "mymaster"
	password := "mypassword"

	// Создание клиента Redis через Failover
	adapter.Client = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
	})

	// Проверка подключения
	if err := adapter.Client.Ping(ctx).Err(); err != nil {
		log.Printf("<redis.go> Не удалось подключиться к Redis: %v", err)
		return err
	}
	log.Printf("<redis.go> Успешное подключение к Redis!")
	return nil
}

func (adapter *RedisAdapter) AddOutputs(outs []transaction.TransactionOutput) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	for _, out := range outs {
		outBytes, err := transaction.SerializeTransactionOutput(&out)
		if err != nil {
			return err
		}

		if err := adapter.Client.Set(adapter.ctx, string(out.Hash), outBytes, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (adapter *RedisAdapter) BlockOutput(out transaction.TransactionOutput) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	//for _, out := range outs {
	// outBytes, err := transaction.SerializeTransactionOutput(&out)
	// if err != nil {
	// 	return err
	// }

	if err := adapter.Client.Del(adapter.ctx, string(out.Hash)).Err(); err != nil {
		return err
	}
	//}

	return nil
}

func (adapter *RedisAdapter) AddTransaction(trn transaction.Transaction) error {
	adapter.mu.Lock()
	defer adapter.mu.Unlock()

	// Код сериализации странзакции
	buff := &bytes.Buffer{}
	err := transaction.SerializeTransaction(buff, trn)
	if err != nil {
		return err
	}

	err = adapter.Client.Publish(adapter.ctx, "transactions1", buff.Bytes()).Err()
	if err != nil {
		return err
	}

	log.Printf("<redis_adapter.go> Транзакция успешно сформирована и отправлена в Redis")

	return nil
}

func (adapter *RedisAdapter) GetAllUnlockOutputs() ([]transaction.TransactionOutput, error) {
	result := make([]transaction.TransactionOutput, 0)

	// Итерация по ключам с помощью SCAN
	iter := adapter.Client.Scan(adapter.ctx, 0, "*", 0).Iterator()

	for iter.Next(adapter.ctx) {
		key := iter.Val()
		val, err := adapter.Client.Get(adapter.ctx, key).Result()

		if err != nil {
			if err == redis.Nil {
				// Пропускаем ключи без значения (например, списки/хеши)
				continue
			}
			return nil, fmt.Errorf("Get val from redis was failed: %v", err)
		}

		tr, err := transaction.DeserializeTransactionOutput([]byte(val))
		if err != nil {
			return nil, fmt.Errorf("Deserialization was failed: %v", err)
		}
		// fmt.Printf("tr.Hash = %x\n", tr.Hash)
		// fmt.Printf("tr.RecipientAddress = %v\n", string(tr.RecipientAddress))
		// fmt.Printf("tr.TimeOfCreation = %v\n", tr.TimeOfCreation)
		// fmt.Printf("tr.Value = %v\n", tr.Value)
		result = append(result, *tr)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("Iterator from redis was failed: %v", err)
	}

	return result, nil
}

func (adapter *RedisAdapter) GetBalance(address []byte) (int32, error) {
	outputs, err := adapter.GetAllUnlockOutputs()
	if err != nil {
		return -1, fmt.Errorf("Can not get all outputs: %v", err)
	}
	// fmt.Printf("len(outputs) = %v\n", len(outputs))

	res := 0
	for _, val := range outputs {
		// fmt.Printf("val.RecipientAddress = %v\n", string(val.RecipientAddress))
		// fmt.Printf("address = %v\n", string(address))
		// fmt.Printf("res = %v\n", res)
		if bytes.Equal(val.RecipientAddress, address) {
			res += val.Value
		}
	}

	return int32(res), nil
}
