package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (node *Node) StartListenForTx() {
	go func() {
		for {
			tx := <-node.TxChannel

			var logLines = []string{}
			var additionalLogLines = []string{}

			if !getNodeByAddress(tx.TxInfoSender).IsDelegate {
				// if transaction came from non-delegate node (new)

				logLines = append(logLines, fmt.Sprintf("GotTX()-node    | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet.Id))

				additionalLogLines = node.validateBlockAndTransmit(&tx, "non-delegate")
			} else {
				//transactions from delegates should be reevaluated
				//TODO: process unseen transactions from other delegates
				//if transaction came from another delegate,
				// check to see if it's been seen before then process it

				logLines = append(logLines, fmt.Sprintf("GotTX()-dlgate  | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet.Id))

				if _, ok := node.TxFromChainById[tx.Id]; !ok {
					additionalLogLines = node.validateBlockAndTransmit(&tx, "delegate")
				} else {
					additionalLogLines = append(additionalLogLines, fmt.Sprintf("delegate %s: skipping received transaction %d from delegate X", node.Wallet.Id, tx.Id))
				}
			}

			additionalLogLines = prefixLinesWith(additionalLogLines, "    ")
			logLines = append(logLines, additionalLogLines...)

			log.Info(strings.Join(logLines, "\n"))
		}
	}()
}

func (node *Node) validateBlockAndTransmit(tx *Transaction, sourceType string) []string {
	var logLines = []string{}
	var additionalLogLines = []string{}

	logLines = append(logLines, fmt.Sprintf("validateBlock()"))

	valid, additionalLogLines := node.processTransaction(tx)

	additionalLogLines = prefixLinesWith(additionalLogLines, "    ")
	logLines = append(logLines, additionalLogLines...)

	if valid {
		logLines = append(logLines, fmt.Sprintf("Node ID: %s, Transaction ID: %d, From: %s, Node: %s, Value: %d", node.Wallet.Id, node.LastBlock.Transaction.Id, sourceType, node.Wallet.Id, tx.Value))
		logLines = append(logLines, fmt.Sprintf("delegate %s: received valid transaction %d from a %s node X with value: %d", node.Wallet.Id, node.LastBlock.Transaction.Id, sourceType, tx.Value))

		// save the transaction to the chain
		newBlock := Block{nil, nil, *tx}
		node.LastBlock.Next = &newBlock
		node.LastBlock = &newBlock

		node.TxFromChainById[tx.Id] = tx

		node.VoteCounterProcessTx(tx)

		// set the delegate id to current id and broadcast the valid transaction to other nodes

		for k, _ := range getNodes() {
			destinationNode := getNodes()[k]
			if destinationNode.IsDelegate &&
				destinationNode.Wallet.Id != node.Wallet.Id &&
				!contains(tx.CurrentValidators, destinationNode.Wallet.Id) {

				logLines = append(logLines, fmt.Sprintf("sendTx()        | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet.Id, destinationNode.Wallet.Id))
				go func() {
					tx.TxInfoSender = node.Wallet.Id
					destinationNode.TxChannel <- *tx
				}()

				logLines = append(logLines, fmt.Sprintf("sendVote()-true | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet.Id, destinationNode.Wallet.Id))
				go func() {
					destinationNode.VoteChannel <- Vote{
						TransactionId: tx.Id,
						YesNo:         true,
					}
				}()
			}
		}
	} else {
		logLines = append(logLines, fmt.Sprintf("Node ID: %s, Transaction: %d, From: %s, From ID: %s, Value: %d", node.Wallet.Id, tx.Id, sourceType, tx.From, tx.Value))

		// Info("Received an invalid transaction")
		logLines = append(logLines, fmt.Sprintf("delegate %s: received invalid transaction %d from an %s X with value: %d\n", node.Wallet.Id, tx.Id, sourceType, tx.Value))

		// go func() {
		for k, _ := range getNodes() {
			destinationNode := getNodes()[k]
			if destinationNode.IsDelegate &&
				destinationNode.Wallet.Id != node.Wallet.Id {

				logLines = append(logLines, fmt.Sprintf("sendVote()-false | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet.Id, destinationNode.Wallet.Id))
				go func() {
					destinationNode.VoteChannel <- Vote{
						TransactionId: tx.Id,
						YesNo:         false,
					}
				}()
			}
		}
		// }()
	}

	return logLines
}

//processing the transaction consists of:
//adding the transaction in it's proper place in the time-sorted linked list
//checking that transaction and the ones following it for validity
//return true if transaction is valid
func (node *Node) processTransaction(tx *Transaction) (bool, []string) {
	var logLines = []string{}

	logLines = append(logLines, fmt.Sprintf("processTransaction()"))

	var validators = []string{}
	for _, v := range tx.CurrentValidators {
		validators = append(validators, string(v))
	}
	logLines = append(logLines, fmt.Sprintf("Tx_%d(%s -> [%d] -> %s) SeenBy{%s}", tx.Id, tx.From, tx.Value, tx.To, strings.Join(validators, ",")))

	// Add THIS node as a validator, to be sent later
	tx.CurrentValidators = append(tx.CurrentValidators, node.Wallet.Id)

	var fromNode = getNodes()[string(tx.From)]
	var toNode = getNodes()[string(tx.To)]

	var fromBalance = fromNode.Wallet.Balance
	var toBalance = toNode.Wallet.Balance

	// Don't validate transactions on 0 or less
	// Don't validate transactions with not enough $$$
	if tx.Value <= 0 || fromBalance < tx.Value {
		logLines = append(logLines, fmt.Sprintf("Invalid TX"))
		return false, logLines
	}

	pointerBlock := node.LastBlock

	var blockIndex = 0

	// iterate until new_balances matches up with blockchain state at time of transaction
	for {
		if pointerBlock.Next == nil { // pointerBlock is end of the chain
			newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

			if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
			}
			break
		} else {
			blockIndex++

			// break if tx goes after pointerBlock, but before pointerBlock.Next
			if tx.Time.After(pointerBlock.Transaction.Time) && tx.Time.Before(pointerBlock.Next.Transaction.Time) {

				logLines = append(logLines, fmt.Sprintf("pointerBlock.Transaction.Id: %d, pointerBlock.Next.Transaction.Id: %d", pointerBlock.Transaction.Id, pointerBlock.Next.Transaction.Id))
				logLines = append(logLines, fmt.Sprintf("tx goes between tx %d and tx %d \n", pointerBlock.Transaction.Id, pointerBlock.Next.Transaction.Id))
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				break
			} else {
				// if tx.time doesn't follow pointerBlock, update new_balance and iterate
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				pointerBlock = pointerBlock.Next
			}
		}
	}

	logLines = append(logLines, fmt.Sprintf("we are at the block index %d", blockIndex))

	logLines = append(logLines, fmt.Sprintf("pointerBlock.Transaction.Id: %d", pointerBlock.Transaction.Id))
	logLines = append(logLines, fmt.Sprintf("new_balances[]: %s, Balance: %d", tx.From, newBalances[tx.From]))

	logLines = append(logLines, fmt.Sprintf("current block is %d", pointerBlock.Transaction.Id))
	logLines = append(logLines, fmt.Sprintf("new_balances[%s]=%d", tx.From, newBalances[tx.From]))

	// if so make a new block and add it to the chain
	newValidBlock := &Block{pointerBlock, pointerBlock.Next, *tx}

	pointerBlock.Next = newValidBlock
	if newValidBlock.Next != nil {
		newValidBlock.Next.Prev = newValidBlock
	}

	// Update balances to reflect the recent addition
	newBalances[tx.From] = newBalances[tx.From] - tx.Value
	newBalances[tx.To] = newBalances[tx.To] + tx.Value

	// And then check following blocks for validity
	// using revoked_blocks to keep track of now invalid transactions
	pointerBlock = newValidBlock
	for pointerBlock.Next != nil {
		// is the following transaction valid?
		if newBalances[pointerBlock.Transaction.From] >= pointerBlock.Transaction.Value {
			logLines = append(logLines, fmt.Sprintf("yay! transaction %d is still valid", pointerBlock.Transaction.Id))
			//yay! transaction is valid - update balances and continue onto next block
			newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
			newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

			if pointerBlock.Next == nil {
				break
			} else {
				pointerBlock = pointerBlock.Next
			}
		}
		if newBalances[pointerBlock.Transaction.From] < pointerBlock.Transaction.Value {
			logLines = append(logLines, fmt.Sprintf("turns out transaction %d is invalid \n", pointerBlock.Transaction.Id))
			//oh no! this previously believed valid block is actually invaid! D:
			//remove block from list and keep going
			pointerBlock.Prev.Next = pointerBlock.Next
			pointerBlock.Next.Prev = pointerBlock.Prev
		}
	}

	// when finished, broadcast awesome new block and potentially broken transactions
	return true, logLines
}

func contains(arr []WalletAddress, str WalletAddress) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func (node *Node) processTransaction_backup(tx *Transaction) (bool, []string) {
	var logLines = []string{}

	logLines = append(logLines, fmt.Sprintf("processTransaction()"))

	var validators = []string{}
	for _, v := range tx.CurrentValidators {
		validators = append(validators, string(v))
	}
	logLines = append(logLines, fmt.Sprintf("Tx_%d(%s -> [%d] -> %s) SeenBy{%s}", tx.Id, tx.From, tx.Value, tx.To, strings.Join(validators, ",")))

	// Add THIS node as a validator, to be sent later
	tx.CurrentValidators = append(tx.CurrentValidators, node.Wallet.Id)

	// Don't validate transactions on 0 or less
	if tx.Value <= 0 {
		logLines = append(logLines, fmt.Sprintf("Invalid TX"))
		return false, logLines
	}

	// new balance maping
	newBalances := make(map[WalletAddress]int)
	newBalances[tx.From] = (*getNodes()[string(tx.From)]).Wallet.Balance
	newBalances[tx.To] = (*getNodes()[string(tx.To)]).Wallet.Balance

	pointerBlock := node.LastBlock

	var blockIndex = 0

	// iterate until new_balances matches up with blockchain state at time of transaction
	for {
		if pointerBlock.Next == nil { // pointerBlock is end of the chain
			newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

			if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
			}
			break
		} else {
			blockIndex++

			// break if tx goes after pointerBlock, but before pointerBlock.Next
			if tx.Time.After(pointerBlock.Transaction.Time) && tx.Time.Before(pointerBlock.Next.Transaction.Time) {

				logLines = append(logLines, fmt.Sprintf("pointerBlock.Transaction.Id: %d, pointerBlock.Next.Transaction.Id: %d", pointerBlock.Transaction.Id, pointerBlock.Next.Transaction.Id))
				logLines = append(logLines, fmt.Sprintf("tx goes between tx %d and tx %d \n", pointerBlock.Transaction.Id, pointerBlock.Next.Transaction.Id))
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				break
			} else {
				// if tx.time doesn't follow pointerBlock, update new_balance and iterate
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				pointerBlock = pointerBlock.Next
			}
		}
	}

	logLines = append(logLines, fmt.Sprintf("we are at the block index %d", blockIndex))

	logLines = append(logLines, fmt.Sprintf("pointerBlock.Transaction.Id: %d", pointerBlock.Transaction.Id))
	logLines = append(logLines, fmt.Sprintf("new_balances[]: %s, Balance: %d", tx.From, newBalances[tx.From]))

	logLines = append(logLines, fmt.Sprintf("current block is %d", pointerBlock.Transaction.Id))
	logLines = append(logLines, fmt.Sprintf("new_balances[%s]=%d", tx.From, newBalances[tx.From]))

	// is new transaction valid?
	if newBalances[tx.From] >= tx.Value {

		// if so make a new block and add it to the chain
		newValidBlock := &Block{pointerBlock, pointerBlock.Next, *tx}

		pointerBlock.Next = newValidBlock
		if newValidBlock.Next != nil {
			newValidBlock.Next.Prev = newValidBlock
		}

		// Update balances to reflect the recent addition
		newBalances[tx.From] = newBalances[tx.From] - tx.Value
		newBalances[tx.To] = newBalances[tx.To] + tx.Value

		// And then check following blocks for validity
		// using revoked_blocks to keep track of now invalid transactions
		pointerBlock = newValidBlock
		for pointerBlock.Next != nil {
			// is the following transaction valid?
			if newBalances[pointerBlock.Transaction.From] >= pointerBlock.Transaction.Value {
				logLines = append(logLines, fmt.Sprintf("yay! transaction %d is still valid", pointerBlock.Transaction.Id))
				//yay! transaction is valid - update balances and continue onto next block
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Next == nil {
					break
				} else {
					pointerBlock = pointerBlock.Next
				}
			}
			if newBalances[pointerBlock.Transaction.From] < pointerBlock.Transaction.Value {
				logLines = append(logLines, fmt.Sprintf("turns out transaction %d is invalid \n", pointerBlock.Transaction.Id))
				//oh no! this previously believed valid block is actually invaid! D:
				//remove block from list and keep going
				pointerBlock.Prev.Next = pointerBlock.Next
				pointerBlock.Next.Prev = pointerBlock.Prev
			}
		}

		// when finished, broadcast awesome new block and potentially broken transactions
		return true, logLines
	}

	// if new transaction is not valid, drop that bitch like it's hot
	logLines = append(logLines, fmt.Sprintf("Invalid TX"))
	return false, logLines
}
