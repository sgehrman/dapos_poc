package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	logger "github.com/nic0lae/golog"
)

func GetRandomNumber(boundary int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(boundary)
}

func getDictKeysAsList() []string {
	keys := make([]string, 0)
	for k, _ := range getNodes() {
		keys = append(keys, k)
	}

	return keys
}

func getRandomNode(nodesToIgnore []*Node) *Node {
	nodes := getNodes()
	nodesNames := getDictKeysAsList()

	for {
		randomNum := GetRandomNumber(len(nodes))
		newNode := getNodes()[nodesNames[randomNum]]
		if len(nodesToIgnore) == 0 {
			return newNode
		}

		var nodeInArray = false
		for _, n := range nodesToIgnore {
			if newNode.Wallet == n.Wallet {
				nodeInArray = true
				break
			}
		}

		if !nodeInArray {
			return newNode
		}
	}
}

func sendRandomTransaction(fromWallet string, toWallet string, transactionId int, amount int, delegate *Node) {
	sendRandomTransactionWithTime(fromWallet, toWallet, transactionId, amount, delegate, time.Now())
}

func sendRandomTransactionWithTime(fromWallet string, toWallet string, transactionId int, amount int, delegate *Node, txTime time.Time) {
	transaction := Transaction{
		transactionId,
		fromWallet,
		toWallet,
		amount,
		txTime,
		fromWallet,
	}


	go func() {
		logger.Instance().LogInfo(
			GlobalLogTag, 0,
			fmt.Sprintf("sendRandomTx() | Tx_%"+strconv.Itoa(len(strconv.Itoa(NrOfTx)))+"d(%5s -> %5s) | send to delegate -> %s",
				transaction.Id,
				transaction.From,
				transaction.To,
				delegate.Wallet,
			))
		delegate.TxChannel <- transaction
		// delegate.WriteToQ(transaction)
	}()

}

var nodes *map[string]*Node

var once sync.Once

var startingBalance = 10000000

func CreateNodeAndAddToList(newMember string) {

	GenesisBlock := Block{
		Prev: nil,
		Next: nil,
		Transaction: Transaction{
			0,
			"dl",
			"dl",
			startingBalance,
			time.Now(),
			"dl",
		},
	}

	node := Node{
		GenesisBlock:    GenesisBlock,
		LastBlock:       &GenesisBlock,
		Wallet:          string(newMember),
		IsDelegate:      true,
		TxFromChainById: map[int]Transaction{},
		AllWallets:      map[string]int{},

		TxChannel: make(chan Transaction), // channels.NewInfiniteChannel(), // Utils.Infinity) //
		// txQueue:   []Transaction{},
		// rwMutex:   sync.RWMutex{},

		TxCount:               0,
		TimeForLastTx:         time.Now(),
		TotaProcessTimeInNano: 0,
	}

	node.AllWallets["dl"] = startingBalance
	getNodes()[newMember] = &node
	//since all nodes are delegates we initialize a node with listening to transactions
	node.StartListenForTx()
}

func getNodes() map[string]*Node {
	once.Do(func() {
		nodes = &map[string]*Node{}
	})
	return *nodes
}

func getNodeByAddress(address string) *Node {
	var theNode = getNodes()[string(address)]
	return theNode
}
