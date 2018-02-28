package main

import (
	"sync"
	"time"
)

type Block struct {
	Prev        *Block
	Next        *Block
	Transaction Transaction
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
	Wallet       string

	txQueue   []Transaction
	rwMutex   sync.RWMutex
	TxChannel chan Transaction // *channels.InfiniteChannel //

	IsDelegate      bool
	TxFromChainById map[int]Transaction
	AllWallets      map[string]int

	TxCount               int
	TimeForLastTx         time.Time
	TotaProcessTimeInNano int64
}
