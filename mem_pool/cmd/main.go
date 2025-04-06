package main

import (
	"log"
	"mem_pool/internal/adapters/pow"
	redisadapter "mem_pool/internal/adapters/redis_adapter"
	grpcclient "mem_pool/internal/adapters/transport/client/grpc_client"
	coinstransfer "mem_pool/internal/adapters/transport/server/http_server/handlers/coins_transfer"
	createwallet "mem_pool/internal/adapters/transport/server/http_server/handlers/create_wallet"
	getwalletbalance "mem_pool/internal/adapters/transport/server/http_server/handlers/get_wallet_balance"
	"mem_pool/internal/core"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	time.Sleep(2 * time.Second)

	hc := pow.NewHashCalculator()

	c := grpcclient.NewClient()
	c.Start()

	rds := redisadapter.NewRedisAdapter()
	rds.Init()

	core := core.NewCore(rds, rds, rds, hc, c)
	core.Init()

	r := gin.Default()

	r.GET("/api/v1/wallet/balance", getwalletbalance.NewGetWallelBalanceHandler(core))
	r.POST("/api/v1/transfer", coinstransfer.NewCoinTransferHandler(core))
	r.POST("/api/v1/wallet/create", createwallet.NewCreateWalletHandler(core))

	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	log.Printf("Server was started!")
	if err := s.ListenAndServe(); err != nil {
		log.Printf("Server was stoped: %v", err)
	}

	for {

	}
}
