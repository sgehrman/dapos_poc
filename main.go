package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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

	var names = []string{
		"BobSt",
		"Chris",
		"GregM",
		"Muham",
		"Nicol",
		"ZaneW",
		"Avery",
	}

	// Create Nodes
	for _, name := range names {
		CreateNodeAndAddToList(name, 100)
	}

	// Elect Delegates
	var nrOfDelegates = 3
	for count := 0; count < nrOfDelegates; count++ {
		var node = getRandomNonDelegateNode(nil)
		ElectDelegate(string(node.Wallet.Id))
	}

	// Log Who is Who
	for _, node := range getNodes() {
		if node.IsDelegate {
			log.Infof("Delegate - %s", node.Wallet.Id)
		} else {
			log.Infof("Node     - %s", node.Wallet.Id)
		}
	}

	// Run Transactions
	go func() {
		logSeparator()
		time.Sleep(time.Second * 5)

		var delegates = []WalletAddress{}
		for _, node := range getNodes() {
			if node.IsDelegate {
				delegates = append(delegates, node.Wallet.Id)

				break
			}
		}

		var nrOfTransactions = 1
		for transactionID := 1; transactionID <= nrOfTransactions; transactionID++ {
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
		<-sigs
		done <- true
	}()
	<-done
}
