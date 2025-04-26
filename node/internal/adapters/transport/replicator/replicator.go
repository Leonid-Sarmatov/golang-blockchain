package replicator

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"node/internal/block"
	"node/internal/transaction"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisReplicator struct {
	ctx         context.Context
	mu          *sync.Mutex
	RedisClient *redis.Client
}

func NewRedisAdapter() *RedisReplicator {
	return &RedisReplicator{}
}

/*
Init инициализирует подключение к redis и
выполняет всю необходимую подготовку к работе

Аргументы:
  - hostPortString string: строка с хостом и портом

Возвращает:
  - error: ошибка
*/
func (r *RedisReplicator) Init() error {
	// Конфигурация Sentinel
	sentinelAddrs := []string{"sentinel1:26379", "sentinel2:26380", "sentinel3:26381"}
	masterName := "mymaster"
	password := "mypassword"

	// Создание клиента
	r.RedisClient = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
	})

	// Создание контекста
	r.ctx = context.Background()
	r.mu = &sync.Mutex{}

	// Проверка подключения
	if err := r.RedisClient.Ping(r.ctx).Err(); err != nil {
		log.Printf("Redis connect error: %v", err)
		return err
	}
	log.Printf("Successful connect to redis sentinel!")

	return nil
}

func (r *RedisReplicator) AddOutput(out transaction.TransactionOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	outBytes, err := transaction.SerializeTransactionOutput(&out)
	if err != nil {
		return err
	}

	if err := r.RedisClient.Set(r.ctx, string(out.Hash), outBytes, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisReplicator) BlockOutput(out transaction.TransactionOutput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.RedisClient.Del(r.ctx, string(out.Hash)).Err(); err != nil {
		return err
	}

	return nil
}

/*
GetAllUnlockOutputs находит все свободные
выходы транзакций запрашивая их из redis

Возвращает:
  - []*transaction.TransactionOutput: слайс транзакций
  - error: ошибка
*/
func (r *RedisReplicator) GetAllUnlockOutputs() ([]*transaction.TransactionOutput, error) {
	result := make([]*transaction.TransactionOutput, 0)

	// Итерация по ключам с помощью SCAN
	iter := r.RedisClient.Scan(r.ctx, 0, "*", 0).Iterator()

	for iter.Next(r.ctx) {
		key := iter.Val()
		val, err := r.RedisClient.Get(r.ctx, key).Result()

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

		result = append(result, tr)
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("Iterator from redis was failed: %v", err)
	}

	return result, nil
}

/*
TransactionReceiver возвращает канал с поступающими транзакциями

Аргументы:
  - transactionSubName string: строка названием публикации с транзакциями

Возвращает:
  - chan *transaction.Transaction: канал с указателями на транзакции
*/
func (r *RedisReplicator) TransactionReceiverProcess(transactionSubName string) chan *transaction.Transaction {
	// Канал с поступающими транзакциями
	output := make(chan *transaction.Transaction)

	// Фоновый процесс записи поступающих транзакций в канал
	go func() {
		defer close(output)
		fmt.Printf("<replicator.go> Запуск фонового процесса...")

		for {
			// Подключение к нужному PubSub
			pubsub := r.RedisClient.Subscribe(context.Background(), transactionSubName)
			//defer pubsub.Close()
			ch := pubsub.Channel()

			// Чтение канала с сообщениями
			for msg := range ch {
				buf := bytes.NewReader([]byte(msg.Payload))
				tr, err := transaction.DeserializeTransaction(buf)
				if err != nil {
					log.Printf("Can not deserialization transaction: %v", err)
					break
				}
				log.Printf("<replicator.go> Пришла транзакция от mem-pool")
				output <- &tr
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			log.Println("<replicator.go> PubSub channel with transactions closed. Reconnecting...")
			time.Sleep(1 * time.Second)
		}
	}()

	return output
}

/*
BlockReceiver возвращает канал с поступающими транзакциями

Аргументы:
  - blockSubName string: строка названием публикации сблоками

Возвращает:
  - chan *block.Block: канал с указателями на приходящие блоки
*/
func (r *RedisReplicator) BlockReceiverProcess(blockSubName string) chan *block.Block {
	// Канал с поступающими транзакциями
	output := make(chan *block.Block)

	// Фоновый процесс записи поступающих транзакций в канал
	go func() {
		defer close(output)

		for {
			// Подключение к нужному PubSub
			pubsub := r.RedisClient.Subscribe(context.Background(), blockSubName)
			defer pubsub.Close()
			ch := pubsub.Channel()

			// Чтение канала с сообщениями
			for msg := range ch {
				blk, err := block.DeserializeBlock([]byte(msg.Payload))
				if err != nil {
					log.Printf("Can not deserialization block: %v", err)
					break
				}
				log.Printf("<replicator.go> Пришел блок от сети. HASH = %x. POW = %v", blk.Hash, blk.ProofOfWorkValue)
				output <- blk
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			log.Println("<replicator.go> PubSub channel with blocks closed. Reconnecting...")
			time.Sleep(1 * time.Second)
		}
	}()

	return output
}

/*
BlockTransmitter создает процесс для отправки
новых блоков в сеть

Аргументы:
  - ctx context.Context: контекст для отмены
  - blks <-chan *block.Block: канал с блоками для отправки
  - blockSubName string: строка названием публикации сблоками

Возвращает:
  - chan erro: канал возникающих ошибок
*/
func (r *RedisReplicator) BlockTransmitterProcess(ctx context.Context, blks <-chan *block.Block, blockSubName string) {
	// Фоновый процесс отправки блоков в сеть
	go func() {
		for {
			select {
			case blk := <-blks:
				// Отправка созданного данным узлом блока в сеть
				msg, err := blk.SerializeBlock()
				if err != nil {
					log.Printf("<replicator.go> Ошибка отправки блока в сеть: %v", err)
					continue
				}
				log.Printf("<replicator.go> Блок отправляется в сеть. HASH = %x. POW = %v", blk.Hash, blk.ProofOfWorkValue)

				r.RedisClient.Publish(context.Background(), blockSubName, string(msg))
			case <-ctx.Done():
				return
			}
		}
	}()
}
