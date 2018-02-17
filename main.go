package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
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

func main() {
	fmt.Println("DAPoS Simulation!")

	// Create Nodes
	CreateNodeAndAddToList("BobSt", 100)
	CreateNodeAndAddToList("Chris", 100)
	CreateNodeAndAddToList("GregM", 100)
	CreateNodeAndAddToList("Muham", 100)
	CreateNodeAndAddToList("Nicol", 100)
	CreateNodeAndAddToList("ZaneW", 100)
	CreateNodeAndAddToList("Avery", 100)

	// Elect Delegates
	ElectDelegate("BobSt")
	ElectDelegate("Chris")
	ElectDelegate("GregM")
	ElectDelegate("Avery")

	var delegates = []WalletAddress{}
	for _, node := range getNodes() {
		if node.IsDelegate {
			delegates = append(delegates, node.Wallet.Id)
		}
	}

	// Run Transactions - DONE by nodes, not Delegates
	go func() {
		time.Sleep(time.Second * 5)

		for transactionID := 1; transactionID < 2; transactionID++ {
			var node1 = getRandomNonDelegateNode(nil)
			var node2 = getRandomNonDelegateNode(node1)
			sendRandomTransaction(node1.Wallet.Id, node2.Wallet.Id, transactionID, delegates)
		}
	}()

	// Wait for Ctrl + C
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	<-done
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
			fmt.Println(fmt.Sprintf("sendRandomTx()  | Tx_%d(%s -> %s) | %s -> %s",
				transaction.Id,
				transaction.From,
				transaction.To,
				fromNode.Wallet.Id,
				delegate.Wallet.Id,
			))
			delegate.TxChannel <- transaction
		}()
	}
}
