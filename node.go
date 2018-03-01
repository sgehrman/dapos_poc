package main

import (
	"time"

	logger "github.com/nic0lae/golog"
)


func (node *Node) StartListenForTx() {
	logger.Instance().LogInfo(node.Wallet, 0, "StartListenForTx()")

	go func() {
		for {
			//get transaction info from transaction channel
			tx := <-node.TxChannel

			//start current transaction time
			var perfClock = time.Now()

			//uncomment to see individual transactions info
			//logger.Instance().LogInfo(node.Wallet, 1,
			//	fmt.Sprintf("GotTX() | Tx_%d(%s -> %d -> %s) | FromDelegate: %s", tx.Id, tx.From, tx.Value, tx.To, tx.DelId),
			//)

			//check if node has already validated this transaction
			seen := node.checkIfValidated(tx.Id)

			if seen {
				// if the tx has already been validated, log and do nothing
				//logger.Instance().LogInfo(node.Wallet, 2, "skipping, Tx already processed")
			} else {
				// else check tx for validity
				node.validateBlockAndTransmit(tx)
			}

			//add transaction to amount of transactions processed and add the time it took
			node.TxCount++
			node.TimeForLastTx = time.Now()
			node.TotaProcessTimeInNano += time.Since(perfClock).Nanoseconds()

			TotalTxProcessed++
		}
	}()
}

func (node *Node) IsDoneProcessing() bool {
	//if no transactions within 5 seconds consider done with transactions
	var sec = time.Since(node.TimeForLastTx).Seconds()

	return (sec > 5)
}

func (node *Node) checkIfValidated(txId int) bool {
	var ok bool
	_, ok = node.TxFromChainById[txId]
	return ok
}

func (node *Node) validateBlockAndTransmit(tx Transaction) {
	//logger.Instance().LogInfo(node.Wallet, 2, "validateBlockAndTransmit()")

	//call Validate(transaction)
	valid := node.validate(tx)

	if valid {
		//add to seen list
		node.TxFromChainById[tx.Id] = tx

		// set the delegate id to current id and broadcast the valid transaction to other nodes
		for k, _ := range getNodes() {

			destinationNode := getNodes()[k]
			//don't send to self or Del that sent us the transaction
			if destinationNode.Wallet == node.Wallet || destinationNode.Wallet == tx.DelId {
				continue
			}

			//logger.Instance().LogInfo(node.Wallet, 3,
			//	fmt.Sprintf("sendTx() | Tx_%d(%s -> %d -> %s) | SentBy: %s | SendingTo: %s", tx.Id, tx.From, tx.Value, tx.To, tx.DelId, destinationNode.Wallet),
			//)

			//send transaction
			go func(node *Node) {
				tx.DelId = node.Wallet
				destinationNode.TxChannel <- tx
			}(destinationNode)
		}
	}
}

//validates the transaction and adds it to the end of the chain
func (node *Node) validate(tx Transaction) bool {

	//logger.Instance().LogInfo(node.Wallet, 3, "validate()")

	var isTxValid = false

	//can't send nothing or negatives
	if tx.Value > 0 {
		//logger.Instance().LogInfo(node.Wallet, 4, fmt.Sprintf("finding block placement"))

		//store blocks to transverse through
		currentBlock := node.LastBlock

		//store balances to reach balance at time of transaction
		heldBalance := make(map[string]int)
		for k, v := range node.AllWallets {
			heldBalance[k] = v
		}

		for {
			//traverse backwards through transactions to find correct place of tx

			//if new tx goes after currentblock, then inser into list and check following blocks for validity
			if tx.Time.After(currentBlock.Transaction.Time) || tx.Time.Equal(currentBlock.Transaction.Time) {

				if heldBalance[tx.From] < tx.Value { //sender doesn't have enough money
					// fmt.Println("Invalid Transaction")
					isTxValid = false
					break
				}
				//Updating temp wallet
				heldBalance[tx.From] -= tx.Value
				heldBalance[tx.To] += tx.Value

				//Make Block
				newBlock := &Block{
					currentBlock,
					currentBlock.Next,
					tx,
				}

				//set pointers to add newest block
				if currentBlock.Next != nil {
					currentBlock.Next.Prev = newBlock
					currentBlock.Next = newBlock
					currentBlock = newBlock
				} else {
					currentBlock.Next = newBlock
					currentBlock = newBlock
					node.LastBlock = newBlock
				}

				//Traverse forwards
				for currentBlock.Next != nil {
					// fmt.Println("Iterating forwards")
					if heldBalance[currentBlock.Next.Transaction.From] < currentBlock.Next.Transaction.Value { //sender doesn't have enough money
						//replace block and move on
						if currentBlock.Next.Next != nil {
							// fmt.Println("Replacing next block ")
							currentBlock.Next.Next.Prev = currentBlock
							currentBlock.Next = currentBlock.Next.Next
						} else {
							// fmt.Println("Dropping next block ")
							currentBlock.Next = nil
							isTxValid = true
							break
						}
						currentBlock = currentBlock.Next

					} else {
						//still Valid, add to temp wallet
						heldBalance[currentBlock.Next.Transaction.From] -= currentBlock.Next.Transaction.Value
						heldBalance[currentBlock.Next.Transaction.To] += currentBlock.Next.Transaction.Value
						currentBlock = currentBlock.Next
					}
				}
				// set new balance to actual balance
				for k, v := range heldBalance {
					node.AllWallets[k] = v
				}

				isTxValid = true
				break

			} else { //if new tx does not go after currentBlock, move currentblock pointer back and check again
				// fmt.Println("Iterating backwards")
				//chance temp balance to be in the past
				heldBalance[currentBlock.Transaction.From] += currentBlock.Transaction.Value
				heldBalance[currentBlock.Transaction.To] -= currentBlock.Transaction.Value

				if currentBlock.Prev == nil {
					// fmt.Println("No Prev Block found! breaking!")
					isTxValid = false
					break
				}

				currentBlock = currentBlock.Prev
			}
		}
	}

	//if isTxValid {
	//	logger.Instance().LogInfo(node.Wallet, 4, "Tx is VALID")
	//} else {
	//	logger.Instance().LogInfo(node.Wallet, 4, "Tx is IN-VALID")
	//}

	return isTxValid
}
