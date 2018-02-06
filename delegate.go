package main

import (
	"fmt"
	"time"
)

func NewDelegate(id int, nodes int, c chan Transaction, v chan Vote) (delegate *Delegate) {
	premineWallet := Transaction{0, "dl", "Genesis", 100, time.Now(), id, nodes}
	genesisBlock := new(Block)
	genesisBlock.Transaction = premineWallet

	return &Delegate{
		Id:           id,
		PeerCount:    nodes,
		GenesisBlock: genesisBlock,
		CurrentBlock: genesisBlock,
		Channel:      c,
		VoteChannel:  v,
	}
}

func (d *Delegate)Start() {
	//listen for transactions forever
	for {
		msg := <- d.Channel
		//if transaction came from non-delegate node (new)
		if msg.DelegateId > d.PeerCount {
			d.validateBlockAndTransmit(msg, "non-delegate")
			time.Sleep(time.Second)

			//transactions from delegates should be reevaluated
		} else {
			//TODO: process unseen transactions from other delegates
			//if transaction came from another delegate, check to see if it's been seen before then process it
			if !seenTransaction(msg.Id, d.GenesisBlock) {
				d.validateBlockAndTransmit(msg, "delegate")
			} else {
				//fmt.Printf("delegate %d: skipping received transaction %d from delegate %d \n", d.Id, msg.Id, msg.DelegateId)
			}
		}
	}
}

func (d *Delegate)validateBlockAndTransmit(msg Transaction, sourceType string) {
	valid := processTransaction(d.CurrentBlock, msg)
	if valid {
		fmt.Printf("delegate %d: received valid transaction %d from a %s node %d with value: %d\n", d.Id, d.CurrentBlock.Transaction.Id, sourceType, msg.DelegateId, msg.Value)
		//save the transaction to the chain
		newBlock := Block{nil, nil, msg}
		d.CurrentBlock.Next_block = &newBlock
		d.CurrentBlock = &newBlock

		//set the delegate id to current id and broadcast the valid transaction to other nodes
		msg.DelegateId = d.Id
		for i := 0; i < d.PeerCount-1; i++ {
			d.Channel <- msg
		}
		d.VoteChannel <- Vote{TransactionId: msg.Id, VoteYesNo: true, DelegateId: d.Id}
	} else {
		fmt.Printf("delegate %d: received invalid transaction %d from an %s %d with value: %d\n", d.Id, msg.Id, sourceType, msg.DelegateId, msg.Value)
		d.VoteChannel <- Vote{TransactionId: msg.Id, VoteYesNo: false, DelegateId: d.Id}
	}

}

func seenTransaction(id int, genesisBlock *Block) bool {
	pointerBlock := genesisBlock
	for pointerBlock != nil {
		if pointerBlock.Transaction.Id == id {
			return true
		}
		pointerBlock = pointerBlock.Next_block
	}
	return false
}

//processing the transaction consists of:
//adding the transaction in it's proper place in the time-sorted linked list
//checking that transaction and the ones following it for validity
//return true if transaction is valid
func processTransaction(currentBlock *Block, msg Transaction) bool {
	//don't validate transactions on 0 or less
	if msg.Value <= 0 {
		return false
	}

	//new balance maping
	newBalances := make(map[string]int)
	newBalances[msg.From] = GetAccount(msg.From).Balance
	newBalances[msg.To] = GetAccount(msg.To).Balance
	pointerBlock := currentBlock

	//iterate until new_balances matches up with blockchain state at time of transaction
	for {
		if pointerBlock.Next_block == nil { //pointerBlock is end of the chain
			fmt.Println("there are no more blocks after pointerBlock")
			newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

			if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
				newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
			}
			break
		} else {
			fmt.Println("there are still more blocks after pointerBlock")
			//break if msg goes after pointerBlock, but before pointerBlock.next_block
			if msg.Time.After(pointerBlock.Transaction.Time) && msg.Time.Before(pointerBlock.Next_block.Transaction.Time) {

				fmt.Println("msg goes between tx %d and tx %d \n", pointerBlock.Transaction.Id, pointerBlock.Next_block.Transaction.Id)
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				break
			} else {
				//if msg.time doesn't follow pointerBlock, update new_balance and iterate
				newBalances[pointerBlock.Transaction.To] += pointerBlock.Transaction.Value

				if pointerBlock.Transaction.From != "dl" { //don't set a negative balance for premined transfers
					newBalances[pointerBlock.Transaction.From] -= pointerBlock.Transaction.Value
				}
				pointerBlock = pointerBlock.Next_block
			}
		}
	}

	fmt.Printf("current block is %d \n", pointerBlock.Transaction.Id)
	fmt.Printf("new_balances[%s]=%d \n", msg.From, newBalances[msg.From])
	//is new transaction valid?
	if newBalances[msg.From] >= msg.Value {

		//if so make a new block and add it to the chain
		new_valid_block := Block{pointerBlock, pointerBlock.Next_block, msg}
		pointerBlock.Next_block = &new_valid_block
		if new_valid_block.Next_block != nil {
			new_valid_block.Next_block.Prev_block = &new_valid_block
		}

		//update new_balances to reflect the recent addition
		newBalances[msg.From] = newBalances[msg.From] - msg.Value
		newBalances[msg.To] = newBalances[msg.To] + msg.Value

		//and then check following blocks for validity
		//using revoked_blocks to keep track of now invalid transactions
		pointerBlock = &new_valid_block
		for pointerBlock.Next_block != nil {
			//is the following transaction valid?
			if newBalances[pointerBlock.Transaction.From] >= pointerBlock.Transaction.Value {
				fmt.Printf("yay! transaction %d is still valid \n", pointerBlock.Transaction.Id)
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