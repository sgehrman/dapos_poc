package main

import (
	"time"
	//	"github.com/aws/aws-sdk-go/service/medialive"
)

type WalletAddress string

type WalletAccount struct {
	Id      WalletAddress
	Name    string
	Balance int
}

type Block struct {
	Prev_block  *Block
	Next_block  *Block
	Transaction Transaction
}

type Transaction struct {
	Id         int
	From       WalletAddress // $5
	To         WalletAddress
	Value      int
	Time       time.Time
	Validators []WalletAddress
}

type Node struct {
	GenesisBlock *Block
	CurrentBlock *Block
	TxChannel    chan Transaction
	VoteChannel  chan Vote
	Wallet       WalletAccount

	IsDelegate      bool
	TxFromChainById map[int]*Transaction
}
