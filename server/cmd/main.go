package main

import (
	"fmt"

	//"golang_blockchain/internal/services/transaction"
	"golang_blockchain/internal/app"
	//transactioncontroller "golang_blockchain/internal/services/transaction_controller"
	//"golang_blockchain/pkg/block"
	"github.com/redis/go-redis/v9"
	"context"
)

func main() {
	// Создаем клиент Redis
    client := redis.NewClient(&redis.Options{
        Addr:     "185.217.198.251:6379", // Замените на адрес вашего Redis-сервера
        Password: "",              // Пароль, если требуется
        DB:       0,               // Номер базы данных
    })

    // Проверяем подключение к Redis
    ctx := context.Background()
    pong, err := client.Ping(ctx).Result()
    if err != nil {
        fmt.Println("Ошибка подключения к Redis:", err)
        return
    }
    fmt.Println("Подключение к Redis успешно:", pong)

    // Пример работы с Redis: запись и чтение значения
    err = client.Set(ctx, "ключ", "значение", 0).Err()
    if err != nil {
        fmt.Println("Ошибка записи в Redis:", err)
        return
    }

    val, err := client.Get(ctx, "ключ").Result()
    if err != nil {
        fmt.Println("Ошибка чтения из Redis:", err)
        return
    }
    fmt.Println("Полученное значение из Redis:", val)


	app, err := app.NewApp()
	fmt.Println(err)

	app.Start()
	//controller, err := transactioncontroller.NewTransactionController()
	//fmt.Println(err)

	//controller.CreateNewCoinBase(20, []byte("Alice"), []byte("Alice"))
	//controller.CreateNewCoinBase(20, []byte("Bob"), []byte("Bob"))
	//controller.CreateNewCoinBase(80, []byte("Minnya"), []byte("Minnya"))

	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Alice")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Bob")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Minnya")))

	//controller.CreateCoinTransfer(5, []byte("Alice"), []byte("Bob"))


	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Alice")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Bob")))
	//fmt.Println(controller.GetBalanceByPublicKey([]byte("Minnya")))
}
