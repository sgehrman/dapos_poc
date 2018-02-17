package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var names = []string{"Bob", "Chris", "Greg", "Muhammad", "Nicolae", "Zane"}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.ErrorLevel)

	log.Info("Dapos POC Starting")
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

func getRandomWallet() *WalletAccount {
	nodes := getNodes()
	randomNum := GetRandomNumber(len(nodes))

	nodesNames := getDictKeysAsList()

	return &(*getNodes()[nodesNames[randomNum]]).Wallet
}

func getRandomNonDelegateNode() *Node {
	nodes := getNodes()
	nodesNames := getDictKeysAsList()

	var theNode *Node
	for {
		randomNum := GetRandomNumber(len(nodes))
		theNode = getNodes()[nodesNames[randomNum]]

		if !theNode.IsDelegate {
			break
		}
	}

	return theNode
}

func main() {
	fmt.Println("DAPoS Simulation!")

	// Create Nodes
	CreateNodeAndAddToList("Bob", 100)
	CreateNodeAndAddToList("Chris", 100)
	CreateNodeAndAddToList("Greg", 100)
	CreateNodeAndAddToList("Muhammad", 100)
	CreateNodeAndAddToList("Nicolae", 100)
	CreateNodeAndAddToList("Zane", 100)
	CreateNodeAndAddToList("Avery", 100)

	// Elect Delegates
	ElectDelegate("Bob")
	ElectDelegate("Chris")
	ElectDelegate("Greg")
	ElectDelegate("Avery")

	var delegates = []WalletAddress{}
	for _, node := range getNodes() {
		if node.IsDelegate {
			delegates = append(delegates, node.Wallet.Id)
		}
	}

	// Run Transactions - DONE by nodes, not Delegates
	for transactionID := 1; transactionID < 1000; transactionID++ {
		var node = getRandomNonDelegateNode()
		node.SendRandomTransaction(transactionID, delegates)
	}
}

func sendTransaction(transactionId int, delegates []WalletAddress) {

	fromWallet := getRandomWallet()
	toWallet := getRandomWallet()
	//amount := GetRandomNumber(20)

	amount := 1

	log.WithFields(log.Fields{
		"From ":   fromWallet.Id,
		"To ":     toWallet.Id,
		"Amount ": amount,
	}).Info("Transaction receipt")

	transaction := Transaction{
		transactionId,
		fromWallet.Id,
		toWallet.Id,
		amount,
		time.Now(),
		delegates,
	}

	log.WithFields(log.Fields{
		"From ":         fromWallet.Id,
		"From Balance ": fromWallet.Balance,
		"To ":           toWallet.Id,
		"To Balance":    toWallet.Balance,
	}).Info("Balances")

	log.WithFields(log.Fields{
		"Transaction ID":    transaction.Id,
		"Transaction From":  transaction.From,
		"Transaction To":    transaction.To,
		"Transaction Value": transaction.Value,
	}).Info("Sending Transaction")

	for k, _ := range getNodes() {
		if getNodes()[k].IsDelegate {
			getNodes()[k].TxChannel <- transaction
		}
	}
}
