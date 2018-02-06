package main

import (
	"time"
	"fmt"
)

//block struct to keep the linked list
type Block struct {
	Prev_block *Block
	Next_block *Block
	Transaction Transaction
}

type Transaction struct {
	Id         int
	From       string
	To         string
	Value      int
	Time       time.Time
	DelegateId int
	Validators int
}

type Delegate struct {
	Id   		 int
	PeerCount 	 int
	GenesisBlock *Block
	CurrentBlock *Block
	Channel		 chan Transaction
	VoteChannel  chan Vote
}

type Account struct {
	Id           string
	Name         string
	Balance      int
	Transactions []Transaction
}

func (b Block) PrintBlock(delegateId int) {
	fmt.Printf("[ DelegateId=%d \tblock -> %d %d]\n", delegateId, b.Transaction.Id, b.Transaction.Value)
}