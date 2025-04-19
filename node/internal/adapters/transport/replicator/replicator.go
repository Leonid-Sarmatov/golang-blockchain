package replicator

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"node/internal/block"
	"node/internal/transaction"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisReplicator struct {
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
	ctx := context.Background()

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
				log.Printf("<replicator.go> От Redis пришло сообщение, распаковываю...")
				buf := bytes.NewReader([]byte(msg.Payload))
				tr, err := transaction.DeserializeTransaction(buf)
				if err != nil {
					log.Printf("Can not deserialization transaction: %v", err)
					break
				}
				log.Printf("<replicator.go> Транзакция принята!")
				output <- &tr
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			log.Println("PubSub channel with transactions closed. Reconnecting...")
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
				output <- blk
			}

			// При возникновении ошибки или закрытии канала, ждем секунду перед переподключением
			log.Println("PubSub channel with blocks closed. Reconnecting...")
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
func (r *RedisReplicator) BlockTransmitterProcess(ctx context.Context, blks <-chan *block.Block, blockSubName string) chan error {
	// Канал с ошибками
	outputs := make(chan error)

	// Фоновый процесс отправки блоков в сеть
	go func() {
		for {
			select {
			case blk := <-blks:
				// Отправка созданного данным узлом блока в сеть
				msg, err := blk.SerializeBlock()
				if err != nil {
					log.Printf("<replication.go> Ошибка отправки блока в сеть: %v", err)
					outputs <- err
					continue
				}
				r.RedisClient.Publish(context.Background(), blockSubName, string(msg))
			case <-ctx.Done():
				// Корректное завершение работы
				close(outputs)
				return
			}
		}
	}()

	return outputs
}
