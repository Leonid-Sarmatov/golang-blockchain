package app

import (
	"fmt"
	//"golang_blockchain/internal/config"
	"encoding/gob"
	"golang_blockchain/internal/mediator"
	"golang_blockchain/internal/web/handlers/coins_transfer"
	createwallet "golang_blockchain/internal/web/handlers/create_wallet"
	"golang_blockchain/internal/web/handlers/get_wallet_balance"
	getwork "golang_blockchain/internal/web/handlers/get_work"
	"golang_blockchain/internal/web/handlers/send_completed_work"
	"golang_blockchain/pkg/block"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	mediator *mediator.Mediator
	server   *http.Server
}

func NewApp() (*App, error) {
	var app App

	gob.Register(&block.Block{})
	
	//
	m, err := mediator.NewMediator()
	if err != nil {
		return nil, fmt.Errorf("App init was failed: %v", err)
	}
	app.mediator = m


	//
	r := gin.Default()

	r.GET("/api/v1/wallet/balance", getwalletbalance.NewGetWallelBalanceHandler(m))
	r.POST("/api/v1/transfer", coinstransfer.NewCoinTransferHandler(m))
	r.GET("/api/v1/work", getwork.NewGetWorkHandler(m))
	r.POST("/api/v1/work/completed", sendcompletedwork.NewSendCompletedWorkHandler(m))
	r.POST("/api/v1/wallet/create", createwallet.NewCreateWalletHandler(m))

	s := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	app.server = s

	return &app, nil
}

func (app *App) Start() {
	log.Printf("Server was started!")
	if err := app.server.ListenAndServe(); err != nil {
		log.Printf("Server was stoped: %v", err)
	}
}
