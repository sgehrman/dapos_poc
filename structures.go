package main

import (
	"time"
)

type Block struct {
	Prev        *Block
	Next        *Block
	Transaction *Transaction
}

type Transaction struct {
	Id    int
	From  string
	To    string
	Value int
	Time  time.Time
	DelId string
}

type Node struct {
	GenesisBlock Block
	LastBlock    *Block
	TxChannel    chan Transaction
	Wallet       string
	TxCount      int
	StartTime    time.Time

	IsDelegate      bool
	TxFromChainById map[int]*Transaction
	AllWallets      map[string]int

	LogLines []string
}
