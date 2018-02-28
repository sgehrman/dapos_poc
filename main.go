package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var oneT = 1000
var oneM = 1000 * oneT
var oneB = 1000 * oneM
var NrOfTx = 10 * oneT

var TotalTxProcessed = 0

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

	// Create Delegates
	for _, name := range names {
		CreateNodeAndAddToList(name)
	}

	// Run Transactions
	go func() {
		//Creates list of the delegates
		delegateCounter := 0

		//send initial amount to each delegate
		sendRandomTransaction("dl   ", "BobSt", 1, 1000, getNodeByAddress(names[0]))
		sendRandomTransaction("dl   ", "Chris", 2, 1000, getNodeByAddress(names[1]))
		sendRandomTransaction("dl   ", "GregM", 3, 1000, getNodeByAddress(names[2]))
		sendRandomTransaction("dl   ", "Muham", 4, 1000, getNodeByAddress(names[3]))
		//send random transactions

		transactionID := 5
		/*
			var genesisWallet = "dl   "
			var chrisWallet = names[1]

			var bobNode = getNodeByAddress(names[0])

			var nowTime = time.Now()
			var time1 = time.Unix(nowTime.Unix()+10, 0)
			var time2 = time.Unix(nowTime.Unix()+5, 0)

			sendRandomTransactionWithTime(genesisWallet, chrisWallet, transactionID, 1, bobNode, time1)
			transactionID++
			sendRandomTransactionWithTime(genesisWallet, chrisWallet, transactionID, 1, bobNode, time2)
		*/
		for ; transactionID <= NrOfTx; transactionID++ {
			//get random node1 for FROM, and random node2 for TO
			var node1 = getRandomNode(nil)
			var node2 = getRandomNode(node1)
			//send to a delegate
			toDelegatePointer := getNodeByAddress(names[delegateCounter%numOfDelegates])
			//Send Transaction
			sendRandomTransaction(node1.Wallet, node2.Wallet, transactionID, 1, toDelegatePointer)
			delegateCounter++
		}
	}()

	go func() {
		startTime := time.Now()
		for {
			if TotalTxProcessed >= (NrOfTx)-1 {
				finalTime := time.Since(startTime)
				fmt.Println(fmt.Sprintf("FINAL: %d nanosecond", finalTime))
				time.Sleep(time.Second * 5)
				break
			}
		}

		log_SeparatorLine()
		//for i := range getNodes() {
		//	//getNodes()[i].DumpLogLines()
		//}

		log_SeparatorLine()
		for i := range getNodes() {
			var node = getNodes()[i]
			var allWaletValues = []string{}
			for wallet, _ := range node.AllWallets {
				if wallet == "dl   " {
					continue
				}
				allWaletValues = append(allWaletValues, fmt.Sprintf("%s - %d", wallet, node.AllWallets[wallet]))
			}

			fmt.Println(strings.Join(allWaletValues, ", "))

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
