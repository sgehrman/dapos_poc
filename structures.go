package main

import (
	"time"
)

type WalletAddress string

type WalletAccount struct {
	Id      WalletAddress
	Name    string
	Balance int
}

type Block struct {
	Prev        *Block
	Next        *Block
	Transaction Transaction
}

type Transaction struct {
	Id                int
	From              WalletAddress
	To                WalletAddress
	Value             int
	Time              time.Time
	CurrentValidators []WalletAddress
	TxInfoSender      WalletAddress
}

type Node struct {
	GenesisBlock *Block
	LastBlock    *Block
	TxChannel    chan Transaction
	VoteChannel  chan Vote
	Wallet       WalletAccount

	IsDelegate      bool
	TxFromChainById map[int]*Transaction
	AllVotes        map[int]*Votes
}
