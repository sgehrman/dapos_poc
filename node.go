package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"time"
)

func (node *Node) StartListenForTx() {
	fmt.Println("StartListenForTx")

	go func() {
		for {
			tx := <-node.TxChannel

			var logLines = []string{}
			var additionalLogLines = []string{}

			logLines = append(logLines, fmt.Sprintf("GotTX()-node    | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet))


			seen := node.checkIfValidated(tx.Id)
			if seen { //if the tx has already been validated, log and do nothing
				additionalLogLines = append(additionalLogLines, fmt.Sprintf("delegate %s: skipping received transaction %d", node.Wallet, tx.Id))
			} else { //else check tx for validity
				additionalLogLines = node.validateBlockAndTransmit(&tx)
			}

			additionalLogLines = prefixLinesWith(additionalLogLines, "    ")
			logLines = append(logLines, additionalLogLines...)

			log.Info(strings.Join(logLines, "\n"))
		}
	}()
}

func (node *Node) checkIfValidated(txId int) bool {
	if node.TxFromChainById[txId] == nil {
		return false
	} else {
		return true
	}
}

func (node *Node) validateBlockAndTransmit(tx *Transaction) []string {
	var logLines = []string{}
	var additionalLogLines = []string{}

	logLines = append(logLines, fmt.Sprintf("validateBlock()"))

	//call Validate(transaction)
	valid := node.validate(tx)

	additionalLogLines = prefixLinesWith(additionalLogLines, "    ")
	logLines = append(logLines, additionalLogLines...)

	if valid {
		logLines = append(logLines, fmt.Sprintf("Node ID: %s, Transaction ID: %d, Value: %d", node.Wallet, tx.Id, tx.Value))
		logLines = append(logLines, fmt.Sprintf("delegate %s: received valid transaction %d with value: %d", node.Wallet, tx.Id, tx.Value))

		//add valid transaction to 'validated' list
		node.TxFromChainById[tx.Id] = tx

		//report back if no more expected tx
		//if tx was last expected (4) then report balances
		node.TxCount++
		if node.TxCount == 1 {
			node.StartTime = time.Now()
		}
		if node.TxCount == 999 {
			fmt.Printf("Node %s thinks balance of BobSt: %d, Chris: %d, GregM: %d, Muham: %d \n",
				node.Wallet,
				node.AllWallets["BobSt"],
				node.AllWallets["Chris"],
				node.AllWallets["GregM"],
				node.AllWallets["Muham"])

			TimeToComplete := time.Since(node.StartTime)
			fmt.Printf("Delegate %s processed %d transactions in %d time", node.Wallet, 4, TimeToComplete)

		}

		// set the delegate id to current id and broadcast the valid transaction to other nodes
		for k, _ := range getNodes() {
			destinationNode := getNodes()[k]

			logLines = append(logLines, fmt.Sprintf("sendTx()        | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet, destinationNode.Wallet))
			go func() {
				destinationNode.TxChannel <- *tx
			}()
		}
	} else {
		logLines = append(logLines, fmt.Sprintf("Node ID: %s, Transaction: %d, From ID: %s, Value: %d", node.Wallet, tx.Id,  tx.From, tx.Value))
		logLines = append(logLines, fmt.Sprintf("delegate %s: received invalid transaction %d with value: %d", node.Wallet, tx.Id,  tx.Value))

	}

	return logLines
}

//validates the transaction and adds it to the end of the chain
func (node *Node) validate(tx *Transaction) bool {
	//don't process a negative tx
	if tx.Value < 0 {
		return false
	}

	//check if transaction goes at end of list, then AllWallets can check validity
	//if tx.Time.After(node.LastBlock.Transaction.Time) {
	if true {
		if (node.AllWallets[tx.From] < tx.Value) { //sender doesn't have enough money
			return false
		} else { //transaction is valid!!!
			//update AllWallets balance
			node.AllWallets[tx.From] -= tx.Value
			node.AllWallets[tx.To] += tx.Value

			//add tx to end of list
			node.LastBlock.Next = &Block{
				node.LastBlock,
				nil,
				tx,
			}
			node.LastBlock = node.LastBlock.Next

			//return true then add to TxFromChainById & broadcast to delegates
			return true
		}
	} else { //if tx is not at end of list, iterate backwards to find balances of time of tx
		//TODO: support tx that come before lastBlock

		//start with node.AllWallets and iterate backwards from node.LastBlock until
		//the state of the chain at time of transaction is discovered

	}

	return false
}