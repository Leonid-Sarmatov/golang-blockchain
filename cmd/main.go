package main

import (
	"golang_blockchain/pkg/blockchain"
	proofofwork "golang_blockchain/pkg/proof_of_work"
)

func main() {
	// Создаем проверяльщика работы
	spw := proofofwork.NewProofOfWork()

	// Создаем генезис
	genesis := blockchain.NewGenesisBlock(spw)

	// Добавляем блок для генезиса
	blockchain.NewBlock("Hello, blockchain!", genesis.Hash, spw)
}
