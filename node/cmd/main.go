package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()

	// Конфигурация Sentinel
	sentinelAddrs := []string{"localhost:26379", "localhost:26380", "localhost:26381"}
	masterName := "mymaster"
	password := "mypassword"

	// Создание клиента
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddrs,
		Password:      password,
	})

	// Проверка подключения
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	fmt.Println("Успешное подключение к Redis через Sentinel!")

	// Тест записи и чтения
	key := "test_key_2"
	value := "Hello Redis 2!"
	
	if err := client.Set(ctx, key, value, 10*time.Second).Err(); err != nil {
		log.Fatalf("Ошибка записи: %v", err)
	}
	
	result, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Fatalf("Ошибка чтения: %v", err)
	}
	
	fmt.Printf("Получено значение: %s\n", result)
}