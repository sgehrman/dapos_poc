package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	logger "github.com/nic0lae/golog"
	gologC "github.com/nic0lae/golog/contracts"
	gologM "github.com/nic0lae/golog/modifiers"
	gologP "github.com/nic0lae/golog/persisters"
)

var oneT = 1000
var oneM = 1000 * oneT
var oneB = 1000 * oneM

// NrOfTx number of transactions
var NrOfTx = oneM

// TotalTxProcessed number of transactions processed
var TotalTxProcessed int64

// GlobalLogTag Global log tag
var GlobalLogTag = "DAPoS"

// GlobalStressTest Global stress test flag
var GlobalStressTest = false

func main() {
	nodeCount := flag.Int("nodeCount", 4, "Total nodes in the simulation")
	txCount := flag.Int("txCount", NrOfTx, "Total transactions in the simulation")
	flag.BoolVar(&GlobalStressTest, "stress", GlobalStressTest, "Adds random delays and random tx amounts")
	flag.Parse()
	finished := make(chan bool)
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

	var names []string
	for i := 0; i < *nodeCount; i++ {
		names = append(names, "node-"+strconv.Itoa(i+1))
	}

	// Create Delegates
	for _, name := range names {
		CreateNodeAndAddToList(name)
	}

	// Run Transactions
	go func() {
		startingBalance := 1000

		// Init everyone with some money
		for i := 0; i < *nodeCount; i++ {
			// randomize the amount
			if GlobalStressTest {
				startingBalance = 10 + GetRandomNumber(1000)
			}

			sendRandomTransaction("dl", names[i], i+1, startingBalance, getNodeByAddress(names[i]))
		}

		//make sure everyone has money before transactions start
		time.Sleep(time.Second * 5)

		amount := 1

		//start sending transactions
		transactionID := *nodeCount + 1
		for ; transactionID <= *txCount; transactionID++ {
			var node1 = getRandomNode(nil)
			var node2 = getRandomNode([]*Node{node1})
			var delegateNode = getRandomNode([]*Node{node1, node2})

			// randomize the amount
			if GlobalStressTest {
				amount = 1 + GetRandomNumber(8)
			}

			sendRandomTransaction(node1.Wallet, node2.Wallet, transactionID, amount, delegateNode)
		}
	}()

	go func(finished chan bool) {
		logger.Instance().LogInfo(GlobalLogTag, 0, "~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")
		//print logs once all transactions are finished
		for {
			var allDone = true
			for _, node := range getNodes() {
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
		for _, node := range getNodes() {
			var allWaletValues = []string{}
			for wallet := range node.AllWallets {
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
		logger.Instance().LogInfo(GlobalLogTag, 0, fmt.Sprintf(
			"Performance: %d Tx in %d Nanos [%d millis]",
			TotalTxProcessed, totalProcessTimeInNano, totalProcessTimeInNano/1000000))
		logger.Instance().LogInfo(GlobalLogTag, 0, "~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~ ~~~~")

		(inmemoryLogger.(*gologM.InmemoryLogger)).Flush()

		fmt.Println("DONE!!!")
		finished <- true
	}(finished)

	<-finished
}
