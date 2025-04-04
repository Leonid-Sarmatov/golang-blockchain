package api

import (
	"log"
	"net/http"
	//"node/internal/adapters/transport/server/http_server/handlers/create_wallet"
	//"node/internal/adapters/transport/server/http_server/handlers/get_wallet_balance"
	"node/internal/core"
	"time"

	"github.com/gin-gonic/gin"
)

type Api struct {
	server *http.Server
	core *core.Core
}

func NewApi(core *core.Core) *Api {
	return &Api{}
}

func (api *Api) Init() {
	r := gin.Default()

	//r.GET("/api/v1/wallet/balance", getwalletbalance.NewGetWallelBalanceHandler(api.core))
	//r.POST("/api/v1/wallet/create", createwallet.NewCreateWalletHandler(api.core))

	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	api.server = s

	go func() {
		log.Printf("Server was started!")
		if err := api.server.ListenAndServe(); err != nil {
			log.Printf("Server was stoped: %v", err)
		}
	}()
}

