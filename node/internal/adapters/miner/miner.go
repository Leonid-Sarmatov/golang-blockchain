package miner

import "node/internal/block"

/*
hashcalulator описывает интерфейс для
хэш-калькулятора
*/
// type hashCalulator interface {
// 	HashCalculate(data []byte) []byte
// }

type powChecker interface {
	Chech(blk *block.Block) (bool, error)
}

type powSolver interface {
	Exec(blk *block.Block) (int, error)
}

type Miner struct {
	checker powChecker
	solver powSolver
}

