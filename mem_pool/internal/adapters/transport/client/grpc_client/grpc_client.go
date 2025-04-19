package grpcclient

import (
	"context"
	"log"
	. "mem_pool/internal/adapters/transport/client/generated"
	"mem_pool/internal/transaction"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	connection *grpc.ClientConn
	client NodeServiceClient
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Start() error {
	// Устанавливаем соединение с сервером
	conn, err := grpc.Dial("my-node:40002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
		return err
	}
	c.connection = conn
	//defer c.connection.Close()

	log.Printf("Connection - OK")

	c.client = NewNodeServiceClient(c.connection)
	
	return nil

	// // Отправляем запрос к серверу
	// response, err := client.ConvertCurrency(context.Background(), &CurrencyRequest{
	// 	Amount:       100.0,
	// 	FromCurrency: "EUR",
	// 	ToCurrency:   "USD",
	// })
	// if err != nil {
	// 	log.Printf("Error converting currency: %v", err)
	// } else {
	// 	fmt.Printf("Converted amount: %.2f %s\n", response.ConvertedAmount, "USD")
	// }
}

func (c *Client) Stop() {
	c.connection.Close()
}

func (c *Client) GetBalance(address []byte) (int32, error) {
	response, err := c.client.GetBalance(context.Background(), &GetBalanceRequest{
		Address: string(address),
	})
	if err != nil {
		log.Printf("Не удалось запросить баланс для кошелька %v, ошибка: %v", string(address), err)
		return -1, err
	}
	log.Printf("Баланс кошелька %v успешно получен: %v",string(address), err)
	return response.Balance, nil
}

func (c *Client) GetFreeTransactionsOutputs() ([]*transaction.TransactionOutput, error) {
	response, err := c.client.GetFreeTransactionsOutputs(context.Background(), &GetFreeTransactionsOutputsRequest{
		MaxOutputs: 32,
	}) 

	if err != nil {
		log.Printf("<grpc_client.go> Не удалось запросить свободные выходы, ошибка: %v", err)
		return nil, err
	}
	log.Printf("<grpc_client.go> Свободные выходы успешно получены")

	res := make([]*transaction.TransactionOutput, len(response.Outputs))
	for i, val := range response.Outputs {
		res[i] = &transaction.TransactionOutput{
			Value: int(val.Value),
			RecipientAddress: []byte(val.RecipientAddress),
			TimeOfCreation: val.TimeOfCreation,
			Hash: val.Hash,
		}
	}
	return res, nil
}
