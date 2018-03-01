package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	logger "github.com/nic0lae/golog"
	gologC "github.com/nic0lae/golog/contracts"
	gologM "github.com/nic0lae/golog/modifiers"
	gologP "github.com/nic0lae/golog/persisters"
)

var oneT = 1000
var oneM = 1000 * oneT
var oneB = 1000 * oneM

var NrOfTx = oneM

var TotalTxProcessed int64 = 0


var GlobalLogTag = "DAPoS"

func main() {
	// Setup logging
	var inmemoryLogger = gologM.NewInmemoryLogger(
		gologM.NewSimpleFormatterLogger(
			gologM.NewMultiLogger(
				[]gologC.Logger{
					// gologP.NewConsoleLogger(),
					gologP.NewFileLogger("dapos_poc.log"),
				},
			),
		),
	)

	logger.StoreSingleton(
		logger.NewLogger(
			inmemoryLogger,
		),
	)

	logger.Instance().LogInfo(GlobalLogTag, 0, "DAPoS Simulation!")
	logger.Instance().LogInfo(GlobalLogTag, 0, "~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")

		//Del Node names
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

		// Init everyone with some money
		sendRandomTransaction("dl", "BobSt", 1, 1000, getNodeByAddress(names[0]))
		sendRandomTransaction("dl", "Chris", 2, 1000, getNodeByAddress(names[1]))
		sendRandomTransaction("dl", "GregM", 3, 1000, getNodeByAddress(names[2]))
		sendRandomTransaction("dl", "Muham", 4, 1000, getNodeByAddress(names[3]))

		//make sure everyone has money before transactions start
		time.Sleep(time.Second * 5)

		//start sending transactions
		transactionID := 5
		for ; transactionID <= NrOfTx; transactionID++ {
			var node1 = getRandomNode(nil)
			var node2 = getRandomNode([]*Node{node1})
			var delegateNode = getRandomNode([]*Node{node1, node2})

			sendRandomTransaction(node1.Wallet, node2.Wallet, transactionID, 1, delegateNode)
		}
	}()

	go func() {
		logger.Instance().LogInfo(GlobalLogTag, 0, "~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")
		//print logs once all transactions are finished
		for {
			var allDone = true
			for key := range getNodes() {
				var node = getNodes()[key]
				var isDone = node.IsDoneProcessing()
				allDone = allDone && isDone
			}

			if allDone {
				break
			}

			time.Sleep(time.Second)
		}

		//get wallet and timed info after all transactions
		var totalProcessTimeInNano int64
		for key := range getNodes() {
			var node = getNodes()[key]

			var allWaletValues = []string{}
			for wallet, _ := range node.AllWallets {
				if wallet == "dl" {
					continue
				}
				allWaletValues = append(allWaletValues, fmt.Sprintf("%s - %6d", wallet, node.AllWallets[wallet]))
			}

			var delta = time.Since(node.TimeForLastTx)
			//printing wallets opinions
			logger.Instance().LogInfo(GlobalLogTag, 0,
				fmt.Sprintf(
					"AllWallets: %s | TxCount: %d | IddleFor: [%d nano] [%d mili] [%f sec] [%f min] | TotaProcessTimeInNano: %d",
					strings.Join(allWaletValues, ", "),
					node.TxCount,
					delta.Nanoseconds(),
					delta.Nanoseconds()/1000000, // mili
					delta.Seconds(),
					delta.Minutes(),
					node.TotaProcessTimeInNano,
				),
			)

			totalProcessTimeInNano += node.TotaProcessTimeInNano
		}

		logger.Instance().LogInfo(GlobalLogTag, 0, fmt.Sprintf("TotalTxProcessed: %d", TotalTxProcessed))
		logger.Instance().LogInfo(GlobalLogTag, 0, fmt.Sprintf("Performance: %d Tx in %d Nano", TotalTxProcessed, totalProcessTimeInNano))
		logger.Instance().LogInfo(GlobalLogTag, 0, "~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")

		(inmemoryLogger.(*gologM.InmemoryLogger)).Flush()

		fmt.Println("DONE!!!")
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
