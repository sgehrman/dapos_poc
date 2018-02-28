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

var NrOfTx = oneT * 30

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

		// Init eveyrone with some money
		sendRandomTransaction("dl", "BobSt", 1, 1000, getNodeByAddress(names[0]))
		sendRandomTransaction("dl", "Chris", 2, 1000, getNodeByAddress(names[1]))
		sendRandomTransaction("dl", "GregM", 3, 1000, getNodeByAddress(names[2]))
		sendRandomTransaction("dl", "Muham", 4, 1000, getNodeByAddress(names[3]))

		time.Sleep(time.Second * 5)

		/* Used to isolate and debug buggy TX
		var genesisWallet = "dl"
		var chrisWallet = names[1]

		var bobNode = getNodeByAddress(names[0])

		var nowTime = time.Now()
		var time1 = time.Unix(nowTime.Unix()+10, 0)
		var time2 = time.Unix(nowTime.Unix()+5, 0)

		sendRandomTransactionWithTime(genesisWallet, chrisWallet, transactionID, 1, bobNode, time1)
		transactionID++
		sendRandomTransactionWithTime(genesisWallet, chrisWallet, transactionID, 1, bobNode, time2)
		*/

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

		// time.Sleep(time.Second * 10)

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

		// var totalPRocessTimeInSec = totalProcessTimeInNano / 1000000000

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
