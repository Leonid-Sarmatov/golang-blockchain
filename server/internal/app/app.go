package app

import (
	"fmt"
	//"golang_blockchain/internal/config"
	"golang_blockchain/internal/mediator"
	getwalletbalance "golang_blockchain/internal/web/handlers/get_wallet_balance"
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

	//
	m, err := mediator.NewMediator()
	if err != nil {
		return nil, fmt.Errorf("App init was failed: %v", err)
	}
	app.mediator = m


	//
	r := gin.Default()

	r.POST("/api/v1/getWalletBalance", getwalletbalance.NewGetWallelBalanceHandler(m))

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
