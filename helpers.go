package main

import (
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func logSeparator() {
	log.Info("~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")
}
func prefixLinesWith(lines []string, prefix string) []string {
	var prefixedLines = []string{}

	for index, line := range lines {
		if index == 0 {
			prefixedLines = append(prefixedLines, line)
		} else {
			prefixedLines = append(prefixedLines, prefix+line)
		}
	}

	return prefixedLines
}

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

func getRandomNonDelegateNode(nodeToIgnore *Node) *Node {
	nodes := getNodes()
	nodesNames := getDictKeysAsList()

	var theNode *Node
	for {
		randomNum := GetRandomNumber(len(nodes))
		theNode = getNodes()[nodesNames[randomNum]]

		if nodeToIgnore != nil && nodeToIgnore.Wallet.Id == theNode.Wallet.Id {
			continue
		}

		if !theNode.IsDelegate {
			break
		}
	}

	return theNode
}

func sendRandomTransaction(fromWallet WalletAddress, toWallet WalletAddress, transactionId int, delegates []WalletAddress) {

	fromNode := getNodes()[string(fromWallet)]
	toNode := getNodes()[string(toWallet)]

	amount := 1

	transaction := Transaction{
		transactionId,
		fromNode.Wallet.Id,
		toNode.Wallet.Id,
		amount,
		time.Now(),
		[]WalletAddress{},
	}

	for _, v := range delegates {
		delegate := getNodes()[string(v)]
		go func() {
			log.Infof("sendRandomTx()  | Tx_%d(%s -> %s) | %s -> %s",
				transaction.Id,
				transaction.From,
				transaction.To,
				fromNode.Wallet.Id,
				delegate.Wallet.Id,
			)
			delegate.TxChannel <- transaction
		}()
	}
}

var nodes *map[string]*Node

var once sync.Once

var GenesisBlock = &Block{
	Prev: nil,
	Next: nil,
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
		CurrentBlock:    nil,
		TxChannel:       make(chan Transaction),
		VoteChannel:     make(chan Vote),
		Wallet:          wallet,
		IsDelegate:      false,
		TxFromChainById: map[int]*Transaction{},
		AllVotes:        make(map[int]*Votes),
	}
	node.CurrentBlock = node.GenesisBlock

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
