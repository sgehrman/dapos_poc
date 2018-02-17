package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// func NewNode(peers int, c chan Transaction, v chan Vote) (delegate *Node) {
// 	premineWallet := Transaction{0, "dl", "Genesis", 100, time.Now(), id, nodes}
// 	genesisBlock := new(Block)
// 	genesisBlock.Transaction = premineWallet

// 	return &Node{
// 		PeerCount:    peers,
// 		GenesisBlock: genesisBlock,
// 		CurrentBlock: genesisBlock,
// 		Channel:      c,
// 		VoteChannel:  v,
// 		IsDelegate:   false,
// 	}
// }

func (node *Node) StartListenForTx() {
	fmt.Println("StartListenForTx()")

	go func() {
		for {
			tx := <-node.TxChannel
			fmt.Println(fmt.Sprintf("GotTX()         | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet.Id))

			if !getNodeByAddress(tx.From).IsDelegate {
				// if transaction came from non-delegate node (new)
				node.validateBlockAndTransmit(tx, "non-delegate")

			} else {
				//transactions from delegates should be reevaluated
				//TODO: process unseen transactions from other delegates
				//if transaction came from another delegate,
				// check to see if it's been seen before then process it

				if _, ok := node.TxFromChainById[tx.Id]; !ok {
					node.validateBlockAndTransmit(tx, "delegate")
				} else {
					//fmt.Printf("delegate %d: skipping received transaction %d from delegate %d \n", d.Id, tx.Id, tx.DelegateId)
				}
			}
		}
	}()
}

func (node *Node) validateBlockAndTransmit(tx Transaction, sourceType string) {
	fmt.Println(fmt.Sprintf("validateBlock() | Tx_%d(%s -> %s) | %s", tx.Id, tx.From, tx.To, node.Wallet.Id))

	valid := node.processTransaction(tx)

	if valid {
		log.WithFields(log.Fields{
			"Node ID":        node.Wallet.Id,
			"Transaction ID": node.CurrentBlock.Transaction.Id,
			"From":           sourceType,
			"Node":           node.Wallet.Id,
			"Value":          tx.Value,
		})

		//		fmt.Printf("delegate %d: received valid transaction %d from a %s node %d with value: %d\n", d.Id, d.CurrentBlock.Transaction.Id, sourceType, tx.DelegateId, tx.Value)
		//save the transaction to the chain
		newBlock := Block{nil, nil, tx}
		node.CurrentBlock.Next_block = &newBlock
		node.CurrentBlock = &newBlock

		node.TxFromChainById[tx.Id] = &tx

		node.VoteCounterProcessTx(&tx)

		//set the delegate id to current id and broadcast the valid transaction to other nodes

		for k, _ := range getNodes() {
			destinationNode := getNodes()[k]
			if destinationNode.IsDelegate &&
				destinationNode.Wallet.Id != node.Wallet.Id {

				// go func() {
				// 	fmt.Println(fmt.Sprintf("sendTx() | TX-ID: %d | Node-ID: %s", tx.Id, node.Wallet.Id))
				// 	destinationNode.TxChannel <- tx
				// }()

				go func() {
					fmt.Println(fmt.Sprintf("sendVote()-true | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet.Id, destinationNode.Wallet.Id))
					destinationNode.VoteChannel <- Vote{
						TransactionId: tx.Id,
						YesNo:         true,
					}
				}()
			}
		}
	} else {
		log.WithFields(log.Fields{
			"Node ID":     node.Wallet.Id,
			"Transaction": tx.Id,
			"From":        sourceType,
			"From ID":     tx.From,
			"Value":       tx.Value,
		}).Info("Received an invalid transaction")
		//		fmt.Printf("delegate %d: received invalid transaction %d from an %s %d with value: %d\n", d.Id, tx.Id, sourceType, tx.DelegateId, tx.Value)

		go func() {
			for k, _ := range getNodes() {
				destinationNode := getNodes()[k]
				if destinationNode.IsDelegate &&
					destinationNode.Wallet.Id != node.Wallet.Id {
					go func() {
						fmt.Println(fmt.Sprintf("sendVote()-false | Tx_%d(%s -> %s) | %s -> %s", tx.Id, tx.From, tx.To, node.Wallet.Id, destinationNode.Wallet.Id))
						destinationNode.VoteChannel <- Vote{
							TransactionId: tx.Id,
							YesNo:         false,
						}
					}()
				}
			}
		}()
	}
}

// func seenTransaction(id int, genesisBlock *Block) bool {
// 	pointerBlock := genesisBlock
// 	for pointerBlock != nil {
// 		// fmt.Println("pointerBlock.Transaction.Id == id", pointerBlock.Transaction.Id, id)

// 		if pointerBlock.Transaction.Id == id {
// 			return true
// 		}
// 		pointerBlock = pointerBlock.Next_block
// 	}

// 	return false
// }

//processing the transaction consists of:
//adding the transaction in it's proper place in the time-sorted linked list
//checking that transaction and the ones following it for validity
//return true if transaction is valid
func (node *Node) processTransaction(tx Transaction) bool {
	// fmt.Println(fmt.Sprintf("processTransaction() | TX-ID: %d | Node-ID: %s", tx.Id, node.Wallet.Id))

	//don't validate transactions on 0 or less
	if tx.Value <= 0 {
		return false
	}

	//new balance maping
	newBalances := make(map[WalletAddress]int)
	newBalances[tx.From] = (*getNodes()[string(tx.From)]).Wallet.Balance
	newBalances[tx.To] = (*getNodes()[string(tx.To)]).Wallet.Balance
	pointerBlock := node.CurrentBlock

	//iterate until new_balances matches up with blockchain state at time of transaction
	for {
		if pointerBlock.Next_block == nil { //pointerBlock is end of the chain
			log.Info("there are no more blocks after pointerBlock")
			newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

			if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
			}
			break
		} else {
			log.Info("there are still more blocks after pointerBlock")
			//break if tx goes after pointerBlock, but before pointerBlock.next_block
			if tx.Time.After(pointerBlock.Transaction.Time) && tx.Time.Before(pointerBlock.Next_block.Transaction.Time) {

				log.WithFields(log.Fields{"pointerBlock.Transaction.Id": pointerBlock.Transaction.Id, "pointerBlock.Next_block.Transaction.Id": pointerBlock.Next_block.Transaction.Id}).Info("Message betwen blocks")
				//				fmt.Println("tx goes between tx %d and tx %d \n", pointerBlock.Transaction.Id, pointerBlock.Next_block.Transaction.Id)
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				break
			} else {
				//if tx.time doesn't follow pointerBlock, update new_balance and iterate
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				pointerBlock = pointerBlock.Next_block
			}
		}
	}

	log.WithFields(log.Fields{"pointerBlock.Transaction.Id": pointerBlock.Transaction.Id})
	log.WithFields(log.Fields{"new_balances[]": tx.From, "Balance": newBalances[tx.From]})

	//	fmt.Printf("current block is %d \n", pointerBlock.Transaction.Id)
	//	fmt.Printf("new_balances[%s]=%d \n", tx.From, newBalances[tx.From])

	//is new transaction valid?
	if newBalances[tx.From] >= tx.Value {

		//if so make a new block and add it to the chain
		new_valid_block := Block{pointerBlock, pointerBlock.Next_block, tx}
		pointerBlock.Next_block = &new_valid_block
		if new_valid_block.Next_block != nil {
			new_valid_block.Next_block.Prev_block = &new_valid_block
		}

		//update new_balances to reflect the recent addition
		newBalances[tx.From] = newBalances[tx.From] - tx.Value
		newBalances[tx.To] = newBalances[tx.To] + tx.Value

		//and then check following blocks for validity
		//using revoked_blocks to keep track of now invalid transactions
		pointerBlock = &new_valid_block
		for pointerBlock.Next_block != nil {
			//is the following transaction valid?
			if newBalances[pointerBlock.Transaction.From] >= pointerBlock.Transaction.Value {
				// fmt.Printf("yay! transaction %d is still valid \n", pointerBlock.Transaction.Id)
				//yay! transaction is valid - update balances and continue onto next block
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Next_block == nil {
					break
				} else {
					pointerBlock = pointerBlock.Next_block
				}
			}
			if newBalances[pointerBlock.Transaction.From] < pointerBlock.Transaction.Value {
				fmt.Printf("turns out transaction %d is invalid \n", pointerBlock.Transaction.Id)
				//oh no! this previously believed valid block is actually invaid! D:
				//remove block from list and keep going
				pointerBlock.Prev_block.Next_block = pointerBlock.Next_block
				pointerBlock.Next_block.Prev_block = pointerBlock.Prev_block
			}
		}
		//when finished, broadcast awesome new block and potentially broken transactions
		return true

	} else { //if new transaction is not valid, drop that bitch like it's hot
		return false
	}
}

func (node *Node) SendRandomTransaction(transactionId int, delegates []WalletAddress) {

	fromWallet := node.Wallet
	toWallet := getRandomWallet()

	amount := 1

	transaction := Transaction{
		transactionId,
		fromWallet.Id,
		toWallet.Id,
		amount,
		time.Now(),
		delegates,
	}

	for k, _ := range getNodes() {
		if getNodes()[k].IsDelegate {
			getNodes()[k].TxChannel <- transaction
		}
	}
}
