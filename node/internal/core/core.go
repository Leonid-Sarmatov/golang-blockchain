package core

import (
	//block "node/internal/block"
)

type blockchain interface {

}

type transactionReciver interface {
	
}

type blockTransmitter interface {

}

type replication interface {

}

type Core struct {
}

/*
NewCore конструктор для ядра

Возвращает:
  - *Core: экземпляр структуры ядра
*/
func NewCore() *Core {
	return &Core{}
}

/*
GetWorkForMining возвращает подготовленную для
майнинга работу: транзакция вознаграждения и
главная транзакция (в виде байтовых слайсов)

Аргументы:
  - rewardAddress []byte: адрес получателя вознаграждения

Возвращает:
  - []byte: транзакция вознаграждения
  - []byte: главная транзакция
  - error: ошибка
*/
