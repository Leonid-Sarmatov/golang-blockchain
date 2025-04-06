package grpcserver

import (
	"context"
	// "fmt"
	"log"
	"net"
	. "node/internal/adapters/transport/server/generated"
	"node/internal/transaction"

	"google.golang.org/grpc"
)

type getterFreeTransactionsOutputs interface {
	GetFreeTransactionsOutputs() ([]*transaction.TransactionOutput, error)
}

type getterBalance interface {
	GetBalance(address []byte) (int32, error)
}

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	getOutputs getterFreeTransactionsOutputs
	getBalance getterBalance
	NodeServiceServer
}

func NewServer(getOutputs getterFreeTransactionsOutputs, getBalance getterBalance) *Server {
	return &Server{
		getOutputs: getOutputs,
		getBalance: getBalance,
	}
}

func (server *Server) Start() error {
	// Создание слушателя для порта
	lis, err := net.Listen("tcp", ":40002")
	if err != nil {
		log.Printf("Can not open tcp port %v", err)
		return err
	}

	// Инициализация полей и регистрация сервера
	server.listener = lis
	server.grpcServer = grpc.NewServer()
	RegisterNodeServiceServer(server.grpcServer, server)

	// Старт сервера
	log.Println("Starting gRPC server on :40002")
	return server.grpcServer.Serve(server.listener)
}

func (server *Server) Stop() {
	log.Println("Stopping gRPC server...")
	server.grpcServer.GracefulStop()
}

func (server *Server) GetBalance(ctx context.Context, req *GetBalanceRequest) (*GetBalanceResponse, error) {
	var response GetBalanceResponse
	res, err := server.getBalance.GetBalance([]byte(req.Address))
	if err != nil {
		log.Printf("Не удалось выдать баланс кошелька %v, ошибка: %v", string(req.Address), err)
		return &response, err
	}
	response.Balance = res
	return &response, nil
}

func (server *Server) GetFreeTransactionsOutputs(ctx context.Context, req *GetFreeTransactionsOutputsRequest) (*GetFreeTransactionsOutputsResponse, error) {
	var response GetFreeTransactionsOutputsResponse
	res, err := server.getOutputs.GetFreeTransactionsOutputs()
	if err != nil {
		log.Printf("Не удалось выдать список свободных транзакций, ошибка: %v", err)
		return &response, err
	}
	response.Outputs = make([]*TransactionOutput, len(res))
	for i, val := range res {
		response.Outputs[i] = &TransactionOutput{
			Value: int32(val.Value),
			RecipientAddress: string(val.RecipientAddress),
			TimeOfCreation: val.TimeOfCreation,
			Hash: string(val.Hash),
		}
	}
	return &response, nil
}
