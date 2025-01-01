package blockchain

import (
	block "golang_blockchain/pkg/block"
)

/* Сама структура блокчейна */
type Blockchain struct {
	Blocks []*block.Block
}

