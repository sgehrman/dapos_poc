package main

import (
	"os"
	"os/signal"
	"syscall"
	//"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	log.Info("DAPoS Simulation!")
	logSeparator()

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

	// Log Who is Who
	for _, node := range getNodes() {
		log.Infof("Delegate - %s", node.Wallet)
	}

	// Run Transactions
	go func() {
		logSeparator()
		//Creates list of the delegates
		delegateCounter := 0

		sendRandomTransaction("dl", "BobSt", 1, 1000, getNodeByAddress(names[0]))
		sendRandomTransaction("dl", "Chris", 2, 1000, getNodeByAddress(names[1]))
		sendRandomTransaction("dl", "GregM", 3, 1000, getNodeByAddress(names[2]))
		sendRandomTransaction("dl", "Muham", 4, 1000, getNodeByAddress(names[3]))

		var nrOfTransactions = 1000
		for transactionID := 1; transactionID <= nrOfTransactions; transactionID++ {
			var node1 = getRandomNode(nil)
			var node2 = getRandomNode(node1)
			toDelegatePointer := getNodeByAddress(names[delegateCounter%numOfDelegates])
			sendRandomTransaction(node1.Wallet, node2.Wallet, transactionID, 1, toDelegatePointer)
			delegateCounter++
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
