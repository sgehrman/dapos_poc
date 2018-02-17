package main

import (
	"sync"
	"time"
)

var nodes *map[string]*Node

var once sync.Once

var GenesisBlock = &Block{
	Prev_block: nil,
	Next_block: nil,
	Transaction: Transaction{
		0,
		"dl",
		"dl",
		100,
		time.Now(),
		[]WalletAddress{},
	},
}

func CreateNodeAndAddToList(newMember string, initialBalance int) {
	wallet := WalletAccount{
		WalletAddress(newMember),
		newMember,
		initialBalance,
	}

	node := Node{
		GenesisBlock:    GenesisBlock,
		CurrentBlock:    GenesisBlock,
		TxChannel:       make(chan Transaction),
		VoteChannel:     make(chan Vote),
		Wallet:          wallet,
		IsDelegate:      false,
		TxFromChainById: map[int]*Transaction{},
		AllVotes:        make(map[int]*Votes),
	}

	getNodes()[newMember] = &node
}

func ElectDelegate(newMember string) {
	getNodes()[newMember].IsDelegate = true
	getNodes()[newMember].StartListenForTx()
	getNodes()[newMember].StartVoteCounting()
}

func getNodes() map[string]*Node {
	once.Do(func() {
		nodes = &map[string]*Node{}
	})
	return *nodes
}

func getNodeByAddress(address WalletAddress) *Node {
	return getNodes()[string(address)]
}
