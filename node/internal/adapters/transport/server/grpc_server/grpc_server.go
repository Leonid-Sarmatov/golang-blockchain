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
	lis, err := net.Listen("tcp", ":40001")
	if err != nil {
		log.Printf("Can not open tcp port %v", err)
		return err
	}

	// Инициализация полей и регистрация сервера
	server.listener = lis
	server.grpcServer = grpc.NewServer()
	RegisterNodeServiceServer(server.grpcServer, server)

	// Старт сервера
	log.Println("Starting gRPC server on :40001")
	return server.grpcServer.Serve(server.listener)
}

func (server *Server) Stop() {
	log.Println("Stopping gRPC server...")
	server.grpcServer.GracefulStop()
}

func (server *Server) GetBalance(context.Context, *GetBalanceRequest) (*GetBalanceResponse, error) {
	return nil, nil
}

func (server *Server) GetFreeTransactionsOutputs(context.Context, *GetFreeTransactionsOutputsRequest) (*GetFreeTransactionsOutputsResponse, error) {
	return nil, nil
}
