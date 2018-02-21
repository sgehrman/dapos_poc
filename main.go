package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("DAPoS Simulation!")
	log_SeparatorLine()

	var numOfDelegates = 4
	var names = []string{
		"BobSt",
		"Chris",
		"GregM",
		"Muham",
	}

	// Create Nodes
	for _, name := range names {
		CreateNodeAndAddToList(name)
	}

	// Run Transactions
	go func() {
		//Creates list of the delegates
		delegateCounter := 0

		sendRandomTransaction("dl   ", "BobSt", 1, 1000, getNodeByAddress(names[0]))
		sendRandomTransaction("dl   ", "Chris", 2, 1000, getNodeByAddress(names[1]))
		sendRandomTransaction("dl   ", "GregM", 3, 1000, getNodeByAddress(names[2]))
		sendRandomTransaction("dl   ", "Muham", 4, 1000, getNodeByAddress(names[3]))

		var nrOfTransactions = 5
		for transactionID := 1; transactionID <= nrOfTransactions; transactionID++ {
			var node1 = getRandomNode(nil)
			var node2 = getRandomNode(node1)
			toDelegatePointer := getNodeByAddress(names[delegateCounter%numOfDelegates])
			sendRandomTransaction(node1.Wallet, node2.Wallet, transactionID, 1, toDelegatePointer)
			delegateCounter++
		}
	}()

	go func() {
		time.Sleep(time.Second * 15) // FIXME: find a way to wait for all processins to be finished

		for i := range getNodes() {
			getNodes()[i].DumpLogLines()
		}
	}()

	// Wait for Ctrl + C
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()
	<-done
}
